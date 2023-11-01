package env

import (
	"fmt"
	"os"
)

const (
	GITLAB_ENV_VAR_PREFIX     = "CUSTOM_ENV_"
	ANKA_CLOUD_ENV_VAR_PREFIX = "ANKA_CLOUD_"
)

func GitlabVar(name string) string {
	return fmt.Sprintf("%s%s", GITLAB_ENV_VAR_PREFIX, name)
}

func AnkaVar(name string) string {
	return GitlabVar(fmt.Sprintf("%s%s", ANKA_CLOUD_ENV_VAR_PREFIX, name))
}

func Get(name string) (string, bool) {
	return os.LookupEnv(name)
}
