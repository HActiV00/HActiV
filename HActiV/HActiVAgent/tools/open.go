// Copyright Authors of HActiV

// tool package: 6.open.go
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
	"strings"
	"syscall"
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

func OpenMonitoring() {
	policies, err := configs.LoadRules("open")
	if err != nil {
		fmt.Fprintf(os.Stderr, "정책 로드 실패: %s\n", err)
		os.Exit(1)
	}

	bpfModule := utils.LoadBPFModule(bpfcode.OpenCcode)
	defer bpfModule.Close()
	syscallName := "syscalls:sys_enter_openat"

	err = AttachEBPFProgram(bpfModule, "trace_sys_enter_openat", syscallName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to attach tracepoint openat %s: %s\n", syscallName, err)
		os.Exit(1)
	}

	table := bpf.NewTable(bpfModule.TableId("events"), bpfModule)
	channel := make(chan []byte)
	lost := make(chan uint64)
	perfMap, err := bpf.InitPerfMap(table, channel, lost)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize PerfMap: %s\n", err)
		os.Exit(1)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	Losts := uint64(0)

	logger, err := utils.NewDualLogger("open.log", "openjson.log")
	if err != nil {
		fmt.Println("로그 생성 실패:", err)
		return
	}
	defer logger.Close()

	go func() {
		var event openEvent

		for {
			select {
			case data := <-channel:
				err := binary.Read(bytes.NewBuffer(data), binary.LittleEndian, &event)
				if err != nil {
					fmt.Printf("데이터 디코딩 실패: %s\n", err)
					continue
				}

				filename := string(event.Filename[:bytes.IndexByte(event.Filename[:], 0)])
				processName := strings.Trim(string(event.Comm[:]), "\x00")
				if strings.Contains(processName, "node") || processName == "node" || filename == "" {
					continue
				}

				containerNamespaces := docker.GetContainer()
				containerInfo, exists := containerNamespaces[uint64(event.NamespaceInum)]
				if !exists {
					continue
				}
				//matchevent Tool open -> file_open 수정 Datasend와 일치 시키기 위해
				matchevent := utils.Event{
					Tool:          "file_open",
					Time:          time.Now().Format(time.RFC3339),
					ContainerName: containerInfo.Name,
					Uid:           event.Uid,
					Gid:           event.Gid,
					Pid:           event.Pid,
					Ppid:          event.PPid,
					Filename:      filename,
					ProcessName:   processName,
				}

				configs.MatchedEvent(policies, matchevent)

				logger.Log(matchevent)

				utils.DataSend(
					"file_open",
					matchevent.Time,
					containerInfo.Name,
					event.Uid,
					event.Gid,
					event.Pid,
					event.PPid,
					"open",
					filename,
					event.ReturnValue,
					processName,
				)

			case lostCountData := <-lost:
				Losts += lostCountData
			case <-sig:
				perfMap.Stop()
				return
			}
		}
	}()

	perfMap.Start()
	fmt.Println("[Open event monitoring start...]")
	<-sig
	fmt.Printf("[OpenEvent] Lost Count: %d\n", Losts)
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
