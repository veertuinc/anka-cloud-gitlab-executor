package env

import (
	"errors"
	"fmt"
)

const (
	GitlabEnvVarPrefix    = "CUSTOM_ENV_"
	AnkaCloudEnvVarPrefix = "ANKA_CLOUD_"
)

var (
	// Custom Executor vars
	VarDebug         = ankaVar("DEBUG")
	VarControllerURL = ankaVar("CONTROLLER_URL")
	VarTemplateId    = ankaVar("TEMPLATE_ID")
	VarTemplateTag   = ankaVar("TEMPLATE_TAG")
	VarNodeId        = ankaVar("NODE_ID")
	VarPriority      = ankaVar("PRIORITY")
	VarNodeGroupId   = ankaVar("NODE_GROUP_ID")

	// Gitlab vars
	VarGitlabJobId = gitlabVar("CI_JOB_ID")
)

var ErrMissingVar = errors.New("missing environment variable")

func gitlabVar(name string) string {
	return fmt.Sprintf("%s%s", GitlabEnvVarPrefix, name)
}

func ankaVar(name string) string {
	return gitlabVar(fmt.Sprintf("%s%s", AnkaCloudEnvVarPrefix, name))
}
