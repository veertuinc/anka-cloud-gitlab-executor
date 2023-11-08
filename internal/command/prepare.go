package command

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"veertu.com/anka-cloud-gitlab-executor/internal/ankacloud"
	"veertu.com/anka-cloud-gitlab-executor/internal/gitlab"
	"veertu.com/anka-cloud-gitlab-executor/internal/log"
)

var prepareCommand = &cobra.Command{
	Use: "prepare",
	RunE: func(cmd *cobra.Command, args []string) error {
		env, ok := cmd.Context().Value(contextKey("env")).(gitlab.Environment)
		if !ok {
			return fmt.Errorf("failed to get environment from context")
		}

		return executePrepare(cmd.Context(), env)
	},
}

func executePrepare(ctx context.Context, env gitlab.Environment) error {
	log.SetOutput(os.Stderr)
	log.Println("Running prepare stage")

	if env.TemplateId == "" {
		return fmt.Errorf("failed to get template id from environment: %w", gitlab.ErrMissingVar)
	}

	req := ankacloud.CreateInstanceRequest{
		TemplateId:  env.TemplateId,
		ExternalId:  env.GitlabJobId,
		Tag:         env.TemplateTag,
		NodeId:      env.NodeId,
		Priority:    env.Priority,
		NodeGroupId: env.NodeGroupId,
	}

	apiClientConfig := getAPIClientConfig(env)
	apiClient, err := ankacloud.NewAPIClient(apiClientConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize API client with config +%v: %w", apiClientConfig, err)
	}

	controller := ankacloud.NewController(apiClient)

	log.Printf("creating instance with config: %+v\n", req)
	instanceId, err := controller.CreateInstance(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create instance: %w", err)
	}

	if err := controller.WaitForInstanceToBeScheduled(ctx, instanceId); err != nil {
		return fmt.Errorf("failed to wait for instance %q to be scheduled: %w", instanceId, err)
	}

	log.Printf("created instance id: %s\n", instanceId)

	return nil
}
