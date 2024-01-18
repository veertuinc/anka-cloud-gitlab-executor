package ankacloud

const (
	statusOK = "OK"
)

type response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Body    interface{} `json:"body,omitempty"`
}

type StartupScriptCondition int

const (
	WaitForNetwork StartupScriptCondition = 0
	NoWait         StartupScriptCondition = 1
)

type CreateInstanceRequest struct {
	TemplateId              string                 `json:"vmid"`
	ExternalId              string                 `json:"external_id,omitempty"`
	Tag                     string                 `json:"tag,omitempty"`
	NodeId                  string                 `json:"node_id,omitempty"`
	Priority                int                    `json:"priority,omitempty"`
	NodeGroupId             string                 `json:"group_id,omitempty"`
	StartupScript           string                 `json:"startup_script,omitempty"`
	StartupScriptMonitoring bool                   `json:"script_monitoring,omitempty"`
	StartupScriptTimeout    int                    `json:"script_timeout,omitempty"`
	StartupScriptCondition  StartupScriptCondition `json:"startup_script_condition"`
	Vcpu                    int                    `json:"vcpu,omitempty"`
	VramMb                  int                    `json:"vram,omitempty"`
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

type GetNodeRequest struct {
	Id string
}
type getNodeResponse struct {
	response
	Nodes []Node `json:"body"`
}

type getTemplatesResponse struct {
	response
	Templates []Template `json:"body"`
}
