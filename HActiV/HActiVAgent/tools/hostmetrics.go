// Copyright Authors of HActiV

// tool package: 1.hostmetrics.go
package tools

import (
	"HActiV/pkg/utils"
	"fmt"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
)

func HostMetrics() {
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		fmt.Println("CPU 사용량을 가져오는 중 오류 발생:", err)
		return
	}

	cpuCores, err := cpu.Counts(true)
	if err != nil {
		fmt.Println("코어 수를 가져오는 중 오류 발생:", err)
		return
	}

	vmStat, err := mem.VirtualMemory()
	if err != nil {
		fmt.Println("메모리 사용량을 가져오는 중 오류 발생:", err)
		return
	}

	diskStat, err := disk.Usage("/")
	if err != nil {
		fmt.Println("디스크 사용량을 가져오는 중 오류 발생:", err)
		return
	}

	fmt.Println("호스트 시스템 매트릭스:")
	fmt.Printf("CPU 사용량: %.2f%%\n", cpuPercent[0])
	fmt.Printf("코어 수: %d\n", cpuCores)
	fmt.Printf("메모리 사용량: %.2f%%\n", vmStat.UsedPercent)
	fmt.Printf("디스크 사용량: %.2f%%\n", diskStat.UsedPercent)
	fmt.Println("=============================")

	utils.DataSend(
		"HostMetrics",
		time.Now().Format(time.RFC3339),
		cpuPercent[0],
		vmStat.UsedPercent,
		diskStat.UsedPercent,
		cpuCores,
	)
}
