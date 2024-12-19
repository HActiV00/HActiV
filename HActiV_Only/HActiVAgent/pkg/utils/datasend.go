// Copyright Authors of HActiV

// utils package for helping other package
package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	apiKey         string
	dataUrl        string
	ruleUrl        string
	HostMonitoring bool
	LogLocation    string
)

func DataSendSetting() {
	file, err := os.Open("/etc/HActiV/Setting.json")
	if err != nil {
		fmt.Println("파일 열기 중 오류 발생")
		os.Exit(-1)
	}
	defer file.Close()

	var existingData map[string]string
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&existingData); err != nil {
		fmt.Println("JSON 읽기 중 오류 발생")
		os.Exit(-1)
	}

	apiKey = existingData["API"]
	dataUrl = existingData["DataUrl"]
	ruleUrl = existingData["RuleUrl"]
	HostMonitoring, err = strconv.ParseBool(existingData["HostMonitoring"])
	if err != nil {
		HostMonitoring = false // 기본값 false
	}
	LogLocation = existingData["LogLocation"]

	if !strings.HasSuffix(LogLocation, "/") {
		LogLocation = LogLocation + "/"
	}
}

type BasicApiData struct {
	EventType     string `json:"event_type"`
	Time          string `json:"timestamp"`
	ContainerName string `json:"container_name"`
	Uid           uint32 `json:"uid"`
	Gid           uint32 `json:"gid"`
	Pid           uint32 `json:"pid"`
	Ppid          uint32 `json:"ppid"`
}

type ExecveApiData struct {
	BasicApiData
	Command     string `json:"command"`
	ProcessName string `json:"process_name"`
	Args        string `json:"arguments"`
}

type OpenApiData struct {
	BasicApiData
	Command     string `json:"command"`
	Filename    string `json:"filename"`
	ReturnValue int32  `json:"status"`
	ProcessName string `json:"process_name"`
}

type Node struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type Link struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type Path struct {
	Nodes []Node `json:"nodes"`
	Links []Link `json:"links"`
}

type NetworkApiData struct {
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

type MemoryApiData struct {
	BasicApiData
	ProcessName  string `json:"process_name"`
	Syscall      string `json:"syscall"`
	Prot         string `json:"prot"`
	Prottemp     uint32 `json:"prottemp"`
	MappingType  string `json:"mapping_type"`
	StartAddress uint64 `json:"start_address"`
	EndAddress   uint64 `json:"end_address"`
	Size         uint64 `json:"size"`
}

type DeleteApiData struct {
	BasicApiData
	ProcessName string `json:"process_name"`
	Filename    string `json:"filename"`
}

type LogAccessApiData struct {
	BasicApiData
	ProcessName string `json:"process_name"`
	Filename    string `json:"filename"`
	FileSize    int64  `json:"file_size"`
	MountStatus string `json:"mount_status"`
}

type ContainerMetricsApiData struct {
	EventType   string  `json:"event_type"`
	Time        string  `json:"timestamp"`
	Name        string  `json:"container_name"`
	CpuUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	RxBytes     uint64  `json:"rx_bytes"`
	TxBytes     uint64  `json:"tx_bytes"`
}

type HostMetricsApiData struct {
	EventType   string  `json:"event_type"`
	Time        string  `json:"timestamp"`
	CpuUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	CpuCores    int     `json:"cpu_cores"`
}

func DataSend(args ...interface{}) {
	var basicData BasicApiData
	if args[0].(string) != "Network_traffic" && args[0].(string) != "ContainerMetrics" && args[0].(string) != "HostMetrics" {
		basicData = BasicApiData{
			EventType:     args[0].(string),
			Time:          args[1].(string),
			ContainerName: args[2].(string),
			Uid:           args[3].(uint32),
			Gid:           args[4].(uint32),
			Pid:           args[5].(uint32),
			Ppid:          args[6].(uint32),
		}
	} else {
		basicData = BasicApiData{
			EventType: args[0].(string),
		}
	}

	var data interface{}
	switch basicData.EventType {
	case "Systemcall":
		data = ExecveApiData{
			BasicApiData: basicData,
			Command:      args[7].(string),
			ProcessName:  args[8].(string),
			Args:         args[9].(string),
		}
	case "file_open":
		data = OpenApiData{
			BasicApiData: basicData,
			Command:      args[7].(string),
			Filename:     args[8].(string),
			ReturnValue:  args[9].(int32),
			ProcessName:  args[10].(string),
		}
	case "Network_traffic":
		srcNode := Node{ID: args[3].(string), Type: args[4].(string)}
		dstNode := Node{ID: args[5].(string), Type: args[6].(string)}
		link := Link{Source: args[3].(string), Target: args[5].(string)}
		path := Path{
			Nodes: []Node{srcNode, dstNode},
			Links: []Link{link},
		}

		pathJSON, err := json.Marshal(path)
		if err != nil {
			fmt.Println("Path JSON 변환 오류:", err)
			return
		}

		data = NetworkApiData{
			EventType:     args[0].(string),
			Time:          args[1].(string),
			ContainerName: args[2].(string),
			SrcIp:         args[3].(string),
			SrcIpLabel:    args[4].(string),
			DstIp:         args[5].(string),
			DstIpLabel:    args[6].(string),
			Protocol:      args[7].(string),
			PacketSize:    args[8].(int),
			TotalPackets:  args[9].(int),
			TotalSize:     args[10].(int),
			Path:          string(pathJSON),
			Direction:     args[12].(string),
			Method:        args[13].(string),
			Host:          args[14].(string),
			URL:           args[15].(string),
			Parameters:    args[16].(string),
		}
	case "Memory":
		data = MemoryApiData{
			BasicApiData: basicData,
			ProcessName:  args[7].(string),
			Syscall:      args[8].(string),
			StartAddress: args[9].(uint64),
			EndAddress:   args[10].(uint64),
			Size:         args[11].(uint64),
			Prottemp:     args[12].(uint32),
			Prot:         args[13].(string),
			MappingType:  args[14].(string),
		}
	case "delete":
		data = DeleteApiData{
			BasicApiData: basicData,
			ProcessName:  args[7].(string),
			Filename:     args[8].(string),
		}
	case "log_file_access":
		data = LogAccessApiData{
			BasicApiData: basicData,
			ProcessName:  args[7].(string),
			Filename:     args[8].(string),
			FileSize:     args[9].(int64),
			MountStatus:  args[10].(string),
		}
	case "ContainerMetrics":
		data = ContainerMetricsApiData{
			EventType:   "ContainerMetrics",
			Time:        args[1].(string),
			Name:        args[2].(string),
			CpuUsage:    args[3].(float64),
			MemoryUsage: args[4].(float64),
			DiskUsage:   args[5].(float64),
			RxBytes:     args[6].(uint64),
			TxBytes:     args[7].(uint64),
		}
	case "HostMetrics":
		data = HostMetricsApiData{
			EventType:   "HostMetrics",
			Time:        args[1].(string),
			CpuUsage:    args[2].(float64),
			MemoryUsage: args[3].(float64),
			DiskUsage:   args[4].(float64),
			CpuCores:    args[5].(int),
		}
	default:
		return
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("JSON 변환 오류:", err)
		return
	}
	//Data Url로 전송
	req, err := http.NewRequest("POST", dataUrl, bytes.NewBuffer(jsonData))
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
