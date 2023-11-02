package ankaCloud

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

type createInstanceRequestPayload struct {
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

type CreateInstanceConfig struct {
	TemplateId  string
	TemplateTag string
	ExternalId  string
	NodeId      string
	Priority    int
	NodeGroupId string
}

type GetInstanceConfig struct {
	Id string
}

type getInstanceResponse struct {
	response
	Instance Instance `json:"body"`
}

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
	VM         VM            `json:"vminfo,omitempty"`
	NodeId     string        `json:"node_id,omitempty"`
}

type InstanceWrapper struct {
	Id         string   `json:"instance_id"`
	ExternalId string   `json:"external_id"`
	Instance   Instance `json:"vm,omitempty"`
}

func (c *Client) GetInstance(config GetInstanceConfig) (Instance, error) {
	body, err := c.Get("/api/v1/vm", map[string]string{"id": config.Id})
	if err != nil {
		return Instance{}, fmt.Errorf("failed getting instance %s: %w", config.Id, err)
	}

	var response getInstanceResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return Instance{}, fmt.Errorf("unexpected response body structure: %s", string(body))
	}

	return response.Instance, nil
}

func (c *Client) CreateInstance(config CreateInstanceConfig) (string, error) {
	payload := createInstanceRequestPayload{
		TemplateId: config.TemplateId,
	}

	if config.ExternalId != "" {
		payload.ExternalId = config.ExternalId
	}

	if config.TemplateTag != "" {
		payload.Tag = config.TemplateTag
	}

	if config.NodeId != "" {
		payload.NodeId = config.NodeId
	}

	if config.Priority != 0 {
		if config.Priority < 0 || config.Priority > 10000 {
			return "", fmt.Errorf("priority must be between 1 and 10000. Got %d", config.Priority)
		}
		payload.Priority = config.Priority
	}

	if config.NodeGroupId != "" {
		payload.NodeGroupId = config.NodeGroupId
	}

	body, err := c.Post("/api/v1/vm", payload)
	if err != nil {
		return "", err
	}

	var response createInstanceResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("unexpected response body structure: %s", string(body))
	}

	return response.InstanceIds[0], nil
}

func (c *Client) WaitForInstanceToBeScheduled(instanceId string) error {
	log.Printf("waiting for instance %s to be scheduled\n", instanceId)
	for {
		instance, err := c.GetInstance(GetInstanceConfig{Id: instanceId})
		if err != nil {
			return fmt.Errorf("failed getting instance status: %w", err)
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

type terminateInstanceRequestPayload struct {
	Id string `json:"id"`
}

type terminateInstanceResponse struct {
	response
}

type TerminateInstanceConfig struct {
	InstanceId string
}

func (c *Client) TerminateInstance(config TerminateInstanceConfig) error {
	payload := terminateInstanceRequestPayload{
		Id: config.InstanceId,
	}

	body, err := c.Delete("/api/v1/vm", payload)
	if err != nil {
		return err
	}

	var response terminateInstanceResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return fmt.Errorf("unexpected response body structure: %s", string(body))
	}

	return nil
}

func (c *Client) GetAllInstances() ([]InstanceWrapper, error) {

	body, err := c.Get("/api/v1/vm", nil)
	if err != nil {
		return nil, fmt.Errorf("failed sending request: %w", err)
	}

	var response getAllInstancesResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("unexpected response body structure: %s", string(body))
	}

	return response.Instances, nil
}

func (c *Client) GetInstanceByExternalId(externalId string) (InstanceWrapper, error) {
	instances, err := c.GetAllInstances()
	if err != nil {
		return InstanceWrapper{}, fmt.Errorf("failed getting all instances: %w", err)
	}

	for _, instance := range instances {
		if instance.ExternalId == externalId {
			return instance, nil
		}
	}

	return InstanceWrapper{}, fmt.Errorf("instance with external id %s not found", externalId)
}
