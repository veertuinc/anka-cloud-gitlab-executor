package cleanup

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"veertu.com/anka-cloud-gitlab-executor/internal/gitlab"
	"veertu.com/anka-cloud-gitlab-executor/internal/log"
	"veertu.com/anka-cloud-gitlab-executor/pkg/ankaCloud"
)

var Command = &cobra.Command{
	Use:  "cleanup",
	RunE: execute,
}

func execute(cmd *cobra.Command, args []string) error {
	log.SetOutput(os.Stdout)

	controllerUrl, err := gitlab.GetAnkaCloudEnvVar("CONTROLLER_URL")
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
	instance, err := controller.GetInstanceByExternalId(jobId)
	if err != nil {
		return fmt.Errorf("failed getting instance by external id: %w", err)
	}

	err = controller.TerminateInstance(ankaCloud.TerminateInstanceConfig{
		InstanceId: instance.Id,
	})
	if err != nil {
		return fmt.Errorf("failed terminating instance: %w", err)
	}

	return nil
}
