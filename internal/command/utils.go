package command

import (
	"strings"

	"veertu.com/anka-cloud-gitlab-executor/internal/ankacloud"
	"veertu.com/anka-cloud-gitlab-executor/internal/gitlab"
)

func getHttpClientConfig(env gitlab.Environment) (*ankacloud.HttpClientConfig, error) {
	httpClientConfig := ankacloud.HttpClientConfig{}

	if strings.HasPrefix(env.ControllerURL, "https") {
		httpClientConfig.IsTLS = true

		if env.CaCertPath != "" {
			httpClientConfig.CaCertPath = env.CaCertPath
		}

		if env.SkipTLSVerify {
			httpClientConfig.SkipTLSVerify = env.SkipTLSVerify
		}

		if env.ClientCertPath != "" {
			httpClientConfig.ClientCertPath = env.ClientCertPath
		}

		if env.ClientCertKeyPath != "" {
			httpClientConfig.ClientCertKeyPath = env.ClientCertKeyPath
		}

	}

	return &httpClientConfig, nil
}
