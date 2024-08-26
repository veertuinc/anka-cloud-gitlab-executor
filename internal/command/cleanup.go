package command

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/veertuinc/anka-cloud-gitlab-executor/internal/ankacloud"
	"github.com/veertuinc/anka-cloud-gitlab-executor/internal/gitlab"
	"github.com/veertuinc/anka-cloud-gitlab-executor/internal/log"
)

var cleanupCommand = &cobra.Command{
	Use: "cleanup",
	RunE: func(cmd *cobra.Command, args []string) error {
		env, ok := cmd.Context().Value(contextKey("env")).(gitlab.Environment)
		if !ok {
			return fmt.Errorf("failed to get environment from context")
		}

		return executeCleanup(cmd.Context(), env)
	},
}

func executeCleanup(ctx context.Context, env gitlab.Environment) error {
	log.SetOutput(os.Stdout)

	log.Debugln("running cleanup stage")

	if env.KeepAliveOnError && env.GitlabJobStatus == gitlab.JobStatusFailed {
		log.Colorln("keeping VM alive on error")
		return nil
	}

	apiClientConfig := getAPIClientConfig(env)
	apiClient, err := ankacloud.NewAPIClient(apiClientConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize API client with config +%v: %w", apiClientConfig, err)
	}

	controller := ankacloud.NewController(apiClient)

	instance, err := controller.GetInstanceByExternalId(ctx, env.GitlabJobUrl)
	if err != nil {
		return fmt.Errorf("failed to get instance by external id %q: %w", env.GitlabJobUrl, err)
	}
	log.Debugf("instance id: %s\n", instance.Id)

	err = controller.TerminateInstance(ctx, ankacloud.TerminateInstanceRequest{
		Id: instance.Id,
	})
	if err != nil {
		return fmt.Errorf("failed to terminate instance %q: %w", instance.Id, err)
	}
	log.Debugf("Issuing termination request for instance %s\n", instance.Id)

	return nil
}
