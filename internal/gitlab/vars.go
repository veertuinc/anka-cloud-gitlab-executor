package gitlab

import (
	"encoding/json"
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
	varCustomHTTPHeaders = ankaVar("CUSTOM_HTTP_HEADERS")
	varKeepAliveOnError  = ankaVar("KEEP_ALIVE_ON_ERROR")
	varTemplateName      = ankaVar("TEMPLATE_NAME")
	varBuildsDir         = ankaVar("BUILDS_DIR")
	varCacheDir          = ankaVar("CACHE_DIR")

	// Gitlab vars
	varGitlabJobId     = gitlabVar("CI_JOB_ID")
	varGitlabJobStatus = gitlabVar("CI_JOB_STATUS")
)

type Environment struct {
	ControllerURL     string
	Debug             bool
	TemplateId        string
	TemplateTag       string
	NodeId            string
	Priority          int
	NodeGroupId       string
	CaCertPath        string
	SkipTLSVerify     bool
	ClientCertPath    string
	ClientCertKeyPath string
	SSHUserName       string
	SSHPassword       string
	GitlabJobId       string
	CustomHttpHeaders map[string]string
	KeepAliveOnError  bool
	GitlabJobStatus   jobStatus
	TemplateName      string
	BuildsDir         string
	CacheDir          string
}

type jobStatus string

var (
	JobStatusSuccess  jobStatus = "success"
	JobStatusFailed   jobStatus = "failed"
	JobStatusCanceled jobStatus = "canceled"
	JobStatusRunning  jobStatus = "running"
)

func InitEnv() (Environment, error) {
	e := Environment{}
	var ok bool
	if e.ControllerURL, ok = os.LookupEnv(varControllerURL); !ok {
		return e, fmt.Errorf("%w: %s", ErrMissingVar, varControllerURL)
	}
	e.ControllerURL = strings.TrimSuffix(e.ControllerURL, "/")
	if !strings.HasPrefix(e.ControllerURL, "http") {
		return e, fmt.Errorf("%w %q: missing http prefix", ErrInvalidVar, e.ControllerURL)
	}

	if e.GitlabJobId, ok = os.LookupEnv(varGitlabJobId); !ok {
		return e, fmt.Errorf("%w: %s", ErrMissingVar, varGitlabJobId)
	}

	e.SSHUserName = os.Getenv(varSshUserName)
	e.SSHPassword = os.Getenv(varSshPassword)
	e.TemplateId = os.Getenv(varTemplateId)
	e.TemplateName = os.Getenv(varTemplateName)
	e.TemplateTag = os.Getenv(varTemplateTag)
	e.NodeId = os.Getenv(varNodeId)
	e.NodeGroupId = os.Getenv(varNodeGroupId)
	e.CaCertPath = os.Getenv(varCaCertPath)
	e.ClientCertPath = os.Getenv(varClientCertPath)
	e.ClientCertKeyPath = os.Getenv(varClientCertKeyPath)
	e.GitlabJobStatus = jobStatus(os.Getenv(varGitlabJobStatus))
	e.BuildsDir = os.Getenv(varBuildsDir)
	e.CacheDir = os.Getenv(varCacheDir)

	if priority, ok, err := GetIntEnvVar(varPriority); ok {
		if err != nil {
			return e, fmt.Errorf("%w %q: %w", ErrInvalidVar, varPriority, err)
		}
		e.Priority = priority
	}

	if debug, ok, err := GetBoolEnvVar(varDebug); ok {
		if err != nil {
			return e, fmt.Errorf("%w %q: %w", ErrInvalidVar, varDebug, err)
		}
		e.Debug = debug
	}

	if skip, ok, err := GetBoolEnvVar(varSkipTLSVerify); ok {
		if err != nil {
			return e, fmt.Errorf("%w %q: %w", ErrInvalidVar, varSkipTLSVerify, err)
		}
		e.SkipTLSVerify = skip
	}

	if customHttpHeaders, ok := os.LookupEnv(varCustomHTTPHeaders); ok {
		err := json.Unmarshal([]byte(customHttpHeaders), &e.CustomHttpHeaders)
		if err != nil {
			return e, fmt.Errorf("%w %q: %w", ErrInvalidVar, varCustomHTTPHeaders, err)
		}
	}

	if keepAlive, ok, err := GetBoolEnvVar(varKeepAliveOnError); ok {
		if err != nil {
			return e, fmt.Errorf("%w %q: %w", ErrInvalidVar, varKeepAliveOnError, err)
		}
		e.KeepAliveOnError = keepAlive
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
