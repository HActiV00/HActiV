package tools

import (
	config "HActiV/configs"
	"HActiV/pkg/docker"
	"HActiV/pkg/utils"
	bpfcode "HActiV/tools/bpf"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	bpf "github.com/iovisor/gobpf/bcc"
)

type openEvent struct {
	Pid         uint32
	PPid        uint32
	Uid         uint32
	Gid         uint32
	ReturnValue int32
	Comm        [16]byte
	Filename    [256]byte
}

func OpenMonitoring() {
	cfg, err := config.LoadConfig("../configs/openconfig.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %s\n", err)
		os.Exit(1)
	}

	bpfModule := utils.LoadBPFModule(bpfcode.OpenCcode)
	defer bpfModule.Close()
	syscall := "syscalls:sys_enter_openat"

	fmt.Printf("Attaching eBPF program to tracepoint: %s\n", syscall)

	err = AttachEBPFProgram(bpfModule, "trace_sys_enter_openat", syscall)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to attach eBPF program to %s: %s\n", syscall, err)
		os.Exit(1)
	}

	table := bpf.NewTable(bpfModule.TableId("events"), bpfModule)
	channel := make(chan []byte)
	perfMap, err := bpf.InitPerfMap(table, channel, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init perf map: %s\n", err)
		os.Exit(1)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get container namespaces: %s\n", err)
		os.Exit(1)
	}
	go func() {
		var event openEvent
		for {
			data := <-channel
			err := binary.Read(bytes.NewBuffer(data), binary.LittleEndian, &event)
			if err != nil {
				fmt.Printf("failed to decode received data: %s\n", err)
				continue
			}

			filename := string(event.Filename[:bytes.IndexByte(event.Filename[:], 0)])
			if !ShouldLog(filename, cfg.FilterPatterns) {
				continue
			}

			inode, err := utils.GetNamespaceInode(event.PPid)
			if err != nil {
				fmt.Printf("failed to get namespace for PID %d: %s\n", event.Pid, err)
				continue
			}
			containerNamespaces := docker.GetContainer()
			containerInfo, exists := containerNamespaces[inode]
			if !exists {
				continue
			}
			utils.DataSend("File_open", time.Now().Format(time.RFC3339), containerInfo.Name, event.Uid, event.Gid, event.Pid, event.PPid,
				string(bytes.Trim(event.Comm[:], "\x00")), filename, event.ReturnValue)

			fmt.Printf("%s | Container: %s (ID: %s) | PPID : %d | PID : %d | GID : %d | UID : %d | FILE : %s | command : %s (return %d)\n",
				time.Now().Format(time.RFC3339),
				containerInfo.Name, containerInfo.ID,
				event.PPid,
				event.Pid,
				event.Gid,
				event.Uid,
				filename,
				bytes.Trim(event.Comm[:], "\x00"),
				event.ReturnValue)
		}
	}()

	perfMap.Start()
	<-sig
	perfMap.Stop()
}

func AttachEBPFProgram(bpfModule *bpf.Module, functionName string, syscall string) error {
	tracepoint, err := bpfModule.LoadTracepoint(functionName)
	if err != nil {
		return fmt.Errorf("failed to load tracepoint: %v", err)
	}

	err = bpfModule.AttachTracepoint(syscall, tracepoint)
	if err != nil {
		return fmt.Errorf("failed to attach tracepoint: %v", err)
	}

	return nil
}

func ShouldLog(filename string, filterPatterns []string) bool {
	for _, pattern := range filterPatterns {
		if strings.Contains(filename, pattern) {
			return false
		}
	}
	return true
}
