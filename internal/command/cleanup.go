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

	log.Println("Running cleanup stage")

	httpClientConfig, err := getHttpClientConfig(env)
	if err != nil {
		return fmt.Errorf("failed to initialize HTTP client config from environment variables: %w", err)
	}

	httpClient, err := ankacloud.NewHTTPClient(httpClientConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize HTTP client with config +%v: %w", httpClientConfig, err)
	}

	controller := ankacloud.Client{
		ControllerURL: env.ControllerURL,
		HttpClient:    httpClient,
	}

	instance, err := controller.GetInstanceByExternalId(ctx, env.GitlabJobId)
	if err != nil {
		return fmt.Errorf("failed to get instance by external id %q: %w", env.GitlabJobId, err)
	}
	log.Printf("instance id: %s\n", instance.Id)

	err = controller.TerminateInstance(ctx, ankacloud.TerminateInstanceRequest{
		Id: instance.Id,
	})
	if err != nil {
		return fmt.Errorf("failed to terminate instance %q: %w", instance.Id, err)
	}
	log.Printf("Issuing termination request for instance %s\n", instance.Id)

	return nil
}
