// Copyright Authors of HActiV

// tool package: 5.network.go
package tools

import (
	"HActiV/configs"
	"HActiV/pkg/utils"
	bpfcode "HActiV/tools/bpf"
	"HActiV/tools/network"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"unsafe"
)

func NetworkMonitoring() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go startHTTPMonitoring()
	go monitorTraffic(ctx)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown
	cancel()
	network.SaveIPInfoToFile()
}

func startHTTPMonitoring() {
	fmt.Println("Starting HTTP monitoring...")

	activeInterfaces, err := network.GetActiveNetworkInterfaces()
	if err != nil {
		log.Fatalf("활성화된 네트워크 인터페이스를 감지하지 못했습니다: %v", err)
	}

	bpfFilter := "tcp port 80"
	for _, iface := range activeInterfaces {
		log.Printf("감지된 네트워크 인터페이스: %s", iface.Name)

		network.StartHTTPMonitor(iface.Name, bpfFilter)
	}
}

func monitorTraffic(ctx context.Context) {
	policies, err := configs.LoadRules("network")
	if err != nil {
		fmt.Fprintf(os.Stderr, "정책 로드 실패: %v\n", err)
		os.Exit(1)
	}
	bpfModule := utils.LoadBPFModule(bpfcode.NetworkCcode)
	defer bpfModule.Close()

	kprobercv, err := bpfModule.LoadKprobe("kprobe__ip_rcv")
	if err != nil {
		log.Fatalf("Failed to load kprobe__ip_rcv: %v", err)
	}
	err = bpfModule.AttachKprobe("ip_rcv", kprobercv, -1)
	if err != nil {
		log.Fatalf("Failed to attach kprobe__ip_rcv: %v", err)
	}

	kprobeout, err := bpfModule.LoadKprobe("kprobe__ip_output")
	if err != nil {
		log.Fatalf("Failed to load kprobe__ip_output: %v", err)
	}
	err = bpfModule.AttachKprobe("ip_output", kprobeout, -1)
	if err != nil {
		log.Fatalf("Failed to attach kprobe__ip_output: %v", err)
	}

	channel := make(chan []byte)
	perfMap := network.InitPerfMap(bpfModule, channel)
	perfMap.Start()

	logger, err := utils.NewDualLogger("network_compress", "networkjson")
	if err != nil {
		fmt.Println("로그 생성 실패:", err)
		return
	}
	defer logger.Close()

	fmt.Println("[Network event monitoring start...]")
	defer perfMap.Stop()
	for {
		select {
		case data := <-channel:
			network.HandleEvent(unsafe.Pointer(&data[0]), policies, logger)
		case <-ctx.Done():
			log.Println("Stopping traffic monitoring")
			return
		}
	}
}
