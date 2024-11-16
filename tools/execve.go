package tools

import (
	"HActiV/configs"
	"HActiV/pkg/docker"
	"HActiV/pkg/utils"
	bpfcode "HActiV/tools/bpf"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"strings"
	"syscall"
	"time"

	bpf "github.com/iovisor/gobpf/bcc"
)

type execvevent struct {
	Uid           uint32
	Gid           uint32
	Pid           uint32
	Ppid          uint32
	Puid          uint32
	Pgid          uint32
	Comm          [16]byte
	Filename      [100]byte // FILENAME_SIZE에 맞게 조정
	Args          [200]byte // MAX_ARGS_SIZE에 맞게 조정
	NamespaceInum uint32
}

func ExecveMonitoring() {
	err := configs.SetupRules("execve")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load rules: %s \n", err)
		os.Exit(1)
	}

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
	lost := make(chan uint64)
	perfMap, err := bpf.InitPerfMap(table, channel, lost)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init perf map: %s\n", err)
		os.Exit(1)
	}
	var Count uint64
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		var event execvevent
		for {
			select {
			case data := <-channel:
				err := binary.Read(bytes.NewBuffer(data), binary.LittleEndian, &event)
				if err != nil {
					fmt.Printf("failed to decode received data: %s\n", err)
					continue
				}

				containerNamespaces := docker.GetContainer()
				containerInfo, exists := containerNamespaces[uint64(event.NamespaceInum)]
				if !exists {
					continue
				}

				processName := string(bytes.Trim(event.Comm[:], "\x00"))
				args := convertByteArrayToString(event.Args)
				filename := string(bytes.TrimRight(event.Filename[:], "\x00"))

				if processName == "go" || (strings.Contains(filename, "ps") && processName == "sh") || processName == "cpuUsage.sh" || processName == "node" || strings.Contains(filename, ".vscode-server") || strings.Contains(filename, "sleep") || strings.Contains(filename, "/usr/bin/which") {
					continue
				}

				eventData := map[string]interface{}{
					"process_name": processName,
					"uid":          event.Uid,
					"gid":          event.Gid,
				}

				if !configs.ParseRules("Process_execution", eventData, "execve") {
					continue
				}

				u, err := user.LookupId(fmt.Sprintf("%d", event.Uid))
				if err != nil {
					continue
				}

				utils.DataSend("Systemcall", time.Now().Format(time.RFC3339), containerInfo.Name, event.Uid, event.Gid, event.Pid, event.Ppid,
					string(bytes.TrimRight(event.Filename[:], "\x00")),
					string(bytes.TrimRight(event.Comm[:], "\x00")),
					strings.Replace(args, "--color=auto", "", 1), 0)

				fmt.Printf("Container Name: %s, UID: %s, GID:%d, PID: %d, PPID: %d, Comm: %s, Filename: %s, Args: %s\n",
					containerInfo.Name, u.Username, event.Gid, event.Pid, event.Ppid,
					processName,
					filename,
					strings.Replace(args, "--color=auto", "", -1))
			case lostc := <-lost:
				Count += lostc
			}
		}
	}()

	perfMap.Start()
	fmt.Println("Execve Tracepoint Monitoring")
	<-sig
	fmt.Println("Lost Count:", Count)
	perfMap.Stop()
}

func convertByteArrayToString(arr [200]byte) string { // Args 크기에 맞게 수정
	n := bytes.IndexByte(arr[:], 0)
	if n == -1 {
		n = len(arr)
	}
	return string(arr[:n])
}
