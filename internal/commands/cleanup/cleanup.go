package cleanup

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"veertu.com/anka-cloud-gitlab-executor/pkg/ankaCloud"
)

var Command = &cobra.Command{
	Use:  "cleanup",
	RunE: execute,
}

func execute(cmd *cobra.Command, args []string) error {
	ankaCloudConfig := ankaCloud.ClientConfig{
		ControllerURL: "http://192.168.122.183",
	}
	controller := ankaCloud.NewClient(ankaCloudConfig)

	instance, err := controller.GetInstanceByExternalId(os.Getenv("CUSTOM_ENV_CI_JOB_ID"))
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
