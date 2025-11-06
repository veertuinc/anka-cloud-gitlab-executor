package ankacloud

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetInstanceByExternalId_PrioritizesActiveInstances(t *testing.T) {
	// This test reproduces the bug from GitHub issue #40
	// When GitLab retries a failed job, multiple instances can have the same external_id
	// We need to return the active instance, not the failed one

	// Mock server that returns multiple instances with the same external_id
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/vm" {
			t.Errorf("Expected path /api/v1/vm, got %s", r.URL.Path)
		}

		response := getAllInstancesResponse{
			response: response{Status: "OK"},
			Instances: []InstanceWrapper{
				{
					Id:         "1f5dccd8-1364-4699-4ecf-4bd8f6f744ec",
					ExternalId: "https://gitlab.com/job/123",
					Instance: &Instance{
						Id:         "1f5dccd8-1364-4699-4ecf-4bd8f6f744ec",
						ExternalId: "https://gitlab.com/job/123",
						State:      StateError, // Old failed instance
						VMInfo:     nil,
					},
				},
				{
					Id:         "8536878f-ed0d-40d7-5899-80aa45ffa468",
					ExternalId: "https://gitlab.com/job/123",
					Instance: &Instance{
						Id:         "8536878f-ed0d-40d7-5899-80aa45ffa468",
						ExternalId: "https://gitlab.com/job/123",
						State:      StateStarted, // New active instance from retry
						VMInfo: &VM{
							Name: "test-vm",
							PortForwardingRules: []PortForwardingRule{
								{VmPort: 22, NodePort: 10022, Protocol: "tcp"},
							},
						},
						NodeId: "node-123",
					},
				},
			},
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	apiClient := &APIClient{
		ControllerURL: server.URL,
		HttpClient:    server.Client(),
	}
	controller := NewController(apiClient)

	// Get instance by external ID - should return the active one, not the failed one
	instance, err := controller.GetInstanceByExternalId(context.Background(), "https://gitlab.com/job/123")
	if err != nil {
		t.Fatalf("Failed to get instance: %v", err)
	}

	// Verify we got the active instance, not the failed one
	if instance.Id != "8536878f-ed0d-40d7-5899-80aa45ffa468" {
		t.Errorf("Expected to get active instance '8536878f-ed0d-40d7-5899-80aa45ffa468', got '%s'", instance.Id)
	}

	if instance.State != StateStarted {
		t.Errorf("Expected instance state to be Started, got %s", instance.State)
	}

	if instance.VMInfo == nil {
		t.Error("Expected active instance to have VM info")
	}
}

func TestGetInstanceByExternalId_SingleInstance(t *testing.T) {
	// Test the normal case where only one instance exists
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := getAllInstancesResponse{
			response: response{Status: "OK"},
			Instances: []InstanceWrapper{
				{
					Id:         "single-instance",
					ExternalId: "https://gitlab.com/job/456",
					Instance: &Instance{
						Id:         "single-instance",
						ExternalId: "https://gitlab.com/job/456",
						State:      StateStarted,
						VMInfo:     &VM{Name: "test-vm"},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	apiClient := &APIClient{
		ControllerURL: server.URL,
		HttpClient:    server.Client(),
	}
	controller := NewController(apiClient)

	instance, err := controller.GetInstanceByExternalId(context.Background(), "https://gitlab.com/job/456")
	if err != nil {
		t.Fatalf("Failed to get instance: %v", err)
	}

	if instance.Id != "single-instance" {
		t.Errorf("Expected instance ID 'single-instance', got '%s'", instance.Id)
	}
}

func TestGetInstanceByExternalId_NotFound(t *testing.T) {
	// Test that we get an error when no matching instance exists
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := getAllInstancesResponse{
			response: response{Status: "OK"},
			Instances: []InstanceWrapper{
				{
					Id:         "different-instance",
					ExternalId: "https://gitlab.com/job/999",
					Instance: &Instance{
						Id:         "different-instance",
						ExternalId: "https://gitlab.com/job/999",
						State:      StateStarted,
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	apiClient := &APIClient{
		ControllerURL: server.URL,
		HttpClient:    server.Client(),
	}
	controller := NewController(apiClient)

	_, err := controller.GetInstanceByExternalId(context.Background(), "https://gitlab.com/job/404")
	if err == nil {
		t.Error("Expected error for non-existent instance, got nil")
	}
}

func TestGetInstanceByExternalId_PrioritizesSchedulingOverError(t *testing.T) {
	// Test that Scheduling state is prioritized over Error state
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := getAllInstancesResponse{
			response: response{Status: "OK"},
			Instances: []InstanceWrapper{
				{
					Id:         "error-instance",
					ExternalId: "https://gitlab.com/job/789",
					Instance: &Instance{
						Id:         "error-instance",
						ExternalId: "https://gitlab.com/job/789",
						State:      StateError,
					},
				},
				{
					Id:         "scheduling-instance",
					ExternalId: "https://gitlab.com/job/789",
					Instance: &Instance{
						Id:         "scheduling-instance",
						ExternalId: "https://gitlab.com/job/789",
						State:      StateScheduling,
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	apiClient := &APIClient{
		ControllerURL: server.URL,
		HttpClient:    server.Client(),
	}
	controller := NewController(apiClient)

	instance, err := controller.GetInstanceByExternalId(context.Background(), "https://gitlab.com/job/789")
	if err != nil {
		t.Fatalf("Failed to get instance: %v", err)
	}

	if instance.Id != "scheduling-instance" {
		t.Errorf("Expected scheduling instance, got '%s'", instance.Id)
	}

	if instance.State != StateScheduling {
		t.Errorf("Expected state Scheduling, got %s", instance.State)
	}
}
