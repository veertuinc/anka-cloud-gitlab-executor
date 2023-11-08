package ankacloud

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"time"

	"veertu.com/anka-cloud-gitlab-executor/internal/log"
)

type HttpClientConfig struct {
	IsTLS               bool
	CaCertPath          string
	ClientCertPath      string
	ClientCertKeyPath   string
	SkipTLSVerify       bool
	MaxIdleConnsPerHost int
	RequestTimeout      time.Duration
}

func (c *HttpClientConfig) certAuthEnabled() bool {
	return c.ClientCertKeyPath != "" && c.ClientCertPath != ""
}

const (
	defaultMaxIdleConnsPerHost = 20
	defaultRequestTimeout      = 10 * time.Second
)

func NewHTTPClient(config *HttpClientConfig) (*http.Client, error) {
	client := &http.Client{
		Timeout: defaultRequestTimeout,
	}

	if config.RequestTimeout > 0 {
		client.Timeout = config.RequestTimeout
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConnsPerHost = defaultMaxIdleConnsPerHost

	if config.MaxIdleConnsPerHost > 0 {
		transport.MaxIdleConnsPerHost = config.MaxIdleConnsPerHost
	}

	if config.IsTLS {
		tlsConfig, err := configureTLS(config)
		if err != nil {
			return nil, fmt.Errorf("failed to configure TLS %+v: %w", config, err)
		}
		transport.TLSClientConfig = tlsConfig
	}

	client.Transport = transport

	return client, nil
}

func configureTLS(config *HttpClientConfig) (*tls.Config, error) {
	log.Println("Handling TLS configuration")

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
		log.Printf("Added CA cert at %q\n", config.CaCertPath)
	}

	if config.SkipTLSVerify {
		log.Println("Allowing to skip server host verification")
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
