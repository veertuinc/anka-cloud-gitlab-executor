package ankacloud

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/veertuinc/anka-cloud-gitlab-executor/internal/log"
)

func configureTLS(config APIClientConfig) (*tls.Config, error) {
	log.Debugln("handling TLS configuration")

	tlsConfig := &tls.Config{}
	caCertPool, _ := x509.SystemCertPool()
	if caCertPool == nil {
		caCertPool = x509.NewCertPool()
	}
	tlsConfig.RootCAs = caCertPool

	if config.CaCertPath != "" {
		if err := appendRootCert(config.CaCertPath, caCertPool); err != nil {
			return nil, fmt.Errorf("failed to add CA cert from %q to pool: %w", config.CaCertPath, err)
		}
		log.ConditionalColorf("using CA cert from %q\n", config.CaCertPath)
	}

	if config.SkipTLSVerify {
		log.ConditionalColorf("allowing to skip server host verification")
		tlsConfig.InsecureSkipVerify = true
	}

	if config.certAuthEnabled() {
		cert, err := tls.LoadX509KeyPair(config.ClientCertPath, config.ClientCertKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to process key pair (cert at %q, key at %q): %w", config.ClientCertPath, config.ClientCertKeyPath, err)
		}

		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}

func appendRootCert(certFilePath string, caCertPool *x509.CertPool) error {
	cert, err := os.ReadFile(certFilePath)
	if err != nil {
		return fmt.Errorf("failed to read file at %q: %w", certFilePath, err)
	}
	ok := caCertPool.AppendCertsFromPEM(cert)
	if !ok {
		return fmt.Errorf("failed to add cert at %q to cert pool: %w", certFilePath, err)
	}
	return nil
}
