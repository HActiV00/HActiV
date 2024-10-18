package docker

/*
import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
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

func Docker() {
	// Docker 클라이언트 생성
	_, err := client.NewClientWithOpts(client.WithHost("unix:///var/run/docker.sock"))
	if err != nil {
		panic(err)
	}

	// HTTP 클라이언트 설정
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", "/var/run/docker.sock")
			},
		},
	}

	// API 요청
	resp, err := httpClient.Get("http://localhost/containers/json")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// JSON 응답 파싱
	var containers []types.Container
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		panic(err)
	}

	// 컨테이너 정보 출력
	for _, container := range containers {
		fmt.Printf("컨테이너 ID: %s\n", container.ID)
		fmt.Printf("이름: %s\n", strings.Replace(container.Names[0], "/", "", 1))
		fmt.Printf("이미지: %s\n", container.Image)
		fmt.Printf("상태: %s\n", container.State)
		fmt.Printf("생성 시간: %d\n", container.Created)
		fmt.Println("---")

		pidCmd := exec.Command("docker", "inspect", "--format", "{{.State.Pid}}", container.ID)
		pidOutput, err := pidCmd.Output()
		if err != nil {
			fmt.Printf("failed to get PID for container %s: %s\n", container.ID, err)
			continue
		}

		pidStr := strings.TrimSpace(string(pidOutput))
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			fmt.Printf("failed to parse PID for container %s: %s\n", container.ID, err)
			continue
		}

		inode, err := GetNamespaceInode(uint32(pid))
		if err != nil {
			fmt.Printf("failed to get namespace inode for PID %d: %s\n", pid, err)
			continue
		}
		containerNamespaces[inode] = ContainerInfo{ID: container.ID, Name: strings.Replace(container.Names[0], "/", "", 1)}
	}
}
*/
