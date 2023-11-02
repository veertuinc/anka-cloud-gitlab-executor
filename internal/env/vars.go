package env

import (
	"errors"
	"fmt"
)

const (
	GITLAB_ENV_VAR_PREFIX     = "CUSTOM_ENV_"
	ANKA_CLOUD_ENV_VAR_PREFIX = "ANKA_CLOUD_"
)

var (
	// Custom Executor vars
	VAR_DEBUG          = ankaVar("DEBUG")
	VAR_CONTROLLER_URL = ankaVar("CONTROLLER_URL")
	VAR_TEMPLATE_ID    = ankaVar("TEMPLATE_ID")
	VAR_TEMPLATE_TAG   = ankaVar("TEMPLATE_TAG")
	VAR_NODE_ID        = ankaVar("NODE_ID")
	VAR_PRIORITY       = ankaVar("PRIORITY")
	VAR_NODE_GROUP_ID  = ankaVar("NODE_GROUP_ID")

	// Gitlab vars
	VAR_GITLAB_JOB_ID = gitlabVar("CI_JOB_ID")
)

var ErrMissingVar = errors.New("missing environment variable")

func gitlabVar(name string) string {
	return fmt.Sprintf("%s%s", GITLAB_ENV_VAR_PREFIX, name)
}

func ankaVar(name string) string {
	return gitlabVar(fmt.Sprintf("%s%s", ANKA_CLOUD_ENV_VAR_PREFIX, name))
}
