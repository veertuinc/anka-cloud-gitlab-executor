package gitlab

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestCustomHttpHeadersEnvVar(t *testing.T) {
	os.Setenv(varControllerURL, "http://fake-controller-url")
	os.Setenv(varGitlabJobId, "fake-gitlab-job-id")
	os.Setenv(varCustomHTTPHeaders, "{\"fake-header\":\"fake-value\", \"fake-header2\":\"fake-value2\"}")
	defer os.Clearenv()
	env, err := InitEnv()
	if err != nil {
		t.Error(err)
	}
	if env.CustomHttpHeaders["fake-header"] != "fake-value" {
		t.Errorf("expected header %q to be %q, got %q", "fake-header", "fake-value", env.CustomHttpHeaders["fake-header"])
	}
	if env.CustomHttpHeaders["fake-header2"] != "fake-value2" {
		t.Errorf("expected header %q to be %q, got %q", "fake-header2", "fake-value2", env.CustomHttpHeaders["fake-header2"])
	}
}

func TestControllerURLInvalid(t *testing.T) {
	os.Setenv(varControllerURL, "fake-controller-url")
	defer os.Clearenv()
	_, err := InitEnv()
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "http") {
		t.Errorf("expected error to contain %q, got %q", "http", err)
	}
}

func TestControllerURLMissing(t *testing.T) {
	_, err := InitEnv()
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if !errors.Is(err, ErrMissingVar) {
		t.Errorf("expected error %q, got %q", ErrMissingVar, err)
	}
	if !strings.Contains(err.Error(), varControllerURL) {
		t.Errorf("expected error to contain %q, got %q", varControllerURL, err)
	}
}

func TestGitlabJobIdMissing(t *testing.T) {
	os.Setenv(varControllerURL, "http://fake-controller-url")
	defer os.Clearenv()
	_, err := InitEnv()
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if !errors.Is(err, ErrMissingVar) {
		t.Errorf("expected error %q, got %q", ErrMissingVar, err)
	}
	if !strings.Contains(err.Error(), varGitlabJobId) {
		t.Errorf("expected error to contain %q, got %q", varGitlabJobId, err)
	}
}
