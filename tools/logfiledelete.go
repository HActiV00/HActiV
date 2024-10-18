package tools

import (
	"C"
	"fmt"
	"os"
	"strings"
	"time"
	"unsafe"

	bcc "github.com/iovisor/gobpf/bcc"
)
import (
	"HActiV/pkg/docker"
	"HActiV/pkg/utils"
	bpfcode "HActiV/tools/bpf"
	"bytes"
)

// 이벤트 데이터 구조체
type logFileDeleteEvent struct {
	Uid      uint32
	Gid      uint32
	Pid      uint32
	PPid     uint32
	Comm     [16]byte
	Filename [200]byte
	Op       uint32 // 1: truncate, 2: delete
}

var opcode_array = [2]string{"truncate", "delete"}

// eBPF 프로그램

func LogFileDeleteMonitoring() {

	// 무시할 프로세스 이름 목록 (텍스트 편집기 및 기타 불필요한 프로세스)
	ignoreList := []string{"vi", "vim", "nano", "less", "tracker-store", "gedit"}

	// BPF 프로그램 컴파일
	bpfModule := utils.LoadBPFModule(bpfcode.LogFileDelete)
	defer bpfModule.Close()
	// kprobe 로드 및 unlink 시스템 콜 추적
	kprobeUnlink, err := bpfModule.LoadKprobe("trace_unlink")
	if err != nil {
		panic(err)
	}

	// __x64_sys_unlink 및 __x64_sys_unlinkat 시스템 콜에 kprobe 연결
	err = bpfModule.AttachKprobe("__x64_sys_unlink", kprobeUnlink, -1)
	if err != nil {
		panic(err)
	}
	err = bpfModule.AttachKprobe("__x64_sys_unlinkat", kprobeUnlink, -1)
	if err != nil {
		panic(err)
	}

	// kprobe 로드 및 truncate 시스템 콜 추적
	kprobeTruncate, err := bpfModule.LoadKprobe("trace_truncate")
	if err != nil {
		panic(err)
	}

	// __x64_sys_truncate 및 __x64_sys_ftruncate 시스템 콜에 kprobe 연결
	err = bpfModule.AttachKprobe("__x64_sys_truncate", kprobeTruncate, -1)
	if err != nil {
		panic(err)
	}
	err = bpfModule.AttachKprobe("__x64_sys_ftruncate", kprobeTruncate, -1)
	if err != nil {
		panic(err)
	}

	// Perf map 생성 및 이벤트 핸들러 설정
	table := bcc.NewTable(bpfModule.TableId("events"), bpfModule)
	channel := make(chan []byte)
	perfMap, err := bcc.InitPerfMap(table, channel, nil)
	if err != nil {
		panic(err)
	}

	// 이벤트 수신 및 처리
	go func() {
		var event logFileDeleteEvent
		for {
			data := <-channel

			// 이벤트 데이터를 언마샬
			event = *(*logFileDeleteEvent)(unsafe.Pointer(&data[0]))

			// 프로세스 이름 확인
			processName := strings.Trim(string(event.Comm[:]), "\x00")

			// 무시할 프로세스인지 확인
			skip := false
			for _, name := range ignoreList {
				if processName == name {
					skip = true
					break
				}
			}
			if skip {
				continue
			}

			// 네임스페이스 정보 확인
			nsPath := fmt.Sprintf("/proc/%d/ns/pid", event.PPid)
			nsLink, err := os.Readlink(nsPath)

			if err != nil {
				// 짧은 수명 프로세스로 인해 네임스페이스 정보를 얻지 못하는 경우 건너뛰기
				if os.IsNotExist(err) {
					fmt.Printf("Process with PID %d already exited. Skipping...\n", event.Pid)
					continue
				}
				fmt.Printf("Error checking namespace for PID %d: %v\n", event.Pid, err)
				continue
			}
			containerNamespaces := docker.GetContainer()

			// 컨테이너 이름 확인
			inode, err := utils.GetNamespaceInode(event.Pid)
			if err != nil {
				fmt.Printf("failed to get namespace for PID %d: %s\n", event.Pid, err)
				continue
			}

			containerInfo, exists := containerNamespaces[inode]
			if !exists {
				continue
			}
			utils.DataSend("log_file_"+opcode_array[int(event.Op)-1], time.Now().Format(time.RFC3339), containerInfo.Name, event.Uid, event.Gid, event.Pid, event.PPid,
				string(bytes.TrimRight(event.Comm[:], "\x00")),
				string(bytes.TrimRight(event.Filename[:], "\x00")))
			// 이벤트 타입에 따라 출력 메시지 변경 (컨테이너 프로세스만)
			fmt.Printf("Syslog file %s attempted by PID: %d (Process: %s, Running inside Container '%s', Namespace: %s)\n", opcode_array[int(event.Op)-1], event.Pid, processName, containerInfo.Name, nsLink)

		}
	}()

	fmt.Println("Monitoring syslog file deletion and truncation (only for containers)...")
	perfMap.Start()
	defer perfMap.Stop()

	// 무기한 대기
	select {}
}
