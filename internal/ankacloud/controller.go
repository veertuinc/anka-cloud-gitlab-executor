package ankacloud

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"veertu.com/anka-cloud-gitlab-executor/internal/log"
)

type controller struct {
	APIClient *APIClient
}

type Node struct {
	IP string `json:"ip_address"`
}

type InstanceState string

const (
	StateScheduling  InstanceState = "Scheduling"
	StatePulling     InstanceState = "Pulling"
	StateStarted     InstanceState = "Started"
	StateTerminating InstanceState = "Terminating"
	StateTerminated  InstanceState = "Terminated"
	StateError       InstanceState = "Error"
	StatePushing     InstanceState = "Pushing"
)

type VM struct {
	PortForwardingRules []PortForwardingRule `json:"port_forwarding,omitempty"`
}
type PortForwardingRule struct {
	VmPort   int    `json:"guest_port"`
	NodePort int    `json:"host_port"`
	Protocol string `json:"protocol"`
}

type Instance struct {
	State      InstanceState `json:"instance_state"`
	Id         string        `json:"instance_id"`
	ExternalId string        `json:"external_id"`
	VM         *VM           `json:"vminfo,omitempty"`
	NodeId     string        `json:"node_id,omitempty"`
}

type InstanceWrapper struct {
	Id         string    `json:"instance_id"`
	ExternalId string    `json:"external_id"`
	Instance   *Instance `json:"vm,omitempty"`
}

func NewController(apiClient *APIClient) *controller {
	return &controller{
		APIClient: apiClient,
	}
}
func (c *controller) GetNode(ctx context.Context, req GetNodeRequest) (*Node, error) {
	body, err := c.APIClient.Get(ctx, "/api/v1/node", map[string]string{"id": req.Id})
	if err != nil {
		return nil, fmt.Errorf("failed to get node %q: %w", req.Id, err)
	}

	var response getNodeResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body %q: %w", string(body), err)
	}

	if len(response.Nodes) == 0 {
		return nil, fmt.Errorf("node %s not found", req.Id)
	}

	return &response.Nodes[0], nil
}

func (c *controller) GetInstance(ctx context.Context, req GetInstanceRequest) (*Instance, error) {
	body, err := c.APIClient.Get(ctx, "/api/v1/vm", map[string]string{"id": req.Id})
	if err != nil {
		return nil, fmt.Errorf("failed to get instance %s: %w", req.Id, err)
	}

	var response getInstanceResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body %q: %w", string(body), err)
	}

	return &response.Instance, nil
}

func (c *controller) CreateInstance(ctx context.Context, payload CreateInstanceRequest) (string, error) {

	if payload.Priority < 0 || payload.Priority > 10000 {
		return "", fmt.Errorf("priority must be between 1 and 10000. Got %d", payload.Priority)
	}

	body, err := c.APIClient.Post(ctx, "/api/v1/vm", payload)
	if err != nil {
		return "", fmt.Errorf("failed to create instance %+v: %w", payload, err)
	}

	var response createInstanceResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("failed to parse response body %q: %w", string(body), err)
	}

	return response.InstanceIds[0], nil
}

func (c *controller) WaitForInstanceToBeScheduled(ctx context.Context, instanceId string) error {
	const pollingInterval = 3 * time.Second

	log.Printf("waiting for instance %s to be scheduled\n", instanceId)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pollingInterval):
			instance, err := c.GetInstance(ctx, GetInstanceRequest{Id: instanceId})
			if err != nil {
				return fmt.Errorf("failed to get instance %q status: %w", instanceId, err)
			}

			log.Printf("instance %s is in state %q\n", instanceId, instance.State)
			switch instance.State {
			case StateScheduling, StatePulling:
				break
			case StateStarted:
				return nil
			default:
				return fmt.Errorf("instance %s is in an unexpected state: %s", instanceId, instance.State)
			}
		}
	}
}

func (c *controller) TerminateInstance(ctx context.Context, payload TerminateInstanceRequest) error {
	body, err := c.APIClient.Delete(ctx, "/api/v1/vm", payload)
	if err != nil {
		return fmt.Errorf("failed to terminate instance %+v: %w", payload, err)
	}

	var response terminateInstanceResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return fmt.Errorf("failed to parse response body %q: %w", string(body), err)
	}

	return nil
}

func (c *controller) GetAllInstances(ctx context.Context) ([]InstanceWrapper, error) {

	body, err := c.APIClient.Get(ctx, "/api/v1/vm", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get instances: %w", err)
	}

	var response getAllInstancesResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body %q: %w", string(body), err)
	}

	return response.Instances, nil
}

func (c *controller) GetInstanceByExternalId(ctx context.Context, externalId string) (*Instance, error) {
	instances, err := c.GetAllInstances(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get instances: %w", err)
	}

	for _, instance := range instances {
		if instance.ExternalId == externalId {
			return instance.Instance, nil
		}
	}

	return nil, fmt.Errorf("instance with external id %s not found", externalId)
}
