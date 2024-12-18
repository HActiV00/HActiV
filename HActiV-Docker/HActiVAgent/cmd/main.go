// Copyright Authors of HActiV
package main

import (
	"HActiV/configs"
	"HActiV/pkg/docker"
	"HActiV/pkg/utils"
	"HActiV/tools"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	if os.Geteuid() != 0 {
		fmt.Println("This program must be run as root!")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Println("------------------------------")
		fmt.Println("Usage: sudo ./HActiV <function_number(s)>")
		fmt.Println("Example: ./HActiV 0 1 2 3 4 5 6")
		fmt.Println("[option 1: system metrics monitoring]")
		fmt.Println("[option 2: execve event monitoring]")
		fmt.Println("[option 3: delete event monitoring]")
		fmt.Println("[option 4: memory event monitoring]")
		fmt.Println("[option 5: network event monitoring]")
		fmt.Println("[option 6: open event monitoring]")
		fmt.Println("------------------------------")
		return
	}

	configs.HActiVSetting()
	utils.DataSendSetting()
	configs.FirstRules()

	fmt.Println("초기 규칙 파일 설정이 완료되었습니다.")

	functions := map[string]func(){
		"1": func() {
			for {
				tools.SystemMetrics()
				time.Sleep(10 * time.Second)
			}
		},
		"2": tools.ExecveMonitoring,
		"3": tools.DeleteMonitoring,
		"4": tools.MemoryMonitoring,
		"5": tools.NetworkMonitoring,
		"6": tools.OpenMonitoring,
	}

	var wg sync.WaitGroup
	docker.SetContainer()

	wg.Add(1)
	go func() {
		defer wg.Done()
		docker.MonitorDockerEvents()
	}()

	for _, arg := range os.Args[1:] {
		if fn, exists := functions[arg]; exists {
			wg.Add(1)
			go func(f func()) {
				defer wg.Done()
				f()
			}(fn)
		} else {
			fmt.Printf("Unknown function: %s\n", arg)
			os.Exit(0)
		}
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt signal. Shutting down...")
		os.Exit(0)
	}()

	wg.Wait()
	fmt.Println("All functions completed")
}
