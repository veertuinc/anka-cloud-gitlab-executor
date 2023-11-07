package ankacloud

import (
	"context"
	"encoding/json"
	"fmt"

	"time"

	"veertu.com/anka-cloud-gitlab-executor/internal/log"
)

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

type CreateInstanceRequest struct {
	TemplateId  string `json:"vmid"`
	ExternalId  string `json:"external_id,omitempty"`
	Tag         string `json:"tag,omitempty"`
	NodeId      string `json:"node_id,omitempty"`
	Priority    int    `json:"priority,omitempty"`
	NodeGroupId string `json:"group_id,omitempty"`
}

type createInstanceResponse struct {
	response
	InstanceIds []string `json:"body"`
}

type GetInstanceRequest struct {
	Id string
}

type getInstanceResponse struct {
	response
	Instance Instance `json:"body"`
}

type TerminateInstanceRequest struct {
	Id string `json:"id"`
}

type terminateInstanceResponse response

type getAllInstancesResponse struct {
	response
	Instances []InstanceWrapper `json:"body"`
}

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

func (c *Client) GetInstance(ctx context.Context, req GetInstanceRequest) (*Instance, error) {
	body, err := c.Get(ctx, "/api/v1/vm", map[string]string{"id": req.Id})
	if err != nil {
		return nil, fmt.Errorf("failed getting instance %s: %w", req.Id, err)
	}

	var response getInstanceResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed parsing response body %q: %w", string(body), err)
	}

	return &response.Instance, nil
}

func (c *Client) CreateInstance(ctx context.Context, payload CreateInstanceRequest) (string, error) {

	if payload.Priority < 0 || payload.Priority > 10000 {
		return "", fmt.Errorf("priority must be between 1 and 10000. Got %d", payload.Priority)
	}

	body, err := c.Post(ctx, "/api/v1/vm", payload)
	if err != nil {
		return "", fmt.Errorf("failed creating instance %+v: %w", payload, err)
	}

	var response createInstanceResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("failed parsing response body %q: %w", string(body), err)
	}

	return response.InstanceIds[0], nil
}

func (c *Client) WaitForInstanceToBeScheduled(ctx context.Context, instanceId string) error {
	const pollingInterval = 3 * time.Second

	log.Printf("waiting for instance %s to be scheduled\n", instanceId)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pollingInterval):
			instance, err := c.GetInstance(ctx, GetInstanceRequest{Id: instanceId})
			if err != nil {
				return fmt.Errorf("failed getting instance %q status: %w", instanceId, err)
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

func (c *Client) TerminateInstance(ctx context.Context, payload TerminateInstanceRequest) error {
	body, err := c.Delete(ctx, "/api/v1/vm", payload)
	if err != nil {
		return fmt.Errorf("failed terminating instance %+v: %w", payload, err)
	}

	var response terminateInstanceResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return fmt.Errorf("failed parsing response body %q: %w", string(body), err)
	}

	return nil
}

func (c *Client) GetAllInstances(ctx context.Context) ([]InstanceWrapper, error) {

	body, err := c.Get(ctx, "/api/v1/vm", nil)
	if err != nil {
		return nil, fmt.Errorf("failed getting instances: %w", err)
	}

	var response getAllInstancesResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed parsing response body %q: %w", string(body), err)
	}

	return response.Instances, nil
}

func (c *Client) GetInstanceByExternalId(ctx context.Context, externalId string) (*Instance, error) {
	instances, err := c.GetAllInstances(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed getting instances: %w", err)
	}

	for _, instance := range instances {
		if instance.ExternalId == externalId {
			return instance.Instance, nil
		}
	}

	return nil, fmt.Errorf("instance with external id %s not found", externalId)
}
