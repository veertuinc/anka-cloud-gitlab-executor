package ankacloud

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/veertuinc/anka-cloud-gitlab-executor/internal/log"
)

type controller struct {
	APIClient *APIClient
}

type Node struct {
	Id   string `json:"node_id"`
	Name string `json:"node_name"`
	IP   string `json:"ip_address"`
}

type Template struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Size int64  `json:"size"`
	Arch string `json:"arch"`
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
	Name                string               `json:"name"`
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
	VMInfo     *VM           `json:"vminfo,omitempty"`
	NodeId     string        `json:"node_id"`
	Node       *Node         `json:"node,omitempty"`
	Progress   float32       `json:"progress,omitempty"`
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

func (c *controller) WaitForInstanceToBeScheduled(ctx context.Context, instanceId string) (*Instance, error) {
	const pollingInterval = 10 * time.Second
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(pollingInterval):
			instance, err := c.GetInstance(ctx, GetInstanceRequest{Id: instanceId})
			if err != nil {
				return nil, fmt.Errorf("failed to get instance %q status: %w", instanceId, err)
			}
			log.ConditionalColorf("instance %s is in state %q\n", instanceId, instance.State)
			switch instance.State {
			case StateScheduling:
				break
			case StatePulling:
				if instance.Progress != 0 {
					log.ConditionalColorf("pulling progress: %.0f%%\n", instance.Progress*100)
				}
			case StateStarted:
				// get the rest of the node details
				node, err := c.GetNode(ctx, GetNodeRequest{Id: instance.NodeId})
				if err != nil {
					return nil, fmt.Errorf("failed to get node %s: %w", instance.Node.Id, err)
				}
				instance.Node = node
				return instance, nil
			default:
				return nil, fmt.Errorf("instance %s is in an unexpected state: %s", instanceId, instance.State)
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

func (c *controller) GetAllInstances(ctx context.Context) ([]Instance, error) {

	body, err := c.APIClient.Get(ctx, "/api/v1/vm", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get all instances: %w", err)
	}

	var response getAllInstancesResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body %q: %w", string(body), err)
	}

	var instances []Instance
	for _, instanceWrapper := range response.Instances {
		instances = append(instances, *instanceWrapper.Instance)
	}

	return instances, nil
}

func (c *controller) GetInstanceByExternalId(ctx context.Context, externalId string) (*Instance, error) {
	instances, err := c.GetAllInstances(ctx)
	if len(instances) == 0 {
		return nil, fmt.Errorf("no instances returned from controller")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get instance by external id %s: %w", externalId, err)
	}

	for _, instance := range instances {
		if instance.ExternalId == externalId {
			return &instance, nil
		}
	}

	return nil, fmt.Errorf("instance with external id %s not found", externalId)
}

func (c *controller) GetTemplateIdByName(ctx context.Context, templateName string) (string, error) {
	body, err := c.APIClient.Get(ctx, "/api/v1/registry/vm", map[string]string{"apiVer": "v1"})
	if err != nil {
		return "", fmt.Errorf("failed to get templates: %w", err)
	}

	var response getTemplatesResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("failed to parse response body %q: %w", string(body), err)
	}

	for _, t := range response.Templates {
		if t.Name == templateName {
			return t.Id, nil
		}
	}

	return "", fmt.Errorf("template %q not found", templateName)
}
