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
	"time"

	bpf "github.com/iovisor/gobpf/bcc"
)

// eBPF 프로그램 소스코드

// event 구조체 정의
type memoryEvent struct {
	Pid         uint32
	PPid        uint32
	Uid         uint32
	Gid         uint32
	ReturnValue int32
	Comm        [16]byte
	Syscall     [16]byte
	EventType   [16]byte
	StartAddr   uint64
	EndAddr     uint64
}

// 필터링 로직을 적용하여 중복되거나 무의미한 이벤트를 필터링
func filterEvent(event memoryEvent) bool {
	// 길이가 0인 이벤트 필터링
	if event.StartAddr == 0 || event.EndAddr == 0 || event.StartAddr == event.EndAddr {
		return false
	}

	// 4KB 이하 메모리 범위 이벤트 필터링
	if event.EndAddr-event.StartAddr <= 4096 {
		return false
	}
	return true
}

func MemoryMonitoring() {
	// eBPF 모듈 로드
	bpfModule := utils.LoadBPFModule(bpfcode.MemoryCcode)
	defer bpfModule.Close()

	// 커널 함수 이름 설정
	mmapFuncName := "__x64_sys_mmap"
	mprotectFuncName := "__x64_sys_mprotect"
	readFuncName := "__x64_sys_read"
	writeFuncName := "__x64_sys_write"

	// mmap 및 mprotect 시스템 호출과 관련된 kprobe 및 kretprobe 설정
	attachKprobe(bpfModule, "kprobe__sys_mmap", mmapFuncName)
	attachKretprobe(bpfModule, "kretprobe__sys_mmap", mmapFuncName)
	attachKprobe(bpfModule, "kprobe__sys_mprotect", mprotectFuncName)
	attachKretprobe(bpfModule, "kretprobe__sys_mprotect", mprotectFuncName)
	attachKprobe(bpfModule, "kprobe__sys_read", readFuncName)
	attachKprobe(bpfModule, "kprobe__sys_write", writeFuncName)

	// 이벤트 테이블 설정 및 시작
	table := bpf.NewTable(bpfModule.TableId("events"), bpfModule)
	channel := make(chan []byte)
	perfMap, err := bpf.InitPerfMap(table, channel, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init perf map: %s\n", err)
		os.Exit(1)
	}

	// 종료 신호 감지
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	// 수신된 이벤트 처리 고루틴
	go processEvents(channel)

	// 퍼포먼스 맵 시작 및 종료 대기
	perfMap.Start()
	<-sig
	perfMap.Stop()
}

// 프로세스 이벤트 핸들러
func processEvents(channel chan []byte) {
	var event memoryEvent
	for {
		data := <-channel
		err := binary.Read(bytes.NewBuffer(data), binary.LittleEndian, &event)
		if err != nil {
			fmt.Printf("failed to decode received data: %s\n", err)
			continue
		}
		containerNamespaces := docker.GetContainer()
		inode, err := utils.GetNamespaceInode(event.PPid)
		if err != nil {
			continue
		}

		containerInfo, exists := containerNamespaces[inode]
		if !exists {
			continue
		}

		if containerInfo.Name == "Host" {
			continue
		}

		// 필터링 적용
		if !filterEvent(event) {
			continue
		}
		//데이터 전송
		/*utils.DataSend("Memory", time.Now().Format(time.RFC3339), containerInfo.Name, event.Uid, event.Gid, event.Pid, event.PPid,
		string(event.Comm[:bytes.IndexByte(event.Comm[:], 0)]),
		string(event.Syscall[:bytes.IndexByte(event.Syscall[:], 0)]),
		string(event.EventType[:bytes.IndexByte(event.EventType[:], 0)]),
		event.StartAddr, event.EndAddr, event.ReturnValue)
		*/
		// 이벤트 출력
		fmt.Printf("%s | Container Name: %s | PPID : %d | PID : %d | GID : %d | UID : %d | command : %s | syscall: %s | event_type: %s | start_addr: 0x%x | end_addr: 0x%x | (return %d)\n",
			time.Now().Format(time.RFC3339),
			containerInfo.Name,
			event.PPid,
			event.Pid,
			event.Gid,
			event.Uid,
			string(event.Comm[:bytes.IndexByte(event.Comm[:], 0)]),
			string(event.Syscall[:bytes.IndexByte(event.Syscall[:], 0)]),
			string(event.EventType[:bytes.IndexByte(event.EventType[:], 0)]),
			event.StartAddr,
			event.EndAddr,
			event.ReturnValue)
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
		fmt.Fprintf(os.Stderr, "Failed to attach %s: %s\n", probeName, err)
		os.Exit(1)
	}
}

// kretprobe 연결 함수
func attachKretprobe(m *bpf.Module, probeName, funcName string) {
	probe, err := m.LoadKprobe(probeName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load %s: %s\n", probeName, err)
		os.Exit(1)
	}
	err = m.AttachKretprobe(funcName, probe, -1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to attach %s: %s\n", probeName, err)
		os.Exit(1)
	}
}
