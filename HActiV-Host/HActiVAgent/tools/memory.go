// Copyright Authors of HActiV

// tool package: 4.memory.go
package tools

import (
	"HActiV/configs"
	"HActiV/pkg/docker"
	"HActiV/pkg/utils"
	bpfcode "HActiV/tools/bpf"
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	bpf "github.com/iovisor/gobpf/bcc"
)

type MemoryEvent struct {
	Uid           uint32
	Gid           uint32
	Pid           uint32
	Ppid          uint32
	Comm          [16]byte
	Syscall       [16]byte
	StartAddr     uint64
	EndAddr       uint64
	Size          uint64
	Prottemp      uint32
	NamespaceInum uint32
	MappingType   [16]byte
}

type MappingCache struct {
	sync.RWMutex
	cache map[uint32]map[uint64]string
}

var mappingCache = &MappingCache{
	cache: make(map[uint32]map[uint64]string),
}

func ProtToString(Prottemp uint32) string {
	var protStr string
	if Prottemp&0x1 != 0 {
		protStr += "r"
	} else {
		protStr += "-"
	}
	if Prottemp&0x2 != 0 {
		protStr += "w"
	} else {
		protStr += "-"
	}
	if Prottemp&0x4 != 0 {
		protStr += "x"
	} else {
		protStr += "-"
	}
	return protStr
}

func updateMappingCache(pid uint32) {
	mappingCache.Lock()
	defer mappingCache.Unlock()

	if _, exists := mappingCache.cache[pid]; !exists {
		mappingCache.cache[pid] = make(map[uint64]string)
	}

	mapsPath := fmt.Sprintf("/proc/%d/maps", pid)
	data, err := ioutil.ReadFile(mapsPath)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		addrRange := strings.Split(fields[0], "-")
		if len(addrRange) != 2 {
			continue
		}

		start, _ := strconv.ParseUint(addrRange[0], 16, 64)
		mappingType := determineTypeFromPath(fields[len(fields)-1], fields[1])
		mappingCache.cache[pid][start] = mappingType
	}
}

func determineTypeFromPath(path, perms string) string {
	switch {
	case strings.Contains(path, "[stack]"):
		return "Stack"
	case strings.Contains(path, "[heap]"):
		return "Heap"
	case strings.Contains(perms, "x"):
		return "Code"
	case strings.Contains(path, ".so"):
		return "Library"
	case strings.Contains(perms, "rw"):
		return "Data"
	case strings.Contains(path, "[vvar]"):
		return "Vvar"
	case strings.Contains(path, "[vdso]"):
		return "Vdso"
	case strings.Contains(path, "[vsyscall]"):
		return "Vsyscall"
	default:
		return "Other"
	}
}

func getCachedMappingType(pid uint32, addr uint64) string {
	mappingCache.RLock()
	defer mappingCache.RUnlock()

	if pidCache, exists := mappingCache.cache[pid]; exists {
		var closestStart uint64
		var mappingType string
		for start, mtype := range pidCache {
			if start <= addr && start > closestStart {
				closestStart = start
				mappingType = mtype
			}
		}
		if mappingType != "" {
			return mappingType
		}
	}

	go updateMappingCache(pid)
	return "Unknown"
}

func MemoryMonitoring() {
	policies, err := configs.LoadRules("memory")
	if err != nil {
		fmt.Fprintf(os.Stderr, "정책 로드 실패: %v\n", err)
		os.Exit(1)
	}

	bpfModule := utils.LoadBPFModule(bpfcode.MemoryCcode)
	defer bpfModule.Close()

	traceMmap, err := bpfModule.LoadTracepoint("tracepoint__syscalls__sys_enter_mmap")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load tracepoint mmap: %v\n", err)
		os.Exit(1)
	}

	err = bpfModule.AttachTracepoint("syscalls:sys_enter_mmap", traceMmap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to attach tracepoint mmap: %v\n", err)
		os.Exit(1)
	}

	traceMprotect, err := bpfModule.LoadTracepoint("tracepoint__syscalls__sys_enter_mprotect")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load tracepoint mprotect: %v\n", err)
		os.Exit(1)
	}

	err = bpfModule.AttachTracepoint("syscalls:sys_enter_mprotect", traceMprotect)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to attach tracepoint mprotect: %v\n", err)
		os.Exit(1)
	}

	table := bpf.NewTable(bpfModule.TableId("events"), bpfModule)
	channel := make(chan []byte)
	lost := make(chan uint64)

	perfMap, err := bpf.InitPerfMap(table, channel, lost)
	//perfMap, err := bpf.InitPerfMapWithPageCnt(table, channel, lost, 512)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize PerfMap: %v\n", err)
		os.Exit(1)
	}

	logger, err := utils.NewDualLogger("memory_compress", "memoryjson")
	if err != nil {
		fmt.Println("로그 생성 실패:", err)
		return
	}
	defer logger.Close()

	var lostCount uint64
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		var event MemoryEvent
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
					continue
				}

				processName := string(bytes.Trim(event.Comm[:], "\x00"))
				syscallName := string(bytes.Trim(event.Syscall[:], "\x00"))
				mappingType := getCachedMappingType(event.Pid, event.StartAddr)
				//matchevent Tool memory -> Memory 수정 Datasend와 일치 시키기 위해
				matchevent := utils.Event{
					Tool:          "Memory",
					Time:          time.Now().Format(time.RFC3339),
					Uid:           event.Uid,
					Gid:           event.Gid,
					Pid:           event.Pid,
					Ppid:          event.Ppid,
					ProcessName:   processName,
					Syscall:       syscallName,
					StartAddr:     event.StartAddr,
					EndAddr:       event.EndAddr,
					Size:          event.Size,
					Prottemp:      event.Prottemp,
					Prot:          ProtToString(event.Prottemp),
					ContainerName: containerInfo.Name,
					MappingType:   mappingType,
				}

				configs.MatchedEvent(policies, matchevent)

				if configs.DataSend {
					utils.DataSend(
						"Memory",
						matchevent.Time,
						matchevent.ContainerName,
						matchevent.Uid,
						matchevent.Gid,
						matchevent.Pid,
						matchevent.Ppid,
						matchevent.ProcessName,
						matchevent.Syscall,
						matchevent.StartAddr,
						matchevent.EndAddr,
						matchevent.Size,
						matchevent.Prottemp,
						matchevent.Prot,
						matchevent.MappingType,
					)
				}

				logger.Log(matchevent)

			case lostCountData := <-lost:
				lostCount += lostCountData
			}
		}
	}()

	perfMap.Start()
	fmt.Println("[Memory event monitoring start...]")
	<-sig
	fmt.Printf("[MemoryEvent] Lost Count: %d\n", lostCount)
	perfMap.Stop()
}
