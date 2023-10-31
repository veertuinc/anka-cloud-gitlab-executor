package run

import (
	"fmt"
	"net"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"veertu.com/anka-cloud-gitlab-executor/internal/gitlab"
	"veertu.com/anka-cloud-gitlab-executor/pkg/ankaCloud"
)

var Command = &cobra.Command{
	Use:  "run",
	RunE: execute,
}

const macUser = "anka"
const macPw = "admin"

func execute(cmd *cobra.Command, args []string) error {
	controllerUrl, err := gitlab.GetAnkaCloudEnvVar("CONTROLLER_URL")
	if err != nil {
		return err
	}
	controller := ankaCloud.NewClient(ankaCloud.ClientConfig{
		ControllerURL: controllerUrl,
	})

	jobId, err := gitlab.GetGitlabEnvVar("CI_JOB_ID")
	if err != nil {
		return err
	}

	instance, err := controller.GetInstanceByExternalId(jobId)
	if err != nil {
		return fmt.Errorf("failed getting instance by external id: %w", err)
	}

	var nodeIp, nodeSshPort string
	for _, rule := range instance.Instance.VM.PortForwardingRules {
		if rule.VmPort == 22 && rule.Protocol == "tcp" {
			nodeSshPort = fmt.Sprintf("%d", rule.NodePort)
		}
	}
	if nodeSshPort == "" {
		return fmt.Errorf("could not find ssh port forwarded for vm")
	}

	nodeId := instance.Instance.NodeId
	node, err := controller.GetNode(ankaCloud.GetNodeConfig{
		Id: nodeId,
	})
	if err != nil {
		return fmt.Errorf("failed getting node %s information: %w", nodeId, err)
	}
	nodeIp = node.IP

	gitlabScriptFile, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer gitlabScriptFile.Close()

	addr := fmt.Sprintf("%s:%s", nodeIp, nodeSshPort)

	dialer := net.Dialer{}
	netConn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return err
	}
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

	sshClient := ssh.NewClient(sshConn, chans, reqs)
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdin = gitlabScriptFile
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	err = session.Start("bash")
	if err != nil {
		return err
	}

	return session.Wait()
}
