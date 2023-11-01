package prepare

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
	Use:  "prepare",
	RunE: execute,
}

func execute(cmd *cobra.Command, args []string) error {
	log.SetOutput(os.Stderr)
	log.Println("Running prepare stage")

	controllerUrl, ok := env.Get(env.AnkaVar("CONTROLLER_URL"))
	if !ok {
		return errors.MissingEnvVar(env.AnkaVar("CONTROLLER_URL"))
	}

	templateId, ok := env.Get(env.AnkaVar("TEMPLATE_ID"))
	if !ok {
		return errors.MissingEnvVar(env.AnkaVar("TEMPLATE_ID"))
	}

	controller := ankaCloud.NewClient(ankaCloud.ClientConfig{
		ControllerURL: controllerUrl,
	})

	jobId, ok := env.Get(env.GitlabVar("CI_JOB_ID"))
	if !ok {
		return errors.MissingEnvVar(env.GitlabVar("CI_JOB_ID"))
	}

	config := ankaCloud.CreateInstanceConfig{
		TemplateId: templateId,
		ExternalId: jobId,
	}

	log.Printf("creating instance with config: %+v\n", config)
	instanceId, err := controller.CreateInstance(config)
	if err != nil {
		return fmt.Errorf("failed creating instance: %w", err)
	}

	if err := controller.WaitForInstanceToBeScheduled(instanceId); err != nil {
		return fmt.Errorf("failed waiting for instance to be scheduled: %w", err)
	}

	log.Printf("created instance id: %s\n", instanceId)

	return nil
}
