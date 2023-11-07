package gitlab

import (
	"fmt"
	"os"
	"strconv"
)

const (
	GitlabEnvVarPrefix    = "CUSTOM_ENV_"
	AnkaCloudEnvVarPrefix = "ANKA_CLOUD_"
)

var (
	// Custom Executor vars
	VarDebug             = ankaVar("DEBUG")
	VarControllerURL     = ankaVar("CONTROLLER_URL")
	VarTemplateId        = ankaVar("TEMPLATE_ID")
	VarTemplateTag       = ankaVar("TEMPLATE_TAG")
	VarNodeId            = ankaVar("NODE_ID")
	VarPriority          = ankaVar("PRIORITY")
	VarNodeGroupId       = ankaVar("NODE_GROUP_ID")
	VarCaCertPath        = ankaVar("CA_CERT_PATH")
	VarSkipTLSVerify     = ankaVar("SKIP_TLS_VERIFY")
	VarClientCertPath    = ankaVar("CLIENT_CERT_PATH")
	VarClientCertKeyPath = ankaVar("CLIENT_CERT_KEY_PATH")
	VarSshUserName       = ankaVar("SSH_USER_NAME")
	VarSshPassword       = ankaVar("SSH_PASSWORD")

	// Gitlab vars
	VarGitlabJobId           = gitlabVar("CI_JOB_ID")
	VarBuildFailureExitCode  = "BUILD_FAILURE_EXIT_CODE"
	VarSystemFailureExitCode = "SYSTEM_FAILURE_EXIT_CODE"
)

func gitlabVar(name string) string {
	return fmt.Sprintf("%s%s", GitlabEnvVarPrefix, name)
}

func ankaVar(name string) string {
	return gitlabVar(fmt.Sprintf("%s%s", AnkaCloudEnvVarPrefix, name))
}

func GetBoolVar(name string) (bool, bool, error) {
	v, ok := os.LookupEnv(name)
	if !ok {
		return false, false, nil
	}

	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, true, fmt.Errorf("could not convert variable %s with value %q to boolean: %w", name, v, err)
	}

	return b, true, nil
}

func GetIntVar(name string) (int, bool, error) {
	v, ok := os.LookupEnv(name)
	if !ok {
		return 0, false, nil
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, true, fmt.Errorf("could not convert variable %s with value %q to boolean: %w", name, v, err)
	}

	return n, true, nil
}
