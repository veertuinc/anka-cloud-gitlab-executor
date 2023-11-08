package command

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"veertu.com/anka-cloud-gitlab-executor/internal/ankacloud"
	"veertu.com/anka-cloud-gitlab-executor/internal/gitlab"
	"veertu.com/anka-cloud-gitlab-executor/internal/log"
)

var prepareCommand = &cobra.Command{
	Use:  "prepare",
	RunE: executePrepare,
}

func executePrepare(cmd *cobra.Command, args []string) error {
	log.SetOutput(os.Stderr)
	log.Println("Running prepare stage")

	env, ok := cmd.Context().Value(contextKey("env")).(gitlab.Environment)
	if !ok {
		return fmt.Errorf("failed to get environment from context")
	}

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

	httpClientConfig, err := getHttpClientConfig(env)
	if err != nil {
		return fmt.Errorf("failed to initialize HTTP client config: %w", err)
	}

	httpClient, err := ankacloud.NewHTTPClient(httpClientConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize HTTP client with config +%v: %w", httpClientConfig, err)
	}

	controller := ankacloud.Client{
		ControllerURL: env.ControllerURL,
		HttpClient:    httpClient,
	}

	log.Printf("creating instance with config: %+v\n", req)
	instanceId, err := controller.CreateInstance(cmd.Context(), req)
	if err != nil {
		return fmt.Errorf("failed to create instance: %w", err)
	}

	if err := controller.WaitForInstanceToBeScheduled(cmd.Context(), instanceId); err != nil {
		return fmt.Errorf("failed to wait for instance %q to be scheduled: %w", instanceId, err)
	}

	log.Printf("created instance id: %s\n", instanceId)

	return nil
}
