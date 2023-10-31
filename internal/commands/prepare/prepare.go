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
	controllerUrl, err := gitlab.GetAnkaCloudEnvVar("CONTROLLER_URL")
	if err != nil {
		return err
	}

	templateId, err := gitlab.GetAnkaCloudEnvVar("TEMPLATE_ID")
	if err != nil {
		return err
	}

	controller := ankaCloud.NewClient(ankaCloud.ClientConfig{
		ControllerURL: controllerUrl,
	})

	jobId, err := gitlab.GetGitlabEnvVar("CI_JOB_ID")
	if err != nil {
		return err
	}

	_, err = controller.CreateInstance(ankaCloud.CreateInstanceConfig{
		TemplateId:         templateId,
		ExternalId:         jobId,
		WaitUntilScheduled: true,
	})
	if err != nil {
		return fmt.Errorf("failed creating instance: %w", err)
	}

	return nil
}
