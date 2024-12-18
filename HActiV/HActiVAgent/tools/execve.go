// Copyright Authors of HActiV

// tool package: 2.execve.go
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

type execvevent struct {
	Uid           uint32
	Gid           uint32
	Pid           uint32
	Ppid          uint32
	Puid          uint32
	Pgid          uint32
	Comm          [16]byte
	Filename      [100]byte
	Args          [200]byte
	NamespaceInum uint32
}

func ExecveMonitoring() {
	policies, err := configs.LoadRules("execve")
	if err != nil {
		fmt.Fprintf(os.Stderr, "정책 로드 실패: %v\n", err)
		os.Exit(1)
	}

	bpfModule := utils.LoadBPFModule(bpfcode.ExecveCcode)
	defer bpfModule.Close()

	tracepoint, err := bpfModule.LoadTracepoint("tracepoint__syscalls__sys_enter_execve")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load tracepoint execve: %v\n", err)
		os.Exit(1)
	}

	err = bpfModule.AttachTracepoint("syscalls:sys_enter_execve", tracepoint)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to attach tracepoint execve: %v\n", err)
		os.Exit(1)
	}

	table := bpf.NewTable(bpfModule.TableId("events"), bpfModule)
	channel := make(chan []byte)
	lost := make(chan uint64)
	//perfMap, err := bpf.InitPerfMap(table, channel, lost)
	perfMap, err := bpf.InitPerfMapWithPageCnt(table, channel, lost, 512)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize PerfMap: %v\n", err)
		os.Exit(1)
	}

	var lostCount uint64
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	logger, err := utils.NewDualLogger("execve_compress", "execvejson")
	if err != nil {
		fmt.Println("로그 생성 실패:", err)
		return
	}
	defer logger.Close()

	go func() {
		var event execvevent
		for {
			select {
			case data := <-channel:
				err := binary.Read(bytes.NewBuffer(data), binary.LittleEndian, &event)
				if err != nil {
					fmt.Printf("데이터 디코딩 실패: %v\n", err)
					continue
				}

				containerNamespaces := docker.GetContainer()
				containerInfo, exists := containerNamespaces[uint64(event.NamespaceInum)]
				if !exists {
					if utils.HostMonitoring {
						containerInfo.Name = "H"
					} else {
						continue
					}
				}

				processName := string(bytes.Trim(event.Comm[:], "\x00"))
				args := convertByteArrayToString(event.Args)
				filename := string(bytes.TrimRight(event.Filename[:], "\x00"))

				//matchevent Tool execve -> Systemcall 수정 Datasend와 일치 시키기 위해
				matchevent := utils.Event{
					Tool:          "Systemcall",
					Time:          time.Now().Format(time.RFC3339),
					Uid:           event.Uid,
					Gid:           event.Gid,
					Pid:           event.Pid,
					Ppid:          event.Ppid,
					Puid:          event.Puid,
					Pgid:          event.Pgid,
					Filename:      filename,
					ProcessName:   processName,
					Args:          args,
					ContainerName: containerInfo.Name,
				}

				configs.MatchedEvent(policies, matchevent)
				logger.Log(matchevent)
				if configs.DataSend {
					utils.DataSend("Systemcall", matchevent.Time, containerInfo.Name, event.Uid, event.Gid, event.Pid, event.Ppid, filename, processName, strings.Replace(args, "--color=auto", "", 1))
				}
			case lostCountData := <-lost:
				lostCount += lostCountData
			}
		}
	}()

	perfMap.Start()
	fmt.Println("[Execve event monitoring start...]")
	<-sig
	fmt.Printf("[ExecveEvent] Lost Count: %d\n", lostCount)
	perfMap.Stop()
}

func convertByteArrayToString(arr [200]byte) string {
	n := bytes.IndexByte(arr[:], 0)
	if n == -1 {
		n = len(arr)
	}
	return string(arr[:n])
}
