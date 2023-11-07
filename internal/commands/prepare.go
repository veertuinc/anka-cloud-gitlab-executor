package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"veertu.com/anka-cloud-gitlab-executor/internal/ankacloud"
	"veertu.com/anka-cloud-gitlab-executor/internal/env"
	"veertu.com/anka-cloud-gitlab-executor/internal/log"
)

var prepareCommand = &cobra.Command{
	Use:  "prepare",
	RunE: executePrepare,
}

func executePrepare(cmd *cobra.Command, args []string) error {
	log.SetOutput(os.Stderr)
	log.Println("Running prepare stage")

	controllerURL, ok := os.LookupEnv(env.VarControllerURL)
	if !ok {
		return fmt.Errorf("%w: %s", env.ErrMissingVar, env.VarControllerURL)
	}
	if !strings.HasPrefix(controllerURL, "http") {
		return fmt.Errorf("controller url %q missing http prefix", controllerURL)
	}

	templateId, ok := os.LookupEnv(env.VarTemplateId)
	if !ok {
		return fmt.Errorf("%w: %s", env.ErrMissingVar, env.VarTemplateId)
	}

	jobId, ok := os.LookupEnv(env.VarGitlabJobId)
	if !ok {
		return fmt.Errorf("%w: %s", env.ErrMissingVar, env.VarGitlabJobId)
	}

	req := ankacloud.CreateInstanceRequest{
		TemplateId: templateId,
		ExternalId: jobId,
	}

	tag, ok := os.LookupEnv(env.VarTemplateTag)
	if ok {
		req.Tag = tag
	}

	nodeId, ok := os.LookupEnv(env.VarNodeId)
	if ok {
		req.NodeId = nodeId
	}

	priorityStr, ok := os.LookupEnv(env.VarPriority)
	if ok {
		priority, err := strconv.Atoi(priorityStr)
		if err != nil {
			return fmt.Errorf("failed converting priority %q to int: %w", priorityStr, err)
		}

		req.Priority = priority
	}

	nodeGroupId, ok := os.LookupEnv(env.VarNodeGroupId)
	if ok {
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
