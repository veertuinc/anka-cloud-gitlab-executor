package command

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/veertuinc/anka-cloud-gitlab-executor/internal/ankacloud"
	"github.com/veertuinc/anka-cloud-gitlab-executor/internal/gitlab"
	"github.com/veertuinc/anka-cloud-gitlab-executor/internal/log"
)

var prepareCommand = &cobra.Command{
	Use: "prepare",
	RunE: func(cmd *cobra.Command, args []string) error {
		env, ok := cmd.Context().Value(contextKey("env")).(gitlab.Environment)
		if !ok {
			return fmt.Errorf("failed to get environment from context")
		}

		return executePrepare(cmd.Context(), env)
	},
}

func executePrepare(ctx context.Context, env gitlab.Environment) error {
	log.SetOutput(os.Stderr)
	log.Debugln("running prepare stage")

	apiClientConfig := getAPIClientConfig(env)
	apiClient, err := ankacloud.NewAPIClient(apiClientConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize API client with config +%v: %w", apiClientConfig, err)
	}
	controller := ankacloud.NewController(apiClient)

	var template string
	templateId := env.TemplateId
	if templateId == "" {
		if env.TemplateName == "" {
			return fmt.Errorf("%w: either template id or temaplte name must be specified", gitlab.ErrMissingVar)
		}
		log.Warnln("please consider using template id instead of template name as template names are not guaranteed to be unique")
		templateId, err = controller.GetTemplateIdByName(ctx, env.TemplateName)
		if err != nil {
			return fmt.Errorf("failed to get template id of template named %q: %w", env.TemplateName, err)
		}
		log.Colorf("template with id %q and name %q will be used\n", templateId, env.TemplateName)
		template = env.TemplateName
	} else {
		template = templateId
	}

	req := ankacloud.CreateInstanceRequest{
		TemplateId:              templateId,
		ExternalId:              env.GitlabJobUrl,
		Tag:                     env.TemplateTag,
		NodeId:                  env.NodeId,
		Priority:                env.Priority,
		NodeGroupId:             env.NodeGroupId,
		StartupScriptCondition:  ankacloud.WaitForNetwork,
		StartupScriptMonitoring: true,
		StartupScriptTimeout:    5 * 60,
		StartupScript:           base64.StdEncoding.EncodeToString([]byte("sleep 5")), // even though we wait for network, it is recommended to wait a bit more
		Vcpu:                    env.VmVcpu,
		VramMb:                  env.VmVramMb,
	}

	var tagName string
	if env.TemplateTag == "" {
		tagName = "(latest)"
	}

	log.Colorf("Creating macOS VM with Template %q and Tag %q -- please be patient...", template, tagName)
	log.Debugf("payload %+v\n", req)
	instanceId, err := controller.CreateInstance(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create instance: %w", err)
	}

	instance, err := controller.WaitForInstanceToBeScheduled(ctx, instanceId)
	if err != nil {
		return gitlab.TransientError(fmt.Errorf("failed to wait for instance %q to be scheduled: %w", instanceId, err))
	}

	log.Colorf("VM %s (%s) is ready for work on node %s (%s)\n", instance.VMInfo.Name, instance.Id, instance.Node.Name, instance.Node.IP)

	return nil
}
