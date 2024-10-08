package gitlab

import (
	"encoding/json"
	"fmt"

	"os"
	"strconv"
	"strings"

	flag "github.com/spf13/pflag"
)

const (
	prefixGitlabEnvVar    = "CUSTOM_ENV_"
	prefixAnkaCloudEnvVar = "ANKA_CLOUD_"
)

var (
	// Custom Executor vars
	varDebug                     = ankaVar("DEBUG")
	varQuietLogging              = ankaVar("QUIETER_LOGGING")
	varControllerURL             = ankaVar("CONTROLLER_URL")
	varTemplateId                = ankaVar("TEMPLATE_ID")
	varTemplateTag               = ankaVar("TEMPLATE_TAG")
	varNodeId                    = ankaVar("NODE_ID")
	varPriority                  = ankaVar("PRIORITY")
	varNodeGroupId               = ankaVar("NODE_GROUP_ID")
	varCaCertPath                = ankaVar("CA_CERT_PATH")
	varSkipTLSVerify             = ankaVar("SKIP_TLS_VERIFY")
	varClientCertPath            = ankaVar("CLIENT_CERT_PATH")
	varClientCertKeyPath         = ankaVar("CLIENT_CERT_KEY_PATH")
	varSshUserName               = ankaVar("SSH_USER_NAME")
	varSshPassword               = ankaVar("SSH_PASSWORD")
	varSshAttempts               = ankaVar("SSH_CONNECTION_ATTEMPTS")
	varSshConnectionAttemptDelay = ankaVar("SSH_CONNECTION_ATTEMPT_DELAY")
	varCustomHTTPHeaders         = ankaVar("CUSTOM_HTTP_HEADERS")
	varKeepAliveOnError          = ankaVar("KEEP_ALIVE_ON_ERROR")
	varTemplateName              = ankaVar("TEMPLATE_NAME")
	varBuildsDir                 = ankaVar("BUILDS_DIR")
	varCacheDir                  = ankaVar("CACHE_DIR")
	varVmVramMb                  = ankaVar("VM_VRAM_MB")
	varVmVcpu                    = ankaVar("VM_VCPU")

	// Gitlab vars
	varGitlabJobUrl    = gitlabVar("CI_JOB_URL")
	varGitlabJobStatus = gitlabVar("CI_JOB_STATUS")
)

type Environment struct {
	ControllerURL             string
	Debug                     bool
	QuietLogging              bool
	TemplateId                string
	TemplateTag               string
	NodeId                    string
	Priority                  int
	NodeGroupId               string
	CaCertPath                string
	SkipTLSVerify             bool
	ClientCertPath            string
	ClientCertKeyPath         string
	SSHUserName               string
	SSHPassword               string
	SSHAttempts               int
	SSHConnectionAttemptDelay int
	GitlabJobUrl              string
	CustomHttpHeaders         map[string]string
	KeepAliveOnError          bool
	GitlabJobStatus           jobStatus
	TemplateName              string
	BuildsDir                 string
	CacheDir                  string
	VmVramMb                  int
	VmVcpu                    int
}

type jobStatus string

var (
	JobStatusSuccess  jobStatus = "success"
	JobStatusFailed   jobStatus = "failed"
	JobStatusCanceled jobStatus = "canceled"
	JobStatusRunning  jobStatus = "running"
)

var sshPassword = flag.String("ssh-password", "", "the password used to SSH into the VM")
var sshUserName = flag.String("ssh-username", "", "the username used to SSH into the VM")

func InitEnv() (Environment, error) {
	// parse command line flags defined above
	flag.Parse()
	e := Environment{
		// load initial values from command line flags
		SSHPassword: *sshPassword,
		SSHUserName: *sshUserName,
	}

	var ok bool
	if e.ControllerURL, ok = os.LookupEnv(varControllerURL); !ok {
		return e, fmt.Errorf("%w: %s", ErrMissingVar, varControllerURL)
	}
	e.ControllerURL = strings.TrimSuffix(e.ControllerURL, "/")
	if !strings.HasPrefix(e.ControllerURL, "http") {
		return e, fmt.Errorf("%w %q: missing http prefix", ErrInvalidVar, e.ControllerURL)
	}

	if e.GitlabJobUrl, ok = os.LookupEnv(varGitlabJobUrl); !ok {
		return e, fmt.Errorf("%w: %s", ErrMissingVar, varGitlabJobUrl)
	}

	if os.Getenv(varSshUserName) != "" {
		e.SSHUserName = os.Getenv(varSshUserName)
	}
	if os.Getenv(varSshPassword) != "" {
		e.SSHPassword = os.Getenv(varSshPassword)
	}

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

	if quietLogging, ok, err := GetBoolEnvVar(varQuietLogging); ok {
		if err != nil {
			return e, fmt.Errorf("%w %q: %w", ErrInvalidVar, varQuietLogging, err)
		}
		e.QuietLogging = quietLogging
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

	if vram, ok, err := GetIntEnvVar(varVmVramMb); ok {
		if err != nil {
			return e, fmt.Errorf("%w %q: %w", ErrInvalidVar, varVmVramMb, err)
		}
		if ok && vram < 1 {
			return e, fmt.Errorf("%w vram must be 1 or higher", ErrInvalidVar)
		}
		e.VmVramMb = vram
	}

	if vcpu, ok, err := GetIntEnvVar(varVmVcpu); ok {
		if err != nil {
			return e, fmt.Errorf("%w %q: %w", ErrInvalidVar, varVmVcpu, err)
		}
		if ok && vcpu < 1 {
			return e, fmt.Errorf("%w vcpu must be 1 or higher", ErrInvalidVar)
		}
		e.VmVcpu = vcpu
	}

	if sshAttempts, ok, err := GetIntEnvVar(varSshAttempts); ok {
		if err != nil {
			return e, fmt.Errorf("%w %q: %w", ErrInvalidVar, varSshAttempts, err)
		}
		e.SSHAttempts = sshAttempts
	}

	if sshConnectionAttemptDelay, ok, err := GetIntEnvVar(varSshConnectionAttemptDelay); ok {
		if err != nil {
			return e, fmt.Errorf("%w %q: %w", ErrInvalidVar, varSshConnectionAttemptDelay, err)
		}
		e.SSHConnectionAttemptDelay = sshConnectionAttemptDelay
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
