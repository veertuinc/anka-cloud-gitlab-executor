package command

import (
	"strings"

	"veertu.com/anka-cloud-gitlab-executor/internal/ankacloud"
	"veertu.com/anka-cloud-gitlab-executor/internal/gitlab"
)

func getAPIClientConfig(env gitlab.Environment) ankacloud.APIClientConfig {
	apiClientConfig := ankacloud.APIClientConfig{
		BaseURL: env.ControllerURL,
	}

	if strings.HasPrefix(env.ControllerURL, "https") {
		apiClientConfig.IsTLS = true

		if env.CaCertPath != "" {
			apiClientConfig.CaCertPath = env.CaCertPath
		}

		if env.SkipTLSVerify {
			apiClientConfig.SkipTLSVerify = env.SkipTLSVerify
		}

		if env.ClientCertPath != "" {
			apiClientConfig.ClientCertPath = env.ClientCertPath
		}

		if env.ClientCertKeyPath != "" {
			apiClientConfig.ClientCertKeyPath = env.ClientCertKeyPath
		}

	}

	return apiClientConfig
}
