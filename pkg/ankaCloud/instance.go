package ankaCloud

import (
	"fmt"
	"time"
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
	TemplateId string `json:"vmid"`
	ExternalId string `json:"external_id,omitempty"`
}

type createInstanceResponse struct {
	response
	InstanceIds []string `json:"body"`
}

type CreateInstanceConfig struct {
	TemplateId         string
	ExternalId         string
	WaitUntilScheduled bool
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
	var response getInstanceResponse
	if err := c.Get("/api/v1/vm", map[string]string{"id": config.Id}, &response); err != nil {
		return Instance{}, fmt.Errorf("failed sending request: %w", err)
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

	var response createInstanceResponse
	err := c.Post("/api/v1/vm", payload, &response)
	if err != nil {
		return "", err
	}

	instanceId := response.InstanceIds[0]

	// TODO: add timeout for scheduling, and timeout for pulling
	// TODO: move this to a separate function
	// TODO: mske sleep between retries expoonential to a limit
	if config.WaitUntilScheduled {
		getInstaneConfig := GetInstanceConfig{Id: instanceId}
		for {
			instance, err := c.GetInstance(getInstaneConfig)
			if err != nil {
				return "", fmt.Errorf("failed getting instance status: %w", err)
			}

			fmt.Printf("instance in %s state\n", instance.State) // TODO: move to debug log
			switch instance.State {
			case StateScheduling, StatePulling:
				time.Sleep(5 * time.Second)
				continue
			}

			if instance.State == StateStarted {
				break
			}

			return "", fmt.Errorf("instance is in an unexpected state: %s", instance.State)
		}
	}

	return instanceId, nil
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

	var response terminateInstanceResponse
	err := c.Delete("/api/v1/vm", payload, &response)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetAllInstances() ([]InstanceWrapper, error) {
	var response getAllInstancesResponse
	if err := c.Get("/api/v1/vm", nil, &response); err != nil {
		return nil, fmt.Errorf("failed sending request: %w", err)
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
