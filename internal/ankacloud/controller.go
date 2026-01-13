package ankacloud

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/veertuinc/anka-cloud-gitlab-executor/internal/log"
)

type Controller struct {
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

func NewController(apiClient *APIClient) *Controller {
	return &Controller{
		APIClient: apiClient,
	}
}
func (c *Controller) GetNode(ctx context.Context, req GetNodeRequest) (*Node, error) {
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

func (c *Controller) GetInstance(ctx context.Context, req GetInstanceRequest) (*Instance, error) {
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

func (c *Controller) CreateInstance(ctx context.Context, payload CreateInstanceRequest) (string, error) {

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

func (c *Controller) WaitForInstanceToBeScheduled(ctx context.Context, instanceId string) (*Instance, error) {
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
					return nil, fmt.Errorf("failed to get node %s: %w", instance.NodeId, err)
				}
				instance.Node = node
				return instance, nil
			default:
				return nil, fmt.Errorf("instance %s is in an unexpected state: %s", instanceId, instance.State)
			}
		}
	}
}

func (c *Controller) TerminateInstance(ctx context.Context, payload TerminateInstanceRequest) error {
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

// TerminateInstanceWithRetry terminates an instance with automatic retry on timeout errors
func (c *Controller) TerminateInstanceWithRetry(ctx context.Context, payload TerminateInstanceRequest) error {
	return WithRetryNoResult(ctx, DefaultRetryConfig(), func() error {
		return c.TerminateInstance(ctx, payload)
	})
}

func (c *Controller) GetAllInstances(ctx context.Context) ([]Instance, error) {

	body, err := c.APIClient.Get(ctx, "/api/v1/vm", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get all instances: %w", err)
	}

	var response getAllInstancesResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body %q: %w", string(body), err)
	}

	log.Debugf("got %d instances back from controller\n", len(response.Instances))
	body, err = json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response body %q: %w", string(body), err)
	}
	log.Debugf("instances response: %s\n", string(body))

	var instances []Instance
	for _, instanceWrapper := range response.Instances {
		instances = append(instances, *instanceWrapper.Instance)
	}

	return instances, nil
}

func (c *Controller) GetInstanceByExternalId(ctx context.Context, externalId string) (*Instance, error) {
	instances, err := c.GetAllInstances(ctx)

	if len(instances) == 0 {
		return nil, fmt.Errorf("no instances returned from controller: %w", err)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get instance by external id %s: %w", externalId, err)
	}

	var matchingInstances []*Instance
	for _, instance := range instances {
		if instance.ExternalId == externalId {
			matchingInstances = append(matchingInstances, &instance)
		}
	}

	if len(matchingInstances) == 0 {
		return nil, fmt.Errorf("instance with external id %s not found", externalId)
	}

	// If multiple instances with the same external ID exist, prioritize by state
	// Return the first instance that is in a good state (Started, Scheduling, Pulling)
	for _, instance := range matchingInstances {
		switch instance.State {
		case StateStarted, StateScheduling, StatePulling:
			return instance, nil
		}
	}

	// No instances in a usable state - fail explicitly instead of returning Error/Terminated instances
	return nil, fmt.Errorf("instance with external id %s exists but is not in a usable state (found state: %s)",
		externalId, matchingInstances[0].State)
}

func (c *Controller) GetTemplateIdByName(ctx context.Context, templateName string) (string, error) {
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
