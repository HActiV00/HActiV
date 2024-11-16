//Copyright Authors of HActiV

// utils package for helping other package
package utils

import (
	"HActiV/configs"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// API-Key and URL for send Data

var (
	apiKey         string
	url            string
	HostMonitoring bool
)

func init() {
	configs.InitSettings()
	apiKey = configs.API
	url = configs.URL
	HostMonitoring = configs.HostMonitoring
}

//Struct for DataSend API

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
	Packets       int    `json:"packets"`
	Path          Path   `json:"path"`
}

type MemoryApiData struct {
	BasicApiData
	Command      string `json:"command"`
	Syscall      string `json:"syscall"`
	Type         string `json:"type"`
	StartAddress uint64 `json:"start_address"`
	EndAddress   uint64 `json:"end_address"`
	ReturnValue  int32  `json:"status"`
}

type DeleteApiData struct { // 이름 수정 LogDeleteApiData -> DeleteApiData
	BasicApiData
	ProcessName string `json:"process_name"`
	Filename    string `json:"filename"`
}

type LogAccessApiData struct { //open이랑 기능 통합으로 인한 삭제
	BasicApiData
	ProcessName string `json:"process_name"`
	Filename    string `json:"filename"`
	FileSize    int64  `json:"file_size"`
	MountStatus string `json:"mount_status"`
}

type MetricsApiData struct {
	EventType  string  `json:"event_type"`
	Time       string  `json:"timestamp"`
	Name       string  `json:"container_name"`
	Cpu        float32 `json:"cpu"`
	Core       int32   `json:"core"`
	Memory     float32 `json:"memory"`
	Disk       float32 `json:"disk"`
	NetworkIn  float64 `json:"networkin"`
	NetworkOut float64 `json:"networkout"`
}

// Data Send
func DataSend(args ...interface{}) {
	var basicData BasicApiData
	if args[0].(string) != "Network_traffic" {
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
		}
	case "Network_traffic":
		srcNode := Node{ID: args[3].(string), Type: args[4].(string)}
		dstNode := Node{ID: args[5].(string), Type: args[6].(string)}
		link := Link{Source: args[3].(string), Target: args[5].(string)}
		path := Path{
			Nodes: []Node{srcNode, dstNode},
			Links: []Link{link},
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
			Packets:       args[8].(int),
			Path:          path,
		}
	case "Memory":
		data = MemoryApiData{
			BasicApiData: basicData,
			Command:      args[7].(string),
			Syscall:      args[8].(string),
			Type:         args[9].(string),
			StartAddress: args[10].(uint64),
			EndAddress:   args[11].(uint64),
			ReturnValue:  args[12].(int32),
		}
	// case "truncate": // truncate랑 동일하게 하기위해 주석 처리
	// 	data = LogDeleteApiData{
	// 		BasicApiData: basicData,
	// 		ProcessName:  args[7].(string),
	// 		Filename:     args[8].(string),
	// 	}
	case "delete": // 이름 수정 log_file_delete -> delete
		data = DeleteApiData{ // 이름 수정 LogDeleteApiData -> DeleteApiData
			BasicApiData: basicData,
			ProcessName:  args[7].(string),
			Filename:     args[8].(string),
		}
	case "log_file_access": //open이랑 기능 통합으로 인한 삭제
		data = LogAccessApiData{
			BasicApiData: basicData,
			ProcessName:  args[7].(string),
			Filename:     args[8].(string),
			FileSize:     args[9].(int64),
			MountStatus:  args[10].(string),
		}
	case "metrics":
		data = MetricsApiData{
			EventType:  args[0].(string),
			Time:       args[1].(string),
			Name:       args[2].(string),
			Cpu:        args[3].(float32),
			Core:       args[4].(int32),
			Memory:     args[5].(float32),
			Disk:       args[6].(float32),
			NetworkIn:  args[7].(float64),
			NetworkOut: args[8].(float64),
		}
	default:
		return
	}

	// 데이터를 JSON으로 변환
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("JSON 변환 오류:", err)
		return
	}

	// POST 요청 생성
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("요청 생성 오류:", err)
		return
	}

	// 헤더 설정
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey) // API 키 추가

	// 클라이언트 생성 및 요청 전송
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("요청 전송 오류:", err)
		return
	}
	defer resp.Body.Close()

	// 응답 처리
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("응답 읽기 오류:", err)
		return
	}

	// 받은 데이터 크기 확인
	receiveSize := len(body)

	fmt.Printf("%d ", receiveSize)
	fmt.Print(resp.Status, " ")
}
