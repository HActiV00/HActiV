// Copyright Authors of HActiV

// tool package: 1.containermetrics.go
package tools

import (
	"HActiV/pkg/docker"
	"HActiV/pkg/utils"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
)

func GetHostDiskTotal() (uint64, error) {
	diskStat, err := disk.Usage("/")
	if err != nil {
		return 0, err
	}
	return diskStat.Total, nil
}

func GetHostCoreCount() (int, error) {
	return cpu.Counts(true)
}

func GetContainerDiskUsage(cli *client.Client, containerID string) (int64, error) {
	containerJSON, _, err := cli.ContainerInspectWithRaw(context.Background(), containerID, true)
	if err != nil {
		return 0, err
	}
	if containerJSON.SizeRw != nil {
		return *containerJSON.SizeRw, nil
	} else {
		return 0, nil
	}
}

func GetContainerNetwork(cli *client.Client, containerID string) (uint64, uint64, error) {
	stats, err := cli.ContainerStats(context.Background(), containerID, false)
	if err != nil {
		return 0, 0, err
	}
	defer stats.Body.Close()

	var statsJSON container.StatsResponse
	if err := json.NewDecoder(stats.Body).Decode(&statsJSON); err != nil {
		return 0, 0, err
	}

	var rx, tx uint64
	for _, v := range statsJSON.Networks {
		rx += v.RxBytes
		tx += v.TxBytes
	}
	return rx, tx, nil
}

func GetContainerCoreCount(cli *client.Client, containerID string) (float64, error) {
	containerJSON, err := cli.ContainerInspect(context.Background(), containerID)
	if err != nil {
		return 0, err
	}

	cpuQuota := containerJSON.HostConfig.CPUQuota
	cpuPeriod := containerJSON.HostConfig.CPUPeriod

	if cpuQuota > 0 && cpuPeriod > 0 {
		cpuCount := float64(cpuQuota) / float64(cpuPeriod)
		return cpuCount, nil
	} else {
		hostCoreCount, err := GetHostCoreCount()
		if err != nil {
			return 0, err
		}
		return float64(hostCoreCount), nil
	}
}

func DisplayMetrics(cli *client.Client) {
	hostDiskTotalBytes, err := GetHostDiskTotal()
	if err != nil {
		fmt.Printf("호스트 디스크 용량을 가져오는 중 오류 발생: %v\n", err)
		return
	}
	var keyToRemove uint64
	found := false
	containersMap := docker.GetContainer()
	for key, container := range containersMap {
		if container.Name == "H" && container.ID == "H" {
			keyToRemove = key
			found = true
			break
		}
	}

	if found {
		delete(containersMap, keyToRemove)
	}

	containers := make([]utils.ContainerInfo, 0, len(containersMap))
	for _, info := range containersMap {
		containers = append(containers, info)
	}

	sort.Slice(containers, func(i, j int) bool {
		return containers[i].Name < containers[j].Name
	})

	for _, info := range containers {
		stats, err := cli.ContainerStatsOneShot(context.Background(), info.ID)
		if err != nil {
			fmt.Printf("[컨테이너 %s: 상태 정보를 가져오는 중 오류 발생: %v]\n", info.Name, err)
			continue
		}

		data, err := io.ReadAll(stats.Body)
		stats.Body.Close()
		if err != nil {
			fmt.Printf("[컨테이너 %s: 상태 정보를 읽는 중 오류 발생: %v]\n", info.Name, err)
			continue
		}

		var stat container.StatsResponse
		if err := json.Unmarshal(data, &stat); err != nil {
			fmt.Printf("[컨테이너 %s: 상태 정보를 해석하는 중 오류 발생: %v]\n", info.Name, err)
			continue
		}

		cpuDelta := float64(stat.CPUStats.CPUUsage.TotalUsage - stat.PreCPUStats.CPUUsage.TotalUsage)
		systemDelta := float64(stat.CPUStats.SystemUsage - stat.PreCPUStats.SystemUsage)
		cpuUsage := 0.0
		if systemDelta > 0.0 && cpuDelta > 0.0 {
			cpuUsage = (cpuDelta / systemDelta) * 100.0
		}

		memUsage := float64(stat.MemoryStats.Usage) / (1024 * 1024)
		memUsagePercent := (float64(stat.MemoryStats.Usage) / float64(stat.MemoryStats.Limit)) * 100

		diskUsageBytes, err := GetContainerDiskUsage(cli, info.ID)
		if err != nil {
			fmt.Printf("[컨테이너 %s: 디스크 사용량을 가져오는 중 오류 발생: %v]\n", info.Name, err)
			continue
		}

		diskUsageMB := float64(diskUsageBytes) / (1024 * 1024)
		diskUsagePercent := (float64(diskUsageBytes) / float64(hostDiskTotalBytes)) * 100
		containerCoreCount, err := GetContainerCoreCount(cli, info.ID)

		if err != nil {
			fmt.Printf("[컨테이너 %s: 코어 수를 가져오는 중 오류 발생: %v]\n", info.Name, err)
			continue
		}
		rxBytes, txBytes, err := GetContainerNetwork(cli, info.ID)
		if err != nil {
			fmt.Printf("[컨테이너 %s: 네트워크 사용량을 가져오는 중 오류 발생: %v]\n", info.Name, err)
			continue
		}

		fmt.Printf("컨테이너 이름: %s\n", info.Name)
		fmt.Printf("CPU 사용량: %.2f%%\n", cpuUsage)
		fmt.Printf("코어 수: %.2f\n", containerCoreCount)
		fmt.Printf("메모리 사용량: %.2f%% (%.2f MB)\n", memUsagePercent, memUsage)
		fmt.Printf("디스크 사용량: %.6f%% (%.2f MB / %.2f GB)\n", diskUsagePercent, diskUsageMB, float64(hostDiskTotalBytes)/(1024*1024*1024))
		fmt.Printf("네트워크 사용량: RX %.2f MB, TX %.2f MB\n", float64(rxBytes)/(1024*1024), float64(txBytes)/(1024*1024))
		fmt.Println("-----------------------------")

		utils.DataSend(
			"ContainerMetrics",
			time.Now().Format(time.RFC3339),
			info.Name,
			cpuUsage,
			memUsage,
			0.0,
			rxBytes,
			txBytes,
		)
	}
}

func ContainerMetrics() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Printf("Docker 클라이언트 생성 중 오류 발생: %v\n", err)
		return
	}
	docker.SetContainer()

	DisplayMetrics(cli)
}
