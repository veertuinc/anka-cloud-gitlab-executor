package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"veertu.com/anka-cloud-gitlab-executor/internal/ankacloud"
	"veertu.com/anka-cloud-gitlab-executor/internal/env"
	"veertu.com/anka-cloud-gitlab-executor/internal/log"
)

var cleanupCommand = &cobra.Command{
	Use:  "cleanup",
	RunE: executeCleanup,
}

func executeCleanup(cmd *cobra.Command, args []string) error {
	log.SetOutput(os.Stdout)

	log.Println("Running cleanup stage")

	controllerURL, ok := os.LookupEnv(env.VarControllerURL)
	if !ok {
		return fmt.Errorf("%w: %s", env.ErrMissingVar, env.VarControllerURL)
	}

	httpClientConfig, err := httpClientConfigFromEnvVars(controllerURL)
	if err != nil {
		return fmt.Errorf("failing initializing HTTP client config from environment variables: %w", err)
	}

	httpClient, err := ankacloud.NewHTTPClient(httpClientConfig)
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

	instance, err := controller.GetInstanceByExternalId(cmd.Context(), jobId)
	if err != nil {
		return fmt.Errorf("failed getting instance by external id %q: %w", jobId, err)
	}
	log.Printf("instance id: %s\n", instance.Id)

	err = controller.TerminateInstance(cmd.Context(), ankacloud.TerminateInstanceRequest{
		Id: instance.Id,
	})
	if err != nil {
		return fmt.Errorf("failed terminating instance %q: %w", instance.Id, err)
	}
	log.Printf("Issuing termination request for instance %s\n", instance.Id)

	return nil
}
