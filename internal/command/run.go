package command

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/veertuinc/anka-cloud-gitlab-executor/internal/ankacloud"
	"github.com/veertuinc/anka-cloud-gitlab-executor/internal/gitlab"
	"github.com/veertuinc/anka-cloud-gitlab-executor/internal/log"
	"golang.org/x/crypto/ssh"
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

	log.Debugf("running run stage %s\n", args[1])

	apiClientConfig := getAPIClientConfig(env)
	apiClient, err := ankacloud.NewAPIClient(apiClientConfig)
	if err != nil {
		return gitlab.TransientError(fmt.Errorf("failed to initialize API client with config +%v: %w", apiClientConfig, err))
	}

	controller := ankacloud.NewController(apiClient)

	instance, err := controller.GetInstanceByExternalId(ctx, env.GitlabJobUrl)
	if err != nil {
		return gitlab.TransientError(fmt.Errorf("failed to get instance by external id %q: %w", env.GitlabJobUrl, err))
	}

	var nodeSshPort string
	if instance.VMInfo == nil {
		return gitlab.TransientError(fmt.Errorf("instance has no VM: %+v", instance))
	}

	for _, rule := range instance.VMInfo.PortForwardingRules {
		if rule.VmPort == 22 && rule.Protocol == "tcp" {
			nodeSshPort = fmt.Sprintf("%d", rule.NodePort)
		}
	}
	if nodeSshPort == "" {
		return gitlab.TransientError(fmt.Errorf("could not find ssh port forwarded for vm"))
	}
	log.Debugf("node SSH port to VM: %s\n", nodeSshPort)

	gitlabScriptFile, err := os.Open(args[0])
	if err != nil {
		return gitlab.TransientError(fmt.Errorf("failed to open script file at %q: %w", args[0], err))
	}
	defer gitlabScriptFile.Close()
	log.Debugf("gitlab script path: %s", args[0])

	sshUserName := env.SSHUserName
	if sshUserName == "" {
		sshUserName = defaultSshUserName
	}

	sshPassword := env.SSHPassword
	if sshPassword == "" {
		sshPassword = defaultSshPassword
	}

	sshClientConfig := &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		User:            sshUserName,
		Auth: []ssh.AuthMethod{
			ssh.Password(sshPassword),
		},
	}

	node, err := controller.GetNode(ctx, ankacloud.GetNodeRequest{Id: instance.NodeId})
	if err != nil {
		return gitlab.TransientError(fmt.Errorf("failed to get node %s: %w", instance.NodeId, err))
	}

	addr := fmt.Sprintf("%s:%s", node.IP, nodeSshPort)
	var sshClient *ssh.Client

	// retry logic mimics what is done by the official Gitlab Runner (true for gitlab runner v16.7.0)
	maxAttempts := env.SSHAttempts
	if maxAttempts < 1 {
		maxAttempts = 4
	}
	sshConnectionAttemptDelay := env.SSHConnectionAttemptDelay
	if sshConnectionAttemptDelay < 1 {
		sshConnectionAttemptDelay = 5
	}
	for i := 0; i < maxAttempts; i++ {
		log.Debugf("attempt #%d to establish ssh connection to %q\n", i+1, addr)
		sshClient, err = ssh.Dial("tcp", addr, sshClientConfig)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(sshConnectionAttemptDelay) * time.Second)
	}
	if err != nil {
		return gitlab.TransientError(fmt.Errorf("failed to create new ssh client connection to %q: %w", addr, err))
	}
	defer sshClient.Close()

	log.Debugln("ssh connection established")

	session, err := sshClient.NewSession()
	if err != nil {
		return gitlab.TransientError(fmt.Errorf("failed to start new ssh session: %w", err))
	}
	defer session.Close()
	log.Debugln("ssh session opened")

	session.Stdin = gitlabScriptFile
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	err = session.Shell()
	if err != nil {
		return gitlab.TransientError(fmt.Errorf("failed to start Shell on SSH session: %w", err))
	}

	log.Debugln("waiting for remote execution to finish")
	err = session.Wait()

	log.Debugln("remote execution finished")
	return err
}
