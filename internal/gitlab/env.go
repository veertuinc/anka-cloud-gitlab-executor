package gitlab

import (
	"fmt"
	"os"
)

const (
	GITLAB_ENV_VAR_PREFIX     = "CUSTOM_ENV_"
	ANKA_CLOUD_ENV_VAR_PREFIX = "ANKA_CLOUD_"
)

var ErrEnvVarNotSet = fmt.Errorf("env var is not set")

func GetGitlabEnvVar(name string) (string, error) {
	return getEnvVar(fmt.Sprintf("%s%s", GITLAB_ENV_VAR_PREFIX, name))
}

func GetAnkaCloudEnvVar(name string) (string, error) {
	return GetGitlabEnvVar(fmt.Sprintf("%s%s", ANKA_CLOUD_ENV_VAR_PREFIX, name))
}

func getEnvVar(name string) (string, error) {
	v, ok := os.LookupEnv(name)
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrEnvVarNotSet, name)
	}
	return v, nil
}
