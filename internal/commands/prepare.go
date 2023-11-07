package commands

import (
	"fmt"
	"os"
	"strings"

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

	controllerURL, ok := os.LookupEnv(gitlab.VarControllerURL)
	if !ok {
		return fmt.Errorf("%w: %s", gitlab.ErrMissingVar, gitlab.VarControllerURL)
	}
	if !strings.HasPrefix(controllerURL, "http") {
		return fmt.Errorf("controller url %q missing http prefix", controllerURL)
	}

	templateId, ok := os.LookupEnv(gitlab.VarTemplateId)
	if !ok {
		return fmt.Errorf("%w: %s", gitlab.ErrMissingVar, gitlab.VarTemplateId)
	}

	jobId, ok := os.LookupEnv(gitlab.VarGitlabJobId)
	if !ok {
		return fmt.Errorf("%w: %s", gitlab.ErrMissingVar, gitlab.VarGitlabJobId)
	}

	req := ankacloud.CreateInstanceRequest{
		TemplateId: templateId,
		ExternalId: jobId,
	}

	if tag, ok := os.LookupEnv(gitlab.VarTemplateTag); ok {
		req.Tag = tag
	}

	if nodeId, ok := os.LookupEnv(gitlab.VarNodeId); ok {
		req.NodeId = nodeId
	}

	if priority, ok, err := gitlab.GetIntVar(gitlab.VarPriority); ok {
		if err != nil {
			return fmt.Errorf("failed parsing priority: %w", err)
		}

		req.Priority = priority
	}

	if nodeGroupId, ok := os.LookupEnv(gitlab.VarNodeGroupId); ok {
		req.NodeGroupId = nodeGroupId
	}

	httpClientConfig, err := httpClientConfigFromEnvVars(controllerURL)
	if err != nil {
		return fmt.Errorf("failing initializing HTTP client config: %w", err)
	}

	httpClient, err := ankacloud.NewHTTPClient(httpClientConfig)
	if err != nil {
		return fmt.Errorf("failing initializing HTTP client with config +%v: %w", httpClientConfig, err)
	}

	controller := ankacloud.Client{
		ControllerURL: controllerURL,
		HttpClient:    httpClient,
	}

	log.Printf("creating instance with config: %+v\n", req)
	instanceId, err := controller.CreateInstance(cmd.Context(), req)
	if err != nil {
		return fmt.Errorf("failed creating instance: %w", err)
	}

	if err := controller.WaitForInstanceToBeScheduled(cmd.Context(), instanceId); err != nil {
		return fmt.Errorf("failed waiting for instance %q to be scheduled: %w", instanceId, err)
	}

	log.Printf("created instance id: %s\n", instanceId)

	return nil
}
