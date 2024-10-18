package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

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
	ReturnValue int    `json:"status"`
}

type OpenApiData struct {
	BasicApiData
	Command     string `json:"command"`
	Filename    string `json:"filename"`
	ReturnValue int32  `json:"status"`
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

type LogDeleteApiData struct {
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

const (
	apiKey = "[]"
	apiURL = "[]"
)

func DataSend(args ...interface{}) {
	// 전송할 데이터 준비
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
			ReturnValue:  args[10].(int),
		}
	case "File_open":
		data = OpenApiData{
			BasicApiData: basicData,
			Command:      args[7].(string),
			Filename:     args[8].(string),
			ReturnValue:  args[9].(int32),
		}
	case "Network_traffic":
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
	case "log_file_truncate":
		data = LogDeleteApiData{
			BasicApiData: basicData,
			ProcessName:  args[7].(string),
			Filename:     args[8].(string),
		}
	case "log_file_delete":
		data = LogDeleteApiData{
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
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
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

	fmt.Println("응답 상태:", resp.Status)
	fmt.Println("응답 본문:", string(body))
}
