package main

import (
	"HActiV/pkg/docker"
	"HActiV/tools"
	"fmt"
	"os"
	"sync"
)

func main() {

	if os.Geteuid() != 0 {
		fmt.Println("This program must be run as root!")
		os.Exit(1)
	}
	if len(os.Args) < 2 {
		fmt.Println("Usage: sudo go run main.go <function_number(s)>")
		fmt.Println("Example: go run main.go 1 2 3 4 5 6")
		return
	}

	functions := map[string]func(){
		"1": tools.ExecveMonitoring,
		"2": tools.LogFileAccessMonitoring,
		"3": tools.LogFileDeleteMonitoring,
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

	wg.Wait()
	fmt.Println("All functions completed")

}
