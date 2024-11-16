package tools

import (
	"HActiV/configs"
	"HActiV/pkg/docker"
	"HActiV/pkg/utils"
	bpfcode "HActiV/tools/bpf"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	bpf "github.com/iovisor/gobpf/bcc"
)

type openEvent struct {
	Pid           uint32
	PPid          uint32
	Uid           uint32
	Gid           uint32
	ReturnValue   int32
	Comm          [16]byte
	Filename      [256]byte
	NamespaceInum uint32
}

// Monitoring Fileopen
func OpenMonitoring() {
	err := configs.SetupRules("open")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load rules: %s\n", err)
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
			processName := strings.Trim(string(event.Comm[:]), "\x00")
			if strings.Contains(filename, "/home/ubuntu/.") || processName == "systemd" || processName == "apport " || strings.Contains(filename, "/usr/bin/which") || filename == "" || strings.Contains(filename, "/usr/share/locale/locale.") || strings.Contains(filename, "/etc/ld.so.cache") || strings.Contains(filename, "/lib/x86_64-linux") || strings.Contains(filename, "/proc/") || processName == "node" || processName == "ps" || processName == "sed" || processName == "gopls" || processName == "sleep" || strings.Contains(filename, "/usr/lib/locale/") || strings.Contains(filename, "cpuUsage.sh") {
				continue
			}
			// 이벤트 데이터 설정
			eventData := map[string]interface{}{
				"event_name":   "File_open",
				"filename":     filename,
				"process_name": processName,
				"uid":          event.Uid,
				"gid":          event.Gid,
			}

			// 규칙과 일치하는 이벤트만 출력
			if configs.ParseRules("File_open", eventData, "open") {
				currentTime := time.Now()
				currentDay := currentTime.Weekday().String()
				currentTimeStr := currentTime.Format("15:04:05")

				// 유효한 요일과 시간인지 확인
				if configs.CheckEffectiveTime("open", currentDay, currentTimeStr) {

					containerNamespaces := docker.GetContainer()
					containerInfo, exists := containerNamespaces[uint64(event.NamespaceInum)]
					if !exists {
						continue
					}

					// 로그 데이터를 JSON으로 출력
					logData := map[string]interface{}{
						"timestamp":    time.Now().Format(time.RFC3339),
						"container":    containerInfo.Name,
						"containerID":  containerInfo.ID,
						"pid":          event.Pid,
						"ppid":         event.PPid,
						"gid":          event.Gid,
						"uid":          event.Uid,
						"file":         filename,
						"command":      processName,
						"return_value": event.ReturnValue,
					}
					utils.DataSend("file_open", time.Now().Format(time.RFC3339), containerInfo.Name, event.Uid, event.Gid, event.Pid, event.PPid,
						string(bytes.Trim(event.Comm[:], "\x00")), filename, event.ReturnValue)
					logJSON, _ := json.Marshal(logData)
					fmt.Println(string(logJSON))
				} else {
					fmt.Printf("Event does not match effective time: %s\n", currentTimeStr)
				}
			}
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
