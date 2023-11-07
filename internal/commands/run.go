package commands

import (
	"fmt"
	"net"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"veertu.com/anka-cloud-gitlab-executor/internal/ankacloud"
	"veertu.com/anka-cloud-gitlab-executor/internal/env"
	"veertu.com/anka-cloud-gitlab-executor/internal/log"
)

var runCommand = &cobra.Command{
	Use:  "run",
	RunE: executeRun,
}

const macUser = "anka"
const macPw = "admin"

func executeRun(cmd *cobra.Command, args []string) error {
	log.SetOutput(os.Stderr)

	log.Printf("Running run stage %s\n", args[1])

	controllerURL, ok := os.LookupEnv(env.VarControllerURL)
	if !ok {
		return fmt.Errorf("%w: %s", env.ErrMissingVar, env.VarControllerURL)
	}

	httpClientConfig, err := httpClientConfigFromEnvVars(controllerURL)
	if err != nil {
		return fmt.Errorf("failing initializing HTTP client config: %w", err)
	}

	httpClient, err := ankacloud.NewHTTPClient(*httpClientConfig)
	if err != nil {
		return fmt.Errorf("failing initializing HTTP client with config +%v: %w", httpClientConfig, err)
	}

	controller := ankacloud.Client{
		ControllerURL: controllerURL,
		HttpClient:    httpClient,
	}

	jobId, ok := os.LookupEnv(env.VarGitlabJobId)
	if !ok {
		return fmt.Errorf("%w: %s", env.ErrMissingVar, env.VarGitlabJobId)
	}

	instance, err := controller.GetInstanceByExternalId(jobId)
	if err != nil {
		return fmt.Errorf("failed getting instance by external id %q: %w", jobId, err)
	}

	log.Printf("instance id: %s\n", instance.Id)

	var nodeIp, nodeSshPort string
	if instance.VM == nil {
		return fmt.Errorf("instance has no VM: %+v", instance)
	}

	for _, rule := range instance.VM.PortForwardingRules {
		if rule.VmPort == 22 && rule.Protocol == "tcp" {
			nodeSshPort = fmt.Sprintf("%d", rule.NodePort)
		}
	}
	if nodeSshPort == "" {
		return fmt.Errorf("could not find ssh port forwarded for vm")
	}
	log.Printf("node SSH port to VM: %s\n", nodeSshPort)

	nodeId := instance.NodeId
	node, err := controller.GetNode(ankacloud.GetNodeRequest{Id: nodeId})
	if err != nil {
		return fmt.Errorf("failed getting node %s: %w", nodeId, err)
	}
	nodeIp = node.IP
	log.Printf("node IP: %s\n", nodeIp)

	gitlabScriptFile, err := os.Open(args[0])
	if err != nil {
		return fmt.Errorf("failed opening script file at %q: %w", args[0], err)
	}
	defer gitlabScriptFile.Close()
	log.Printf("gitlab script path: %s", args[0])

	addr := fmt.Sprintf("%s:%s", nodeIp, nodeSshPort)
	dialer := net.Dialer{}
	netConn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed creating tcp dialer: %w", err)
	}
	log.Printf("connected to %s\n", addr)
	sshConfig := &ssh.ClientConfig{
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		User: macUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(macPw),
		},
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(netConn, addr, sshConfig)
	if err != nil {
		return fmt.Errorf("failed creating new ssh client connection to %q with config %+v: %w", addr, sshConfig, err)
	}
	defer sshConn.Close()

	log.Println("ssh connection established")
	sshClient := ssh.NewClient(sshConn, chans, reqs)
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		return fmt.Errorf("failed starting new ssh session: %w", err)
	}
	defer session.Close()
	log.Println("ssh session opened")

	session.Stdin = gitlabScriptFile
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	err = session.Shell()
	if err != nil {
		return fmt.Errorf("failed starting Shell on SSH session: %w", err)
	}

	log.Println("waiting for remote execution to finish")
	err = session.Wait()

	log.Println("remote execution finished")
	return err
}
