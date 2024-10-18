package utils

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	bpf "github.com/iovisor/gobpf/bcc"
)

func LoadBPFModule(bpfCode string) *bpf.Module {
	if strings.Contains(bpfCode, "Host_Pid") {
		bpfCode = strings.ReplaceAll(bpfCode, "Host_Pid", strconv.Itoa(os.Getpid()))

	}
	ebpfMoudle := bpf.NewModule(bpfCode, []string{})
	if ebpfMoudle == nil {
		exitWithError("Failed to create BPF module")
	}
	return ebpfMoudle
}

func AttachTracepoint(m *bpf.Module) {
	tracepoint, err := m.LoadTracepoint("")
	if err != nil {
		exitWithError("Failed to load tracepoint: %v", err)
	}

	err = m.AttachTracepoint("syscalls:sys_enter_execve", tracepoint)
	if err != nil {
		exitWithError("Failed to attach tracepoint: %v", err)
	}
}

func InitPerfMap(m *bpf.Module, table_id string) (*bpf.PerfMap, chan []byte) {
	table := bpf.NewTable(m.TableId(table_id), m)
	channel := make(chan []byte)

	perfMap, err := bpf.InitPerfMap(table, channel, nil)
	if err != nil {
		exitWithError("Failed to init perf map: %v", err)
	}
	perfMap.Start()
	return perfMap, channel
}

func HandleSignals(perfMap *bpf.PerfMap) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	perfMap.Stop()
}

func exitWithError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)

}
