// Copyright Authors of HActiV

// utils package for helping other package
package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"syscall"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// Basic container struct
type ContainerInfo struct {
	ID   string
	Name string
}

func getInode(filePath string) (uint64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, fmt.Errorf("failed to get Stat_t")
	}
	return stat.Ino, nil
}

// Get All Container Name and ID
func GetAllContainerNamespaces() (map[uint64]ContainerInfo, error) {
	containerNamespaces := make(map[uint64]ContainerInfo)

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// HTTP Client for docker
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", "/var/run/docker.sock")
			},
		},
	}

	// Request API
	resp, err := httpClient.Get("http://localhost/containers/json")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// JSON Parsing
	var containers []types.Container
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		panic(err)
	}

	for _, container := range containers {
		inspect, err := cli.ContainerInspect(context.Background(), container.ID)
		if err != nil {
			fmt.Printf("failed to inspect container %s: %s\n", container.ID, err)
			continue
		}

		sandboxKey := inspect.NetworkSettings.SandboxKey
		if sandboxKey == "" {
			fmt.Printf("Container %s has no network namespace\n", container.ID)
			continue
		}

		inode, err := getInode(sandboxKey)
		if err != nil {
			fmt.Printf("Failed to get inode for %s: %s\n", sandboxKey, err)
			continue
		}

		containerNamespaces[inode] = ContainerInfo{ID: container.ID, Name: strings.Replace(container.Names[0], "/", "", 1)}

	}

	return containerNamespaces, err
}
