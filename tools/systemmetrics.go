package tools

import (
    "sync"
)

func SystemMetrics() {
    var wg sync.WaitGroup
    wg.Add(2)

    go func() {
        defer wg.Done()
        HostMetrics()
    }()

    go func() {
        defer wg.Done()
        ContainerMetrics()
    }()

    wg.Wait()
}

