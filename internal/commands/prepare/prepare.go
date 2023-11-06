package prepare

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"veertu.com/anka-cloud-gitlab-executor/internal/ankacloud"
	"veertu.com/anka-cloud-gitlab-executor/internal/env"
	"veertu.com/anka-cloud-gitlab-executor/internal/log"
)

var Command = &cobra.Command{
	Use:  "prepare",
	RunE: execute,
}

func execute(cmd *cobra.Command, args []string) error {
	log.SetOutput(os.Stderr)
	log.Println("Running prepare stage")

	controllerURL, ok := os.LookupEnv(env.VarControllerURL)
	if !ok {
		return fmt.Errorf("%w: %s", env.ErrMissingVar, env.VarControllerURL)
	}

	templateId, ok := os.LookupEnv(env.VarTemplateId)
	if !ok {
		return fmt.Errorf("%w: %s", env.ErrMissingVar, env.VarTemplateId)
	}

	jobId, ok := os.LookupEnv(env.VarGitlabJobId)
	if !ok {
		return fmt.Errorf("%w: %s", env.ErrMissingVar, env.VarGitlabJobId)
	}

	config := ankacloud.CreateInstanceConfig{
		TemplateId: templateId,
		ExternalId: jobId,
	}

	tag, ok := os.LookupEnv(env.VarTemplateTag)
	if ok {
		config.TemplateTag = tag
	}

	nodeId, ok := os.LookupEnv(env.VarNodeId)
	if ok {
		config.NodeId = nodeId
	}

	priorityStr, ok := os.LookupEnv(env.VarPriority)
	if ok {
		priority, err := strconv.Atoi(priorityStr)
		if err != nil {
			return fmt.Errorf("failed converting priority to int: %w", err)
		}

		config.Priority = priority
	}

	nodeGroupId, ok := os.LookupEnv(env.VarNodeGroupId)
	if ok {
		config.NodeGroupId = nodeGroupId
	}

	clientConfig := ankacloud.ClientConfig{
		ControllerURL: controllerURL,
	}

	caCertPath, ok := os.LookupEnv(env.VarCaCertPath)
	if ok {
		clientConfig.CACertPath = caCertPath
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

	log.Printf("creating instance with config: %+v\n", config)
	instanceId, err := controller.CreateInstance(config)
	if err != nil {
		return fmt.Errorf("failed creating instance: %w", err)
	}

	if err := controller.WaitForInstanceToBeScheduled(instanceId); err != nil {
		return fmt.Errorf("failed waiting for instance to be scheduled: %w", err)
	}

	log.Printf("created instance id: %s\n", instanceId)

	return nil
}
