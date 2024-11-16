//Copyright Authors of HActiV

// docker package for docker information
package docker

import (
	"HActiV/pkg/utils"
	"fmt"
	"os"
	"sync"
)

// shared variable
type SafeContainer struct {
	mu    sync.RWMutex
	value map[uint64]utils.ContainerInfo
}

var container SafeContainer // 전역 변수로 선언

// Set shared variable
func SetContainer() {
	container.mu.Lock()
	defer container.mu.Unlock()
	containerNamespaces, err := utils.GetAllContainerNamespaces()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get container namespaces: %s\n", err)
		os.Exit(1)
	}
	container.value = containerNamespaces
}

// Get shared variable
func GetContainer() map[uint64]utils.ContainerInfo {
	return container.value
}
