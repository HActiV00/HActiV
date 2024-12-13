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

type ContainerInfo struct {
	ID   string
	Name string
}

func GetNamespaceInode(pid uint32) (uint64, error) {
	if int(pid) == 0 {
		pid = 1
	}
	nsPath := fmt.Sprintf("/proc/%d/ns/mnt", pid)
	stat, err := os.Stat(nsPath)
	if err != nil {
		return 0, err
	}

	stat_t, ok := stat.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, fmt.Errorf("failed to get Stat_t")
	}
	return stat_t.Ino, nil
}

func GetAllContainerNamespaces() (map[uint64]ContainerInfo, error) {
	containerNamespaces := make(map[uint64]ContainerInfo)

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", "/var/run/docker.sock")
			},
		},
	}

	resp, err := httpClient.Get("http://localhost/containers/json")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

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

		pid := inspect.State.Pid
		inode, err := GetNamespaceInode(uint32(pid))
		if err != nil {
			fmt.Printf("failed to get namespace inode for PID %d: %s\n", pid, err)
			continue
		}
		containerNamespaces[inode] = ContainerInfo{ID: container.ID, Name: strings.Replace(container.Names[0], "/", "", 1)}

	}
	if HostMonitoring {
		var hostInode uint64
		hostInode, _ = GetNamespaceInode(1)
		containerNamespaces[hostInode] = ContainerInfo{ID: "H", Name: "H"}
	}
	return containerNamespaces, err
}
