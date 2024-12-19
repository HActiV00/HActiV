// Copyright Authors of HActiV

// utils package for helping other package
package utils

type Event struct {
	Tool          string
	Time          string
	ContainerName string
	Uid           uint32
	Gid           uint32
	Pid           uint32
	Ppid          uint32
	Puid          uint32
	Pgid          uint32
	ProcessName   string
	Filename      string
	Args          string
	SrcIp         string
	SrcIpLabel    string
	DstIp         string
	DstIpLabel    string
	Direction     string
	Protocol      string
	Syscall       string
	StartAddr     uint64
	EndAddr       uint64
	Size          uint64
	Prottemp      uint32
	Prot          string
	MappingType   string
	HTTPInfo      *HTTPData
	SrcPort       uint16
	DstPort       uint16
	PacketSize    int
	TotalSize     int    //추가
	PacketCount   int    // 추가
	PathJson      string // 추가
	ReturnValue   int32
	Method        string
	Host          string
	URL           string
	Parameters    string
}

type HTTPData struct {
}
