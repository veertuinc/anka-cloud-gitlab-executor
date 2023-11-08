package gitlab

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	prefixGitlabEnvVar    = "CUSTOM_ENV_"
	prefixAnkaCloudEnvVar = "ANKA_CLOUD_"
)

var (
	// Custom Executor vars
	varDebug             = ankaVar("DEBUG")
	varControllerURL     = ankaVar("CONTROLLER_URL")
	varTemplateId        = ankaVar("TEMPLATE_ID")
	varTemplateTag       = ankaVar("TEMPLATE_TAG")
	varNodeId            = ankaVar("NODE_ID")
	varPriority          = ankaVar("PRIORITY")
	varNodeGroupId       = ankaVar("NODE_GROUP_ID")
	varCaCertPath        = ankaVar("CA_CERT_PATH")
	varSkipTLSVerify     = ankaVar("SKIP_TLS_VERIFY")
	varClientCertPath    = ankaVar("CLIENT_CERT_PATH")
	varClientCertKeyPath = ankaVar("CLIENT_CERT_KEY_PATH")
	varSshUserName       = ankaVar("SSH_USER_NAME")
	varSshPassword       = ankaVar("SSH_PASSWORD")

	// Gitlab vars
	varGitlabJobId = gitlabVar("CI_JOB_ID")
)

type Environment struct {
	ControllerURL               string
	Debug                       bool
	TemplateId                  string
	TemplateTag                 string
	NodeId                      string
	Priority                    int
	NodeGroupId                 string
	CaCertPath                  string
	SkipTLSVerify               bool
	ClientCertPath              string
	ClientCertKeyPath           string
	SSHUserName                 string
	SSHPassword                 string
	GitlabJobId                 string
	GitlabBuildFailureExitCode  int
	GitlabSystemFailureExitCode int
}

func InitEnv() (Environment, error) {
	e := Environment{}
	var ok bool
	if e.ControllerURL, ok = os.LookupEnv(varControllerURL); !ok {
		return e, fmt.Errorf("%w: %s", ErrMissingVar, varControllerURL)
	}
	if !strings.HasPrefix(e.ControllerURL, "http") {
		return e, fmt.Errorf("controller url %q missing http prefix", e.ControllerURL)
	}

	if e.GitlabJobId, ok = os.LookupEnv(varGitlabJobId); !ok {
		return e, fmt.Errorf("%w: %s", ErrMissingVar, varGitlabJobId)
	}

	e.SSHUserName = os.Getenv(varSshUserName)
	e.SSHPassword = os.Getenv(varSshPassword)
	e.TemplateId = os.Getenv(varTemplateId)
	e.TemplateTag = os.Getenv(varTemplateTag)
	e.NodeId = os.Getenv(varNodeId)
	e.NodeGroupId = os.Getenv(varNodeGroupId)
	e.CaCertPath = os.Getenv(varCaCertPath)
	e.ClientCertPath = os.Getenv(varClientCertPath)
	e.ClientCertKeyPath = os.Getenv(varClientCertKeyPath)

	if priority, ok, err := GetIntEnvVar(varPriority); ok {
		if err != nil {
			return e, fmt.Errorf("failed to parse priority: %w", err)
		}
		e.Priority = priority
	}

	if debug, ok, err := GetBoolEnvVar(varDebug); ok {
		if err != nil {
			return e, fmt.Errorf("failed to read debug variable: %w", err)
		}
		e.Debug = debug
	}

	if skip, ok, err := GetBoolEnvVar(varSkipTLSVerify); ok {
		if err != nil {
			return e, fmt.Errorf("failed to read debug variable: %w", err)
		}
		e.SkipTLSVerify = skip
	}

	return e, nil
}

func gitlabVar(name string) string {
	return fmt.Sprintf("%s%s", prefixGitlabEnvVar, name)
}

func ankaVar(name string) string {
	return gitlabVar(fmt.Sprintf("%s%s", prefixAnkaCloudEnvVar, name))
}

func GetBoolEnvVar(name string) (bool, bool, error) {
	v, ok := os.LookupEnv(name)
	if !ok {
		return false, false, nil
	}

	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, true, fmt.Errorf("failed to convert variable %s with value %q to boolean: %w", name, v, err)
	}

	return b, true, nil
}

func GetIntEnvVar(name string) (int, bool, error) {
	v, ok := os.LookupEnv(name)
	if !ok {
		return 0, false, nil
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, true, fmt.Errorf("failed to convert variable %s with value %q to boolean: %w", name, v, err)
	}

	return n, true, nil
}
