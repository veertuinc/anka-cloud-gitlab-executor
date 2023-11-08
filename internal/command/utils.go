package command

import (
	"fmt"
	"os"
	"strings"

	"veertu.com/anka-cloud-gitlab-executor/internal/ankacloud"
	"veertu.com/anka-cloud-gitlab-executor/internal/gitlab"
)

func httpClientConfigFromEnvVars(controllerURL string) (*ankacloud.HttpClientConfig, error) {
	httpClientConfig := ankacloud.HttpClientConfig{}

	if strings.HasPrefix(controllerURL, "https") {
		httpClientConfig.IsTLS = true

		if caCertPath, ok := os.LookupEnv(gitlab.VarCaCertPath); ok {
			httpClientConfig.CaCertPath = caCertPath
		}

		if skipTLSVerify, ok, err := gitlab.GetBoolVar(gitlab.VarSkipTLSVerify); ok {
			if err != nil {
				return nil, fmt.Errorf("failed to parse skip TLS verify: %w", err)
			}
			httpClientConfig.SkipTLSVerify = skipTLSVerify
		}

		if clientCertPath, ok := os.LookupEnv(gitlab.VarClientCertPath); ok {
			httpClientConfig.ClientCertPath = clientCertPath
		}

		if clientCertKeyPath, ok := os.LookupEnv(gitlab.VarClientCertKeyPath); ok {
			httpClientConfig.ClientCertKeyPath = clientCertKeyPath
		}
	}

	return &httpClientConfig, nil
}
