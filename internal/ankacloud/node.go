package ankacloud

import (
	"encoding/json"
	"fmt"
)

type Node struct {
	IP string `json:"ip_address"`
}

type getNodeResponse struct {
	response
	Nodes []Node `json:"body"`
}

func (c *Client) GetNode(id string) (*Node, error) {
	body, err := c.Get("/api/v1/node", map[string]string{"id": id})
	if err != nil {
		return nil, fmt.Errorf("failed getting node %q: %w", id, err)
	}

	var response getNodeResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed parsing response body %q: %w", string(body), err)
	}

	if len(response.Nodes) == 0 {
		return nil, fmt.Errorf("node %s not found", id)
	}

	return &response.Nodes[0], nil
}
