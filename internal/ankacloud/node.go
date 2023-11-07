package ankacloud

import (
	"context"
	"encoding/json"
	"fmt"
)

type Node struct {
	IP string `json:"ip_address"`
}

type GetNodeRequest struct {
	Id string
}
type getNodeResponse struct {
	response
	Nodes []Node `json:"body"`
}

func (c *Client) GetNode(ctx context.Context, req GetNodeRequest) (*Node, error) {
	body, err := c.Get(ctx, "/api/v1/node", map[string]string{"id": req.Id})
	if err != nil {
		return nil, fmt.Errorf("failed getting node %q: %w", req.Id, err)
	}

	var response getNodeResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed parsing response body %q: %w", string(body), err)
	}

	if len(response.Nodes) == 0 {
		return nil, fmt.Errorf("node %s not found", req.Id)
	}

	return &response.Nodes[0], nil
}
