package run

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"veertu.com/anka-cloud-gitlab-executor/internal/ankacloud"
	"veertu.com/anka-cloud-gitlab-executor/internal/env"
	"veertu.com/anka-cloud-gitlab-executor/internal/log"
)

var Command = &cobra.Command{
	Use:  "run",
	RunE: execute,
}

const macUser = "anka"
const macPw = "admin"

func execute(cmd *cobra.Command, args []string) error {
	log.SetOutput(os.Stderr)

	log.Printf("Running run stage %s\n", args[1])

	controllerURL, ok := os.LookupEnv(env.VarControllerURL)
	if !ok {
		return fmt.Errorf("%w: %s", env.ErrMissingVar, env.VarControllerURL)
	}

	clientConfig := ankacloud.ClientConfig{
		ControllerURL: controllerURL,
	}

	caCertPath, ok := os.LookupEnv(env.VarCaCertPath)
	if ok {
		clientConfig.CACertPath = caCertPath
	}

	clientCertPath, ok := os.LookupEnv(env.VarClientCertPath)
	if ok {
		clientConfig.ClientCertPath = clientCertPath
	}

	clientCertKeyPath, ok := os.LookupEnv(env.VarClientCertKeyPath)
	if ok {
		clientConfig.ClientCertKeyPath = clientCertKeyPath
	}

	skipTLSVerify, ok := os.LookupEnv(env.VarSkipTLSVerify)
	if ok {
		skip, err := strconv.ParseBool(skipTLSVerify)
		if err != nil {
			return fmt.Errorf("could not convert variable %q(%s) to boolean: %w", env.VarSkipTLSVerify, skipTLSVerify, err)
		}
		clientConfig.SkipTLSVerify = skip
	}

	controller, err := ankacloud.NewClient(clientConfig)
	if err != nil {
		return fmt.Errorf("failed creating anka cloud client: %w", err)
	}

	jobId, ok := os.LookupEnv(env.VarGitlabJobId)
	if !ok {
		return fmt.Errorf("%w: %s", env.ErrMissingVar, env.VarGitlabJobId)
	}

	instance, err := controller.GetInstanceByExternalId(jobId)
	if err != nil {
		return fmt.Errorf("failed getting instance by external id: %w", err)
	}

	log.Printf("instance id: %s\n", instance.Id)

	var nodeIp, nodeSshPort string
	for _, rule := range instance.Instance.VM.PortForwardingRules {
		if rule.VmPort == 22 && rule.Protocol == "tcp" {
			nodeSshPort = fmt.Sprintf("%d", rule.NodePort)
		}
	}
	if nodeSshPort == "" {
		return fmt.Errorf("could not find ssh port forwarded for vm")
	}
	log.Printf("node SSH port to VM: %s\n", nodeSshPort)

	nodeId := instance.Instance.NodeId
	node, err := controller.GetNode(ankacloud.GetNodeConfig{
		Id: nodeId,
	})
	if err != nil {
		return fmt.Errorf("failed getting node %s information: %w", nodeId, err)
	}
	nodeIp = node.IP
	log.Printf("node IP: %s\n", nodeIp)

	gitlabScriptFile, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer gitlabScriptFile.Close()
	log.Printf("gitlab script path: %s", args[0])

	addr := fmt.Sprintf("%s:%s", nodeIp, nodeSshPort)
	dialer := net.Dialer{}
	netConn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return err
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
		return err
	}
	defer sshConn.Close()

	log.Println("ssh connection established")
	sshClient := ssh.NewClient(sshConn, chans, reqs)
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	log.Println("ssh session opened")

	session.Stdin = gitlabScriptFile
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	err = session.Shell()
	if err != nil {
		return err
	}

	log.Println("waiting for remote execution to finish")
	err = session.Wait()

	log.Println("remote execution finished")
	return err
}
