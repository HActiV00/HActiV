// Copyright Authors of HActiV

// utils package for helping other package
package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// API Rule 구조체 기존 값 + 정책이름 + printFormat 추가
type BasicApiRule struct {
	PolicyName    string `json:"policy_name"`
	PrintFormat   string `json:"print_format"`
	EventType     string `json:"event_type"`
	Time          string `json:"timestamp"`
	ContainerName string `json:"container_name"`
	Uid           uint32 `json:"uid"`
	Gid           uint32 `json:"gid"`
	Pid           uint32 `json:"pid"`
	Ppid          uint32 `json:"ppid"`
}

type ExecveApiRule struct {
	BasicApiRule
	Command     string `json:"command"`
	ProcessName string `json:"process_name"`
	Args        string `json:"arguments"`
}

type OpenApiRule struct {
	BasicApiRule
	Command     string `json:"command"`
	Filename    string `json:"filename"`
	ReturnValue int32  `json:"status"`
	ProcessName string `json:"process_name"`
}

type NetworkApiRule struct {
	PolicyName    string `json:"policy_name"`
	PrintFormat   string `json:"print_format"`
	EventType     string `json:"event_type"`
	Time          string `json:"timestamp"`
	ContainerName string `json:"container_name"`
	SrcIp         string `json:"src_ip"`
	SrcIpLabel    string `json:"src_ip_label"`
	DstIp         string `json:"dst_ip"`
	DstIpLabel    string `json:"dst_ip_label"`
	Protocol      string `json:"protocol"`
	PacketSize    int    `json:"packet_size"`
	TotalPackets  int    `json:"total_packets"`
	TotalSize     int    `json:"total_size"`
	Path          string `json:"path"`
	Direction     string `json:"direction"`
	Method        string `json:"http_method,omitempty"`
	Host          string `json:"http_host,omitempty"`
	URL           string `json:"http_url,omitempty"`
	Parameters    string `json:"http_parameters,omitempty"`
}

type MemoryApiRule struct {
	BasicApiRule
	ProcessName  string `json:"process_name"`
	Syscall      string `json:"syscall"`
	Prot         string `json:"prot"`
	Prottemp     uint32 `json:"prottemp"`
	MappingType  string `json:"mapping_type"`
	StartAddress uint64 `json:"start_address"`
	EndAddress   uint64 `json:"end_address"`
	Size         uint64 `json:"size"`
}

type DeleteApiRule struct {
	BasicApiRule
	ProcessName string `json:"process_name"`
	Filename    string `json:"filename"`
}

type LogAccessApiRule struct {
	BasicApiRule
	ProcessName string `json:"process_name"`
	Filename    string `json:"filename"`
	FileSize    int64  `json:"file_size"`
	MountStatus string `json:"mount_status"`
}

func RuleSend(args ...interface{}) {
	var basicData BasicApiRule
	//EventType 구조체 2번으로 변경 기존 0번
	if args[2].(string) != "Network_traffic" && args[0].(string) != "ContainerMetrics" && args[0].(string) != "HostMetrics" {
		basicData = BasicApiRule{
			PolicyName:    args[0].(string),
			PrintFormat:   args[1].(string),
			EventType:     args[2].(string),
			Time:          args[3].(string),
			ContainerName: args[4].(string),
			Uid:           args[5].(uint32),
			Gid:           args[6].(uint32),
			Pid:           args[7].(uint32),
			Ppid:          args[8].(uint32),
		}
	} else {
		//EventType 구조체 2번으로 변경 기존 0번
		basicData = BasicApiRule{
			EventType: args[2].(string),
		}
	}
	//기존 필드값 +2로 변경 정책 이름, PrintFormat 추가로 2칸씩 +
	var data interface{}
	switch basicData.EventType {
	case "Systemcall":
		data = ExecveApiRule{
			BasicApiRule: basicData,
			Command:      args[9].(string),
			ProcessName:  args[10].(string),
			Args:         args[11].(string),
		}
	case "file_open":
		data = OpenApiRule{
			BasicApiRule: basicData,
			Command:      args[9].(string),
			Filename:     args[10].(string),
			ReturnValue:  args[11].(int32),
			ProcessName:  args[12].(string),
		}
	case "Network_traffic":
		srcNode := Node{ID: args[5].(string), Type: args[6].(string)}
		dstNode := Node{ID: args[7].(string), Type: args[8].(string)}
		link := Link{Source: args[5].(string), Target: args[7].(string)}
		path := Path{
			Nodes: []Node{srcNode, dstNode},
			Links: []Link{link},
		}

		pathJSON, err := json.Marshal(path)
		if err != nil {
			fmt.Println("Path JSON 변환 오류:", err)
			return
		}

		data = NetworkApiRule{
			PolicyName:    args[0].(string),
			PrintFormat:   args[1].(string),
			EventType:     args[2].(string),
			Time:          args[3].(string),
			ContainerName: args[4].(string),
			SrcIp:         args[5].(string),
			SrcIpLabel:    args[6].(string),
			DstIp:         args[7].(string),
			DstIpLabel:    args[8].(string),
			Protocol:      args[9].(string),
			PacketSize:    args[10].(int),
			TotalPackets:  args[11].(int),
			TotalSize:     args[12].(int),
			Path:          string(pathJSON),
			Direction:     args[14].(string),
			Method:        args[15].(string),
			Host:          args[16].(string),
			URL:           args[17].(string),
			Parameters:    args[18].(string),
		}
	case "Memory":
		data = MemoryApiRule{
			BasicApiRule: basicData,
			ProcessName:  args[9].(string),
			Syscall:      args[10].(string),
			StartAddress: args[11].(uint64),
			EndAddress:   args[12].(uint64),
			Size:         args[13].(uint64),
			Prottemp:     args[14].(uint32),
			Prot:         args[15].(string),
			MappingType:  args[16].(string),
		}
	case "delete":
		data = DeleteApiRule{
			BasicApiRule: basicData,
			ProcessName:  args[9].(string),
			Filename:     args[10].(string),
		}
	case "log_file_access":
		data = LogAccessApiRule{
			BasicApiRule: basicData,
			ProcessName:  args[9].(string),
			Filename:     args[10].(string),
			FileSize:     args[11].(int64),
			MountStatus:  args[12].(string),
		}
	// case "ContainerMetrics":
	// 	data = ContainerMetricsApiData{
	// 		EventType:   "ContainerMetrics",
	// 		Time:        args[1].(string),
	// 		Name:        args[2].(string),
	// 		CpuUsage:    args[3].(float64),
	// 		MemoryUsage: args[4].(float64),
	// 		DiskUsage:   args[5].(float64),
	// 		RxBytes:     args[6].(uint64),
	// 		TxBytes:     args[7].(uint64),
	// 	}
	// case "HostMetrics":
	// 	data = HostMetricsApiData{
	// 		EventType:   "HostMetrics",
	// 		Time:        args[1].(string),
	// 		CpuUsage:    args[2].(float64),
	// 		MemoryUsage: args[3].(float64),
	// 		DiskUsage:   args[4].(float64),
	// 		CpuCores:    args[5].(int),
	// 	}
	default:
		return
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("JSON 변환 오류:", err)
		return
	}
	// ruleUrl로 전송
	req, err := http.NewRequest("POST", ruleUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("요청 생성 오류:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}
