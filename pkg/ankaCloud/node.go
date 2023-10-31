package ankaCloud

import "fmt"

type GetNodeConfig struct {
	Id string
}

type Node struct {
	IP string `json:"ip_address"`
}

type getNodeResponse struct {
	response
	Nodes []Node `json:"body"`
}

func (c *Client) GetNode(config GetNodeConfig) (Node, error) {
	var response getNodeResponse
	if err := c.Get("/api/v1/node", map[string]string{"id": config.Id}, &response); err != nil {
		return Node{}, fmt.Errorf("failed sending request: %w", err)
	}

	if len(response.Nodes) == 0 {
		return Node{}, fmt.Errorf("node %s not found", config.Id)
	}

	return response.Nodes[0], nil
}
