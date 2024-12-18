// Copyright Authors of HActiV

// tool package: 3.delete.go
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
	"syscall"
	"time"

	bcc "github.com/iovisor/gobpf/bcc"
)

type DeleteEvent struct {
	Uid           uint32
	Gid           uint32
	Pid           uint32
	PPid          uint32
	Comm          [16]byte
	Filename      [200]byte
	Op            uint32
	NamespaceInum uint32
}

func DeleteMonitoring() {
	policies, err := configs.LoadRules("delete")
	if err != nil {
		fmt.Fprintf(os.Stderr, "정책 로드 실패: %v\n", err)
		os.Exit(1)
	}

	bpfModule := utils.LoadBPFModule(bpfcode.Delete)
	defer bpfModule.Close()

	tracepointUnlink, err := bpfModule.LoadKprobe("trace_unlinkat")
	if err != nil {
		fmt.Printf("Failed to load kprobe unlink: %v\n", err)
		return
	}

	err = bpfModule.AttachKprobe("do_unlinkat", tracepointUnlink, -1)
	if err != nil {
		fmt.Printf("Failed to attach kprobe do_unlinkat: %v\n", err)
		return
	}

	tracepointTruncate, err := bpfModule.LoadKprobe("trace_truncate")
	if err != nil {
		fmt.Printf("Failed to load kprobe truncate: %v\n", err)
		return
	}
	err = bpfModule.AttachKprobe("do_truncate", tracepointTruncate, -1)
	if err != nil {
		fmt.Printf("Failed to attach kprobe do_truncate: %v\n", err)
		return
	}

	table := bcc.NewTable(bpfModule.TableId("events"), bpfModule)
	channel := make(chan []byte)
	lost := make(chan uint64)
	//perfMap, err := bcc.InitPerfMap(table, channel, lost)
	perfMap, err := bcc.InitPerfMapWithPageCnt(table, channel, lost, 512)
	if err != nil {
		fmt.Printf("Failed to initialize PerfMap: %v\n", err)
		return
	}

	var lostCount uint64
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	logger, err := utils.NewDualLogger("delete_compress", "deletejson")
	if err != nil {
		fmt.Println("로그 생성 실패:", err)
		return
	}
	defer logger.Close()

	go func() {
		var event DeleteEvent
		containerNamespaces := docker.GetContainer()

		for {
			select {
			case data := <-channel:
				err := binary.Read(bytes.NewBuffer(data), binary.LittleEndian, &event)
				if err != nil {
					fmt.Printf("수신한 데이터 디코딩 실패: %s\n", err)
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
				filename := string(bytes.TrimRight(event.Filename[:], "\x00"))

				matchevent := utils.Event{
					Tool:          "delete",
					Time:          time.Now().Format(time.RFC3339),
					Uid:           event.Uid,
					Gid:           event.Gid,
					Pid:           event.Pid,
					Filename:      filename,
					ProcessName:   processName,
					ContainerName: containerInfo.Name,
				}

				configs.MatchedEvent(policies, matchevent)
				logger.Log(matchevent)
				if configs.DataSend {
					utils.DataSend("delete", matchevent.Time, containerInfo.Name, event.Uid, event.Gid, event.Pid, event.PPid, processName, filename)
				}
			case lostCountData := <-lost:
				lostCount += lostCountData
			}
		}
	}()

	fmt.Println("[Delete event monitoring start...]")
	perfMap.Start()
	defer perfMap.Stop()
	<-sig
	fmt.Printf("[DeleteEvent] Lost Count: %d\n", lostCount)
}
