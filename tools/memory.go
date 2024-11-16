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
	"reflect"
	"time"

	bpf "github.com/iovisor/gobpf/bcc"
)

// memoryEvent 구조체 정의 (NamespaceInum 제거)
type memoryEvent struct {
	Pid           uint32
	PPid          uint32
	Uid           uint32
	Gid           uint32
	Comm          [16]byte
	Syscall       [16]byte
	EventType     [16]byte
	NamespaceInum uint32
	StartAddr     uint64
	EndAddr       uint64
}

type SaveData struct {
	StartAddr     uint64
	EndAddr       uint64
	Pid           uint32
	ContainerName string
	Time          string
}

func MemoryMonitoring() {
	err := configs.SetupRules("memory")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load rules: %s\n", err)
		os.Exit(1)
	}

	bpfModule := utils.LoadBPFModule(bpfcode.MemoryCcode)
	defer bpfModule.Close()

	// mmap, brk, mprotect 시스템 호출을 위한 kprobe/kretprobe 설정
	attachKprobe(bpfModule, "kprobe__sys_mmap", "__x64_sys_mmap")
	//attachKretprobe(bpfModule, "kretprobe__sys_mmap", "__x64_sys_mmap")
	attachKprobe(bpfModule, "kprobe__sys_brk", "__x64_sys_brk")
	//attachKretprobe(bpfModule, "kretprobe__sys_brk", "__x64_sys_brk")
	attachKprobe(bpfModule, "kprobe__sys_mprotect", "__x64_sys_mprotect")
	//attachKretprobe(bpfModule, "kretprobe__sys_mprotect", "__x64_sys_mprotect")

	table := bpf.NewTable(bpfModule.TableId("events"), bpfModule)
	channel := make(chan []byte)
	perfMap, err := bpf.InitPerfMap(table, channel, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init perf map: %s\n", err)
		os.Exit(1)
	}

	// 이벤트 수신 및 처리
	go processEvents(channel)
	perfMap.Start()
	defer perfMap.Stop()

	// 종료 시그널 처리
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
}

// 이벤트 처리 함수
func processEvents(channel chan []byte) {
	var event memoryEvent
	var save = make(map[string]SaveData)
	for {
		data := <-channel
		err := binary.Read(bytes.NewBuffer(data), binary.LittleEndian, &event)
		if err != nil {
			fmt.Printf("Error decoding received data: %s\n", err)
			continue
		}

		containerNamespaces := docker.GetContainer()
		containerInfo, exists := containerNamespaces[uint64(event.NamespaceInum)]
		if !exists {
			continue
		}
		var RealTime = time.Now().Format(time.RFC3339)
		var syscall = string(event.Syscall[:bytes.IndexByte(event.Syscall[:], 0)])

		tmpSave := SaveData{
			StartAddr:     event.StartAddr,
			EndAddr:       event.EndAddr,
			Pid:           event.Pid,
			ContainerName: containerInfo.Name,
			Time:          RealTime,
		}

		if reflect.DeepEqual(tmpSave, save[syscall]) {
			continue
		} else {
			save[syscall] = tmpSave
		}

		eventData := map[string]interface{}{
			"event_name":   "Memory_access",
			"process_name": string(event.Comm[:bytes.IndexByte(event.Comm[:], 0)]),
			"uid":          event.Uid,
			"gid":          event.Gid,
			"start_addr":   event.StartAddr,
			"end_addr":     event.EndAddr,
		}

		utils.DataSend("Memory", time.Now().Format(time.RFC3339), containerInfo.Name, event.Uid, event.Gid, event.Pid, event.PPid,
			string(event.Comm[:bytes.IndexByte(event.Comm[:], 0)]),
			string(event.Syscall[:bytes.IndexByte(event.Syscall[:], 0)]),
			string(event.EventType[:bytes.IndexByte(event.EventType[:], 0)]),
			event.StartAddr, event.EndAddr, int32(0))

		fmt.Printf("%s | Container Name: %s | PPID: %d | PID: %d | GID: %d | UID: %d | Command: %s | Syscall: %s | Event Type: %s | Start Addr: 0x%x | End Addr: 0x%x\n",
			RealTime,
			containerInfo.Name,
			event.PPid,
			event.Pid,
			event.Gid,
			event.Uid,
			eventData["process_name"],
			syscall,
			string(event.EventType[:bytes.IndexByte(event.EventType[:], 0)]),
			event.StartAddr,
			event.EndAddr)
	}

}

// kprobe 연결 함수
func attachKprobe(m *bpf.Module, probeName, funcName string) {
	probe, err := m.LoadKprobe(probeName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load %s: %s\n", probeName, err)
		os.Exit(1)
	}
	err = m.AttachKprobe(funcName, probe, -1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to attach %s to function %s: %s\n", probeName, funcName, err)
		os.Exit(1)
	}
}

// kretprobe 연결 함수
// func attachKretprobe(m *bpf.Module, probeName, funcName string) {
// 	probe, err := m.LoadKprobe(probeName)
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "Failed to load kretprobe %s: %s\n", probeName, err)
// 		os.Exit(1)
// 	}
// 	err = m.AttachKretprobe(funcName, probe, -1)
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "Failed to attach kretprobe %s to function %s: %s\n", probeName, funcName, err)
// 		os.Exit(1)
// 	}
// }
