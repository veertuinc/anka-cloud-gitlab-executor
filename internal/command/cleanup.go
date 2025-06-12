package command

import (
	"context"
	"fmt"

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
	// log.SetOutput(os.Stdout) // prevents us from logging Println, etc

	log.Println("cleanup stage started for job: ", env.GitlabJobUrl)

	if env.KeepAliveOnError && env.GitlabJobStatus == gitlab.JobStatusFailed {
		log.Colorln("keeping VM alive on error")
		return nil
	}

	apiClientConfig := getAPIClientConfig(env)
	apiClient, err := ankacloud.NewAPIClient(apiClientConfig)
	if err != nil {
		log.Errorf("cleanup: failed to initialize API client with config +%v: %v", apiClientConfig, err)
		return fmt.Errorf("cleanup: failed to initialize API client with config +%v: %v", apiClientConfig, err)
	}

	controller := ankacloud.NewController(apiClient)

	instance, err := controller.GetInstanceByExternalId(ctx, env.GitlabJobUrl)
	if err != nil {
		log.Errorf("cleanup: failed to get instance by external id %q: %v", env.GitlabJobUrl, err)
		return fmt.Errorf("cleanup: failed to get instance by external id %q: %v", env.GitlabJobUrl, err)
	}
	log.Debugf("instance id: %s\n", instance.Id)

	log.Debugf("Issuing termination request for instance %s\n", instance.Id)
	err = controller.TerminateInstance(ctx, ankacloud.TerminateInstanceRequest{
		Id: instance.Id,
	})
	if err != nil {
		log.Errorf("cleanup: failed to terminate instance %q: %v", instance.Id, err)
		return fmt.Errorf("cleanup: failed to terminate instance %q: %v", instance.Id, err)
	}

	log.Println("cleanup stage completed for job: ", env.GitlabJobUrl)
	return nil
}
