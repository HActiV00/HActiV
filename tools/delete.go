// delete.go
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

var opcodeArray = [2]string{"truncate", "delete"}

func DeleteMonitoring() {
	err := configs.SetupRules("delete")
	if err != nil {
		fmt.Fprintf(os.Stderr, "정책 로드 실패: %s\n", err)
		os.Exit(1)
	}

	bpfModule := utils.LoadBPFModule(bpfcode.Delete)
	defer bpfModule.Close()

	// `do_unlinkat` 및 `do_truncate`로 변경
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
	perfMap, err := bcc.InitPerfMap(table, channel, nil)
	if err != nil {
		fmt.Printf("Failed to initialize PerfMap: %v\n", err)
		return
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func() {
		var event DeleteEvent
		containerNamespaces := docker.GetContainer()
		recentEvents := make(map[string]time.Time)

		for {
			data := <-channel

			err := binary.Read(bytes.NewBuffer(data), binary.LittleEndian, &event)
			if err != nil {
				fmt.Printf("수신한 데이터 디코딩 실패: %s\n", err)
				continue
			}

			containerInfo, exists := containerNamespaces[uint64(event.NamespaceInum)]
			if !exists {
				continue
			}

			processName := strings.Trim(string(event.Comm[:]), "\x00")
			filename := strings.TrimRight(string(event.Filename[:]), "\x00")
			operation := opcodeArray[int(event.Op)-1]

			cacheKey := fmt.Sprintf("%s:%s:%d:%d", containerInfo.ID, filename, event.Uid, event.Pid)
			if lastTime, found := recentEvents[cacheKey]; found && time.Since(lastTime) < time.Second*5 {
				continue
			}
			recentEvents[cacheKey] = time.Now()

			fmt.Printf("[%s] 컨테이너 이름: %s, ID: %s, UID: %d, GID: %d, PID: %d, PPID: %d, 프로세스명: %s, 파일명: %s, 작업: %s\n",
				time.Now().Format(time.RFC3339),
				containerInfo.Name,
				containerInfo.ID,
				event.Uid,
				event.Gid,
				event.Pid,
				event.PPid,
				processName,
				filename,
				operation,
			)

			eventData := map[string]interface{}{
				"event_name":   "file_delete",
				"filename":     filename,
				"process_name": processName,
				"uid":          event.Uid,
				"gid":          event.Gid,
			}
			utils.DataSend("delete", time.Now().Format(time.RFC3339), containerInfo.Name, event.Uid, event.Gid, event.Pid, event.PPid,
				processName,
				filename)
			if configs.ParseRules("delete", eventData, "delete") {
				utils.DataSend("delete", time.Now().Format(time.RFC3339), containerInfo.Name, event.Uid, event.Gid, event.Pid, event.PPid, processName, filename)
			}
		}
	}()

	fmt.Println("컨테이너 내에서만 파일 삭제 및 축소 모니터링 중...")
	perfMap.Start()
	defer perfMap.Stop()

	<-sig
}
