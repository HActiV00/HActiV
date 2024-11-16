package main

import (
	"HActiV/configs" // 설정 초기화 및 출력용 패키지
	"HActiV/pkg/docker"
	"HActiV/tools"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	// 루트 권한 확인
	if os.Geteuid() != 0 {
		fmt.Println("This program must be run as root!")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: sudo go run main.go <function_number(s)>")
		fmt.Println("Example: go run main.go 0 1 2 3 4 5 6")
		return
	}

	// 설정 초기화
	configs.InitSettings()

	// 초기 규칙 파일 설정
	if err := configs.InitRules(); err != nil {
		fmt.Printf("규칙 파일 초기화 실패: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("초기 규칙 파일 설정이 완료되었습니다.")

	functions := map[string]func(){
		"1": func() {
			for {
				tools.SystemMetrics()
				time.Sleep(10 * time.Second) // 10초마다 실행
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

	// docker container check 항상 실행
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
		}
	}

	// Ctrl+C 처리
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
