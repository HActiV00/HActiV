package tools

import (
    "fmt"
    "github.com/shirou/gopsutil/cpu"
    "github.com/shirou/gopsutil/disk"
    "github.com/shirou/gopsutil/mem"
    "time"
)

func HostMetrics() {
    // CPU 사용량
    cpuPercent, err := cpu.Percent(time.Second, false)
    if err != nil {
        fmt.Println("CPU 사용량을 가져오는 중 오류 발생:", err)
        return
    }

    // 코어 수
    cpuCores, err := cpu.Counts(true)
    if err != nil {
        fmt.Println("코어 수를 가져오는 중 오류 발생:", err)
        return
    }

    // 메모리 사용량
    vmStat, err := mem.VirtualMemory()
    if err != nil {
        fmt.Println("메모리 사용량을 가져오는 중 오류 발생:", err)
        return
    }

    // 디스크 사용량
    diskStat, err := disk.Usage("/")
    if err != nil {
        fmt.Println("디스크 사용량을 가져오는 중 오류 발생:", err)
        return
    }

    // 결과 출력
    fmt.Println("호스트 시스템 매트릭스:")
    fmt.Printf("CPU 사용량: %.2f%%\n", cpuPercent[0])
    fmt.Printf("코어 수: %d\n", cpuCores)
    fmt.Printf("메모리 사용량: %.2f%%\n", vmStat.UsedPercent)
    fmt.Printf("디스크 사용량: %.2f%%\n", diskStat.UsedPercent)
    fmt.Println("=============================")
}

