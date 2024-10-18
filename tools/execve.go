package tools

import (
	"HActiV/pkg/docker"
	"HActiV/pkg/utils"
	bpfcode "HActiV/tools/bpf"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	bpf "github.com/iovisor/gobpf/bcc"
)

type execvevent struct {
	UID      uint32
	GID      uint32
	PID      uint32
	PPID     uint32
	Comm     [16]byte
	Filename [200]byte
	Args     [200]byte
}

func ExecveMonitoring() {
	bpfModule := utils.LoadBPFModule(bpfcode.ExecveCcode)
	defer bpfModule.Close()

	tracepoint, err := bpfModule.LoadTracepoint("tracepoint__syscalls__sys_enter_execve")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load tracepoint: %s\n", err)
		os.Exit(1)
	}

	err = bpfModule.AttachTracepoint("syscalls:sys_enter_execve", tracepoint)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to attach tracepoint: %s\n", err)
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
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		var event execvevent
		for {
			data := <-channel
			err := binary.Read(bytes.NewBuffer(data), binary.LittleEndian, &event)
			if err != nil {
				fmt.Printf("failed to decode received data: %s\n", err)
				continue
			}
			inode, err := utils.GetNamespaceInode(event.PPID)
			if err != nil {
				fmt.Printf("failed to get namespace for PID %d: %s\n", event.PID, err)
				continue
			}
			containerNamespaces := docker.GetContainer()

			containerInfo, exists := containerNamespaces[inode]
			if !exists {
				continue
			}

			fmt.Println(containerNamespaces)

			utils.DataSend("Systemcall", time.Now().Format(time.RFC3339), containerInfo.Name, event.UID, event.GID, event.PID, event.PPID,
				string(bytes.TrimRight(event.Filename[:], "\x00")),
				string(bytes.TrimRight(event.Comm[:], "\x00")),
				strings.Replace(convertByteArrayToString(event.Args), "--color=auto", "", 1), 0)

			fmt.Printf("Container Name: %s, UID: %d, GID:%d, PID: %d, PPID: %d, Comm: %s, Filename: %s, Args: %s\n",
				containerInfo.Name, event.UID, event.GID, event.PID, event.PPID,
				string(bytes.TrimRight(event.Comm[:], "\x00")),
				string(bytes.TrimRight(event.Filename[:], "\x00")),
				strings.Replace(convertByteArrayToString(event.Args), "--color=auto", "", -1)) // 인자 출력 추가
		}
	}()

	perfMap.Start()
	fmt.Println("Execve Tracepoint Monitoring")
	<-sig
	perfMap.Stop()
}

func convertByteArrayToString(arr [200]byte) string {
	n := bytes.IndexByte(arr[:], 0)
	if n == -1 {
		n = len(arr)
	}
	return string(arr[:n])
}
