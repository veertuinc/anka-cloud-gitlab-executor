package cleanup

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"veertu.com/anka-cloud-gitlab-executor/internal/ankacloud"
	"veertu.com/anka-cloud-gitlab-executor/internal/env"
	"veertu.com/anka-cloud-gitlab-executor/internal/log"
)

var Command = &cobra.Command{
	Use:  "cleanup",
	RunE: execute,
}

func execute(cmd *cobra.Command, args []string) error {
	log.SetOutput(os.Stdout)

	log.Println("Running cleanup stage")

	controllerURL, ok := os.LookupEnv(env.VarControllerURL)
	if !ok {
		return fmt.Errorf("%w: %s", env.ErrMissingVar, env.VarControllerURL)
	}

	clientConfig := ankacloud.ClientConfig{
		ControllerURL: controllerURL,
	}

	caCertPath, ok := os.LookupEnv(env.VarCaCertPath)
	if ok {
		clientConfig.CACertPath = caCertPath
	}

	skipTLSVerify, ok := os.LookupEnv(env.VarSkipTLSVerify)
	if ok {
		skip, err := strconv.ParseBool(skipTLSVerify)
		if err != nil {
			return fmt.Errorf("could not convert variable %q(%s) to boolean: %w", env.VarSkipTLSVerify, skipTLSVerify, err)
		}
		clientConfig.SkipTLSVerify = skip
	}

	controller, err := ankacloud.NewClient(clientConfig)
	if err != nil {
		return fmt.Errorf("failed creating anka cloud client: %w", err)
	}

	jobId, ok := os.LookupEnv(env.VarGitlabJobId)
	if !ok {
		return fmt.Errorf("%w: %s", env.ErrMissingVar, env.VarGitlabJobId)
	}

	instance, err := controller.GetInstanceByExternalId(jobId)
	if err != nil {
		return fmt.Errorf("failed getting instance by external id: %w", err)
	}
	log.Printf("instance id: %s\n", instance.Id)

	err = controller.TerminateInstance(ankacloud.TerminateInstanceConfig{
		InstanceId: instance.Id,
	})
	if err != nil {
		return fmt.Errorf("failed terminating instance: %w", err)
	}
	log.Printf("Issuing termination request for instance %s\n", instance.Id)

	return nil
}
