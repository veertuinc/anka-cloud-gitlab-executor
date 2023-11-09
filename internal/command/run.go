package command

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"veertu.com/anka-cloud-gitlab-executor/internal/ankacloud"
	"veertu.com/anka-cloud-gitlab-executor/internal/gitlab"
	"veertu.com/anka-cloud-gitlab-executor/internal/log"
)

const (
	defaultSshUserName = "anka"
	defaultSshPassword = "admin"
)

var runCommand = &cobra.Command{
	Use: "run",
	RunE: func(cmd *cobra.Command, args []string) error {
		env, ok := cmd.Context().Value(contextKey("env")).(gitlab.Environment)
		if !ok {
			return fmt.Errorf("failed to get environment from context")
		}

		return executeRun(cmd.Context(), env, args)
	},
}

func executeRun(ctx context.Context, env gitlab.Environment, args []string) error {
	log.SetOutput(os.Stderr)

	log.Printf("Running run stage %s\n", args[1])

	apiClientConfig := getAPIClientConfig(env)
	apiClient, err := ankacloud.NewAPIClient(apiClientConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize API client with config +%v: %w", apiClientConfig, err)
	}

	controller := ankacloud.NewController(apiClient)

	instance, err := controller.GetInstanceByExternalId(ctx, env.GitlabJobId)
	if err != nil {
		return fmt.Errorf("failed to get instance by external id %q: %w", env.GitlabJobId, err)
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
	node, err := controller.GetNode(ctx, ankacloud.GetNodeRequest{Id: nodeId})
	if err != nil {
		return fmt.Errorf("failed to get node %s: %w", nodeId, err)
	}
	nodeIp = node.IP
	log.Printf("node IP: %s\n", nodeIp)

	gitlabScriptFile, err := os.Open(args[0])
	if err != nil {
		return fmt.Errorf("failed to open script file at %q: %w", args[0], err)
	}
	defer gitlabScriptFile.Close()
	log.Printf("gitlab script path: %s", args[0])

	addr := fmt.Sprintf("%s:%s", nodeIp, nodeSshPort)
	dialer := net.Dialer{}
	netConn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to create tcp dialer: %w", err)
	}
	log.Printf("connected to %s\n", addr)

	sshUserName := env.SSHUserName
	if sshUserName == "" {
		sshUserName = defaultSshUserName
	}

	sshPassword := env.SSHPassword
	if sshPassword == "" {
		sshPassword = defaultSshPassword
	}

	sshConfig := &ssh.ClientConfig{
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		User: sshUserName,
		Auth: []ssh.AuthMethod{
			ssh.Password(sshPassword),
		},
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(netConn, addr, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to create new ssh client connection to %q with config %+v: %w", addr, sshConfig, err)
	}
	defer sshConn.Close()

	log.Println("ssh connection established")
	sshClient := ssh.NewClient(sshConn, chans, reqs)
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		return fmt.Errorf("failed to start new ssh session: %w", err)
	}
	defer session.Close()
	log.Println("ssh session opened")

	session.Stdin = gitlabScriptFile
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	err = session.Shell()
	if err != nil {
		return fmt.Errorf("failed to start Shell on SSH session: %w", err)
	}

	log.Println("waiting for remote execution to finish")
	err = session.Wait()

	log.Println("remote execution finished")
	return err
}
