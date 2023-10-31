package prepare

import (
	"fmt"

	"github.com/spf13/cobra"
	"veertu.com/anka-cloud-gitlab-executor/internal/gitlab"
	"veertu.com/anka-cloud-gitlab-executor/pkg/ankaCloud"
)

var Command = &cobra.Command{
	Use:  "prepare",
	RunE: execute,
}

func execute(cmd *cobra.Command, args []string) error {
	ankaCloudConfig := ankaCloud.ClientConfig{
		ControllerURL: "http://192.168.122.183",
	}
	controller := ankaCloud.NewClient(ankaCloudConfig)

	jobId, err := gitlab.GetGitlabEnvVar("CI_JOB_ID")
	if err != nil {
		return fmt.Errorf("failed getting CI_JOB_ID: %w", err)
	}

	_, err = controller.CreateInstance(ankaCloud.CreateInstanceConfig{
		TemplateId:         "8c592f53-65a4-444e-9342-79d3ff07837c",
		ExternalId:         jobId,
		WaitUntilScheduled: true,
	})
	if err != nil {
		return fmt.Errorf("failed creating instance: %w", err)
	}

	return nil
}
