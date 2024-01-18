package gitlab

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestControllerWithTrailingSlash(t *testing.T) {
	os.Setenv(varControllerURL, "http://fake-controller-url/")
	os.Setenv(varGitlabJobId, "fake-gitlab-job-id")
	defer os.Clearenv()

	env, err := InitEnv()
	if err != nil {
		t.Error(err)
	}

	if env.ControllerURL != "http://fake-controller-url" {
		t.Errorf("expected controller url %q, got %q", "http://fake-controller-url", env.ControllerURL)
	}
}

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

func TestInvalidVram(t *testing.T) {
	defer os.Clearenv()

	testCases := []struct {
		name string
		vram string
	}{
		{
			name: "negative vram",
			vram: "-1",
		},
		{
			name: "zero vram",
			vram: "0",
		},
		{
			name: "non-numeric vram",
			vram: "fake-vram",
		},
	}

	for _, tc := range testCases {
		os.Clearenv()
		os.Setenv(varControllerURL, "http://fake-controller-url")
		os.Setenv(varGitlabJobId, "fake-gitlab-job-id")

		os.Setenv(varVmVcpu, tc.vram)

		_, err := InitEnv()
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if !errors.Is(err, ErrInvalidVar) {
			t.Errorf("expected error %q, got %q", ErrInvalidVar, err)
		}
	}
}

func TestInvalidVcpu(t *testing.T) {
	defer os.Clearenv()

	testCases := []struct {
		name string
		vcpu string
	}{
		{
			name: "negative vcpu",
			vcpu: "-1",
		},
		{
			name: "zero vcpu",
			vcpu: "0",
		},
		{
			name: "non-numeric vcpu",
			vcpu: "fake-vram",
		},
	}

	for _, tc := range testCases {
		os.Clearenv()
		os.Setenv(varControllerURL, "http://fake-controller-url")
		os.Setenv(varGitlabJobId, "fake-gitlab-job-id")

		os.Setenv(varVmVcpu, tc.vcpu)

		_, err := InitEnv()
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if !errors.Is(err, ErrInvalidVar) {
			t.Errorf("expected error %q, got %q", ErrInvalidVar, err)
		}
	}
}
