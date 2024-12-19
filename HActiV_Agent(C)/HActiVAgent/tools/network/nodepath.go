// Copyright Authors of HActiV

// network package
package network

import (
	"encoding/json"
)

type PathNode struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type PathLink struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type TrafficPath struct {
	Nodes []PathNode `json:"nodes"`
	Links []PathLink `json:"links"`
}

func generateTrafficPath(srcIP, srcType, dstIP, dstType string) TrafficPath {
	path := TrafficPath{}
	path.Nodes = append(path.Nodes, PathNode{ID: srcIP, Type: srcType})

	if srcType == "Docker internal" && dstType == "External" {
		path.Nodes = append(path.Nodes,
			PathNode{ID: hostIP, Type: "Host internal"},
			PathNode{ID: gatewayIP, Type: "Gateway"},
			PathNode{ID: dnsServers[0], Type: "DNS"})
	}

	path.Nodes = append(path.Nodes, PathNode{ID: dstIP, Type: dstType})

	for i := 0; i < len(path.Nodes)-1; i++ {
		path.Links = append(path.Links, PathLink{
			Source: path.Nodes[i].ID,
			Target: path.Nodes[i+1].ID,
		})
	}
	return path
}

func GenerateTrafficPathJSON(srcIP, srcType, dstIP, dstType string) (string, error) {
	path := generateTrafficPath(srcIP, srcType, dstIP, dstType)
	pathJSON, err := json.Marshal(path)
	if err != nil {
		return "", err
	}
	return string(pathJSON), nil
}
