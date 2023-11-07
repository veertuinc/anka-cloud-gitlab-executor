package ankacloud

import (
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

func (c *Client) GetInstance(req GetInstanceRequest) (*Instance, error) {
	body, err := c.Get("/api/v1/vm", map[string]string{"id": req.Id})
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

func (c *Client) CreateInstance(payload CreateInstanceRequest) (string, error) {

	if payload.Priority < 0 || payload.Priority > 10000 {
		return "", fmt.Errorf("priority must be between 1 and 10000. Got %d", payload.Priority)
	}

	body, err := c.Post("/api/v1/vm", payload)
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

func (c *Client) WaitForInstanceToBeScheduled(instanceId string) error {
	log.Printf("waiting for instance %s to be scheduled\n", instanceId)
	for {
		instance, err := c.GetInstance(GetInstanceRequest{Id: instanceId})
		if err != nil {
			return fmt.Errorf("failed getting instance %q status: %w", instanceId, err)
		}

		log.Printf("instance %s is in state %q\n", instanceId, instance.State)
		switch instance.State {
		case StateScheduling, StatePulling:
			time.Sleep(3 * time.Second)
			continue
		}

		if instance.State == StateStarted {
			break
		}

		return fmt.Errorf("instance is in an unexpected state: %s", instance.State)
	}

	return nil
}

func (c *Client) TerminateInstance(payload TerminateInstanceRequest) error {
	body, err := c.Delete("/api/v1/vm", payload)
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

func (c *Client) GetAllInstances() ([]InstanceWrapper, error) {

	body, err := c.Get("/api/v1/vm", nil)
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

func (c *Client) GetInstanceByExternalId(externalId string) (*Instance, error) {
	instances, err := c.GetAllInstances()
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
