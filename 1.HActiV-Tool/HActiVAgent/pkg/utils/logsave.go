// Copyright Authors of HActiV

// utils package for helping other package
package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/klauspost/compress/zstd"
)

type DualLogger struct {
	jsonEncoder  *zstd.Encoder
	compressFile *os.File
	jsonFile     *os.File
	logChan      chan Event
	wg           sync.WaitGroup
}

// execve -> Systemcall Datasend와 필드 일치를 위해해
var ToolFields = map[string]string{
	"Systemcall":      "Time ContainerName Uid Gid Pid Ppid Puid Pgid Filename ProcessName Args",
	"file_open":            "Time ContainerName Uid Gid Pid Ppid Filename ProcessName",
	"delete":          "Time ContainerName Uid Gid Pid Ppid Filename ProcessName",
	"Memory":          "Time ContainerName Uid Gid Pid Ppid  ProcessName Syscall StartAddr EndAddr Size Prottemp Prot MappingType",
	"Network_traffic": "Time ContainerName SrcIp SrcIpLabel DstIp DstIpLabel Direction Protocol SrcPort DstPort PacketSize",
}

func NewDualLogger(compressFilename, jsonFilename string) (*DualLogger, error) {
	compressFile, err := os.OpenFile(LogLocation+compressFilename+".zst", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	jsonFile, err := os.OpenFile(LogLocation+jsonFilename+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		compressFile.Close()
		return nil, err
	}

	jsonEncoder, err := zstd.NewWriter(compressFile, zstd.WithEncoderLevel(zstd.SpeedBetterCompression))
	if err != nil {
		compressFile.Close()
		jsonFile.Close()
		return nil, err
	}

	logger := &DualLogger{
		jsonEncoder:  jsonEncoder,
		compressFile: compressFile,
		jsonFile:     jsonFile,
		logChan:      make(chan Event, 1000),
	}

	logger.wg.Add(1)
	go logger.writeLoop()
	return logger, nil
}

func (l *DualLogger) writeLoop() {
	defer l.wg.Done()
	for event := range l.logChan {
		l.writeCompressLog(event)
		l.writeJSONLog(event)
	}
}

func (l *DualLogger) writeCompressLog(event Event) {
	v := reflect.ValueOf(event)
	t := v.Type()

	var jsonFields []string
	tool := v.Field(0).String()
	for i := 1; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := t.Field(i).Name
		if tool == "network" && (fieldName == "Prot" || fieldName == "Size") {
			continue
		}
		if strings.Contains(ToolFields[tool], fieldName) {
			fieldValue := field.Interface()
			jsonValue, _ := json.Marshal(fieldValue)
			jsonFields = append(jsonFields, fmt.Sprintf(`"%s":%s`, fieldName, string(jsonValue)))
		}
	}

	if len(jsonFields) > 0 {
		jsonString := "{" + strings.Join(jsonFields, ",") + "}\n"
		_, err := l.jsonEncoder.Write([]byte(jsonString))
		if err != nil {
			fmt.Println("Error writing to JSON log:", err)
		}
		l.jsonEncoder.Flush()
	}
}

func (l *DualLogger) writeJSONLog(event Event) {
	v := reflect.ValueOf(event)
	t := v.Type()

	var jsonFields []string
	tool := v.Field(0).String()
	for i := 1; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := t.Field(i).Name
		if tool == "network" && (fieldName == "Prot" || fieldName == "Size") {
			continue
		}
		if strings.Contains(ToolFields[tool], fieldName) {
			fieldValue := field.Interface()
			jsonValue, _ := json.Marshal(fieldValue)
			jsonFields = append(jsonFields, fmt.Sprintf(`"%s":%s`, fieldName, string(jsonValue)))
		}
	}

	if len(jsonFields) > 0 {
		jsonString := "{" + strings.Join(jsonFields, ",") + "}\n"
		l.jsonFile.WriteString(jsonString)
	}
}

func (l *DualLogger) Log(event Event) {
	l.logChan <- event
}

func (l *DualLogger) Close() {
	close(l.logChan)
	l.wg.Wait()
	l.jsonEncoder.Flush()
	l.jsonEncoder.Close()
	l.compressFile.Sync()
	l.jsonFile.Sync()
	l.compressFile.Close()
	l.jsonFile.Close()
}
