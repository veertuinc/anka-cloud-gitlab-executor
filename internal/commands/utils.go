package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"veertu.com/anka-cloud-gitlab-executor/internal/ankacloud"
	"veertu.com/anka-cloud-gitlab-executor/internal/env"
)

func httpClientConfigFromEnvVars(controllerURL string) (*ankacloud.HttpClientConfig, error) {
	httpClientConfig := ankacloud.HttpClientConfig{}

	if strings.HasPrefix(controllerURL, "https") {
		httpClientConfig.IsTLS = true

		caCertPath, ok := os.LookupEnv(env.VarCaCertPath)
		if ok {
			httpClientConfig.CaCertPath = caCertPath
		}

		skipTLSVerify, ok := os.LookupEnv(env.VarSkipTLSVerify)
		if ok {
			skip, err := strconv.ParseBool(skipTLSVerify)
			if err != nil {
				return nil, fmt.Errorf("could not convert variable %s with value %q to boolean: %w", env.VarSkipTLSVerify, skipTLSVerify, err)
			}
			httpClientConfig.SkipTLSVerify = skip
		}

		clientCertPath, ok := os.LookupEnv(env.VarClientCertPath)
		if ok {
			httpClientConfig.ClientCertPath = clientCertPath
		}

		clientCertKeyPath, ok := os.LookupEnv(env.VarClientCertKeyPath)
		if ok {
			httpClientConfig.ClientCertKeyPath = clientCertKeyPath
		}
	}

	return &httpClientConfig, nil
}
