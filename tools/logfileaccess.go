package tools

import (
	"HActiV/pkg/utils"
	bpfcode "HActiV/tools/bpf"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
	"unsafe"

	bcc "github.com/iovisor/gobpf/bcc"
)

// 이벤트 구조체
type logFileAccessEvent struct {
	Uid      uint32
	Gid      uint32
	Pid      uint32
	PPid     uint32
	Comm     [16]byte
	Filename [200]byte
}

var lastAccessedPID uint32
var lastAccessedTime time.Time

// 호스트의 PID 네임스페이스 가져오기
func getHostNamespace() (string, error) {
	nsLink, err := os.Readlink("/proc/1/ns/pid")
	if err != nil {
		return "", err
	}
	return nsLink, nil
}

// 네임스페이스 파일을 읽어와서 호스트 또는 컨테이너를 판별
func checkNamespace(pid uint32, hostNS string) (bool, string, error) {
	nsPath := fmt.Sprintf("/proc/%d/ns/pid", pid)

	// PID가 존재하는지 먼저 확인
	if _, err := os.Stat(nsPath); os.IsNotExist(err) {
		// 프로세스가 종료된 경우
		return false, "", fmt.Errorf("PID %d does not exist (process likely exited)", pid)
	}

	nsLink, err := os.Readlink(nsPath)
	if err != nil {
		return false, "", err
	}

	// 현재 호스트의 네임스페이스와 비교
	isHost := (nsLink == hostNS)

	return isHost, nsLink, nil
}

// 네임스페이스와 일치하는 컨테이너 이름 가져오기 및 마운트 여부 확인
func getContainerNameAndMountStatus(nsLink string) (string, bool, error) {
	// 모든 컨테이너의 정보를 가져와 네임스페이스 비교
	output, err := exec.Command("sudo", "docker", "ps", "-q").Output()
	if err != nil {
		return "", false, err
	}
	containerIDs := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, containerID := range containerIDs {
		containerID = strings.TrimSpace(containerID)
		if containerID == "" {
			continue
		}

		// 컨테이너의 PID 가져오기
		pidOutput, err := exec.Command("sudo", "docker", "inspect", "--format", "{{.State.Pid}}", containerID).Output()
		if err != nil {
			return "", false, err
		}
		containerPID := strings.TrimSpace(string(pidOutput))

		// 컨테이너의 네임스페이스 가져오기
		containerNSPath := fmt.Sprintf("/proc/%s/ns/pid", containerPID)
		containerNSLink, err := os.Readlink(containerNSPath)
		if err != nil {
			return "", false, err
		}

		// 네임스페이스가 일치하면 컨테이너 이름 반환
		if containerNSLink == nsLink {
			nameOutput, err := exec.Command("sudo", "docker", "inspect", "--format", "{{.Name}}", containerID).Output()
			if err != nil {
				return "", false, err
			}
			containerName := strings.Trim(string(nameOutput), "\n/")

			// 마운트 정보 확인
			mountOutput, err := exec.Command("sudo", "docker", "inspect", "--format", "{{.Mounts}}", containerID).Output()
			if err != nil {
				return containerName, false, err
			}
			isMounted := strings.Contains(string(mountOutput), "/var/log/syslog")

			return containerName, isMounted, nil
		}
	}

	return "", false, nil
}

// syslog 파일의 크기 가져오기
func getSyslogSize(syslogPath string) (int64, error) {
	fileInfo, err := os.Stat(syslogPath)
	if err != nil {
		fmt.Printf("Error accessing syslog file at %s: %v\n", syslogPath, err)
		return 0, err
	}
	return fileInfo.Size(), nil
}

func LogFileAccessMonitoring() {
	syslogPath := "/var/log/syslog" // 호스트에 마운트된 syslog 파일 경로

	// 호스트의 네임스페이스 가져오기
	hostNS, err := getHostNamespace()
	if err != nil {
		panic(fmt.Sprintf("Failed to get host namespace: %v", err))
	}

	// BPF 프로그램 컴파일
	bpfModule := utils.LoadBPFModule(bpfcode.LogFileAccess)
	defer bpfModule.Close()

	// kprobe 로드 및 파일 접근 함수 추적
	kprobe, err := bpfModule.LoadKprobe("trace_open")
	if err != nil {
		panic(err)
	}

	err = bpfModule.AttachKprobe("vfs_open", kprobe, -1)
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

	fmt.Println("Monitoring log file access (only for containers)...")

	// 이벤트 수신 및 처리
	go func() {
		var event logFileAccessEvent
		for {
			data := <-channel
			event = *(*logFileAccessEvent)(unsafe.Pointer(&data[0]))

			// 중복 이벤트 무시 로직: 1초 이내 같은 PID의 반복된 이벤트는 무시
			if event.Pid == lastAccessedPID && time.Since(lastAccessedTime) < time.Second {
				continue
			}

			lastAccessedPID = event.Pid
			lastAccessedTime = time.Now()

			// PID 네임스페이스 확인
			_, nsLink, err := checkNamespace(event.Pid, hostNS)
			if err != nil {
				// 프로세스가 종료된 경우 에러 메시지를 출력하지 않고 건너뜀
				if strings.Contains(err.Error(), "process likely exited") {
					continue
				}
				fmt.Printf("Error checking namespace for PID %d: %v\n", event.Pid, err)
				continue
			}

			// 컨테이너 내부에서만 로그 출력
			// 네임스페이스가 컨테이너일 경우, 컨테이너 이름 및 마운트 상태 확인
			containerName, isMounted, err := getContainerNameAndMountStatus(nsLink)
			if err != nil {
				fmt.Printf("Error getting container name or mount status for PID %d: %v\n", event.Pid, err)
				continue
			}

			// syslog 파일 크기 가져오기
			size, err := getSyslogSize(syslogPath)
			if err != nil {
				fmt.Printf("Error getting syslog file size: %v\n", err)
				continue
			}

			// 마운트 여부를 로그에 출력
			mountStatus := "not mounted"
			if isMounted {
				mountStatus = "mounted"
			}

			fmt.Printf("Log file accessed by PID: %d (Running inside Container '%s', Namespace: %s), Syslog Size: %d bytes, Mount Status: %s\n", event.Pid, containerName, nsLink, size, mountStatus)
			utils.DataSend("log_file_access", time.Now().Format(time.RFC3339), containerName, event.Uid, event.Gid, event.Pid, event.PPid,
				string(bytes.TrimRight(event.Comm[:], "\x00")),
				string(bytes.TrimRight(event.Filename[:], "\x00")), size, mountStatus)
		}
	}()

	// PerfMap 시작
	perfMap.Start()
	defer perfMap.Stop()

	// 무기한 대기
	select {}
}
