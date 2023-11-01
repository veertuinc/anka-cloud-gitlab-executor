package cleanup

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"veertu.com/anka-cloud-gitlab-executor/internal/ankaCloud"
	"veertu.com/anka-cloud-gitlab-executor/internal/env"
	"veertu.com/anka-cloud-gitlab-executor/internal/errors"
	"veertu.com/anka-cloud-gitlab-executor/internal/log"
)

var Command = &cobra.Command{
	Use:  "cleanup",
	RunE: execute,
}

func execute(cmd *cobra.Command, args []string) error {
	log.SetOutput(os.Stdout)

	log.Println("Running cleanup stage")

	controllerUrl, ok := env.Get(env.AnkaVar("CONTROLLER_URL"))
	if !ok {
		return errors.MissingEnvVar(env.AnkaVar("CONTROLLER_URL"))
	}

	controller := ankaCloud.NewClient(ankaCloud.ClientConfig{
		ControllerURL: controllerUrl,
	})

	jobId, ok := env.Get(env.GitlabVar("CI_JOB_ID"))
	if !ok {
		return errors.MissingEnvVar(env.GitlabVar("CI_JOB_ID"))
	}
	instance, err := controller.GetInstanceByExternalId(jobId)
	if err != nil {
		return fmt.Errorf("failed getting instance by external id: %w", err)
	}
	log.Printf("instance id: %s\n", instance.Id)

	err = controller.TerminateInstance(ankaCloud.TerminateInstanceConfig{
		InstanceId: instance.Id,
	})
	if err != nil {
		return fmt.Errorf("failed terminating instance: %w", err)
	}
	log.Printf("Issuing termination request for instance %s\n", instance.Id)

	return nil
}
