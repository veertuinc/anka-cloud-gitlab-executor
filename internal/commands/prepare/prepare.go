package prepare

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"veertu.com/anka-cloud-gitlab-executor/internal/ankaCloud"
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

	controllerURL, ok := os.LookupEnv(env.VAR_CONTROLLER_URL)
	if !ok {
		return fmt.Errorf("%w: %s", env.ErrMissingVar, env.VAR_CONTROLLER_URL)
	}

	templateId, ok := os.LookupEnv(env.VAR_TEMPLATE_ID)
	if !ok {
		return fmt.Errorf("%w: %s", env.ErrMissingVar, env.VAR_TEMPLATE_ID)
	}

	jobId, ok := os.LookupEnv(env.VAR_GITLAB_JOB_ID)
	if !ok {
		return fmt.Errorf("%w: %s", env.ErrMissingVar, env.VAR_GITLAB_JOB_ID)
	}

	config := ankaCloud.CreateInstanceConfig{
		TemplateId: templateId,
		ExternalId: jobId,
	}

	tag, ok := os.LookupEnv(env.VAR_TEMPLATE_TAG)
	if ok {
		config.TemplateTag = tag
	}

	nodeId, ok := os.LookupEnv(env.VAR_NODE_ID)
	if ok {
		config.NodeId = nodeId
	}

	priorityStr, ok := os.LookupEnv(env.VAR_PRIORITY)
	if ok {
		priority, err := strconv.Atoi(priorityStr)
		if err != nil {
			return fmt.Errorf("failed converting priority to int: %w", err)
		}

		config.Priority = priority
	}

	controller := ankaCloud.NewClient(ankaCloud.ClientConfig{
		ControllerURL: controllerURL,
	})

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
