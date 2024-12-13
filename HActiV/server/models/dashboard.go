package models

import (
	"encoding/json"
	"time"
	"sync"

	"github.com/beego/beego/v2/core/logs"
	"server/kafka"
)

// BasicApiData contains common fields for all API data structures
type BasicApiData struct {
	EventType     string    `json:"event_type"`
	Time          time.Time `json:"timestamp"`
	ContainerName string    `json:"container_name"`
	Uid           uint32    `json:"uid"`
	Gid           uint32    `json:"gid"`
	Pid           uint32    `json:"pid"`
	Ppid          uint32    `json:"ppid"`
}

// ExecveApiData represents data for execve events
type ExecveApiData struct {
	BasicApiData
	Command     string `json:"command"`
	ProcessName string `json:"process_name"`
	Args        string `json:"arguments"`
}

// OpenApiData represents data for file open events
type OpenApiData struct {
	BasicApiData
	Command     string `json:"command"`
	Filename    string `json:"filename"`
	ReturnValue int32  `json:"status"`
	ProcessName string `json:"process_name"`
}

// Node represents a node in the network path
type Node struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// Link represents a link between nodes in the network path
type Link struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

// Path represents the network path
type Path struct {
	Nodes []Node `json:"nodes"`
	Links []Link `json:"links"`
}

// NetworkApiData represents data for network events
type NetworkApiData struct {
	EventType     string    `json:"event_type"`
	Time          time.Time `json:"timestamp"`
	ContainerName string    `json:"container_name"`
	SrcIp         string    `json:"src_ip"`
	SrcIpLabel    string    `json:"src_ip_label"`
	DstIp         string    `json:"dst_ip"`
	DstIpLabel    string    `json:"dst_ip_label"`
	Protocol      string    `json:"protocol"`
	PacketSize    int       `json:"packet_size"`
	TotalPackets  int       `json:"total_packets"`
	TotalSize     int       `json:"total_size"`
	Path          string    `json:"path"`
	Direction     string    `json:"direction"`
	Method        string    `json:"http_method,omitempty"`
	Host          string    `json:"http_host,omitempty"`
	URL           string    `json:"http_url,omitempty"`
	Parameters    string    `json:"http_parameters,omitempty"`
}

// MemoryApiData represents data for memory events
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

// DeleteApiData represents data for delete events
type DeleteApiData struct {
	BasicApiData
	ProcessName string `json:"process_name"`
	Filename    string `json:"filename"`
}

// LogAccessApiData represents data for log access events
type LogAccessApiData struct {
	BasicApiData
	ProcessName string `json:"process_name"`
	Filename    string `json:"filename"`
	FileSize    int64  `json:"file_size"`
	MountStatus string `json:"mount_status"`
}

// ContainerMetricsApiData represents data for container metrics
type ContainerMetricsApiData struct {
	EventType   string    `json:"event_type"`
	Time        time.Time `json:"timestamp"`
	Name        string    `json:"container_name"`
	CpuUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	DiskUsage   float64   `json:"disk_usage"`
	RxBytes     uint64    `json:"rx_bytes"`
	TxBytes     uint64    `json:"tx_bytes"`
}

// HostMetricsApiData represents data for host metrics
type HostMetricsApiData struct {
	EventType   string    `json:"event_type"`
	Time        time.Time `json:"timestamp"`
	CpuUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	DiskUsage   float64   `json:"disk_usage"`
	CpuCores    int       `json:"cpu_cores"`
}

var (
	dashboardData []interface{}
	dataMutex     sync.RWMutex
)

func init() {
	dashboardData = make([]interface{}, 0)
}

// SaveExecveData saves execve event data to Kafka
func SaveExecveData(data *ExecveApiData) error {
	return kafka.ProduceMessage("execve_events", data.ContainerName, data)
}

// SaveOpenData saves file open event data to Kafka
func SaveOpenData(data *OpenApiData) error {
	return kafka.ProduceMessage("open_events", data.ContainerName, data)
}

// SaveNetworkData saves network event data to Kafka
func SaveNetworkData(data *NetworkApiData) error {
	return kafka.ProduceMessage("network_events", data.ContainerName, data)
}

// SaveMemoryData saves memory event data to Kafka
func SaveMemoryData(data *MemoryApiData) error {
	return kafka.ProduceMessage("memory_events", data.ContainerName, data)
}

// SaveDeleteData saves delete event data to Kafka
func SaveDeleteData(data *DeleteApiData) error {
	return kafka.ProduceMessage("delete_events", data.ContainerName, data)
}

// SaveLogAccessData saves log access event data to Kafka
func SaveLogAccessData(data *LogAccessApiData) error {
	return kafka.ProduceMessage("log_access_events", data.ContainerName, data)
}

// SaveContainerMetricsData saves container metrics data to Kafka
func SaveContainerMetricsData(data *ContainerMetricsApiData) error {
	return kafka.ProduceMessage("container_metrics_events", data.Name, data)
}

// SaveHostMetricsData saves host metrics data to Kafka
func SaveHostMetricsData(data *HostMetricsApiData) error {
	return kafka.ProduceMessage("host_metrics_events", "host", data)
}

// GetDashboardData retrieves dashboard data for a specific event type
func GetDashboardData(eventType string) ([]interface{}, error) {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	if eventType == "all" {
		return dashboardData, nil
	}

	filteredData := make([]interface{}, 0)
	for _, item := range dashboardData {
		if itemMap, ok := item.(map[string]interface{}); ok {
			if itemMap["event_type"] == eventType {
				filteredData = append(filteredData, item)
			}
		}
	}

	return filteredData, nil
}

// InitializeKafkaConsumers sets up Kafka consumers for all required topics
func InitializeKafkaConsumers() error {
	topics := []string{
		"execve_events",
		"open_events",
		"network_events",
		"memory_events",
		"delete_events",
		"log_access_events",
		"container_metrics_events",
		"host_metrics_events",
	}

	for _, topic := range topics {
		err := kafka.ConsumeMessages(topic, handleKafkaMessage)
		if err != nil {
			logs.Error("Error initializing Kafka consumer for topic %s: %v", topic, err)
			return err
		}
	}

	return nil
}

// handleKafkaMessage processes incoming Kafka messages and stores them in memory
func handleKafkaMessage(msg []byte) error {
	var data interface{}
	err := json.Unmarshal(msg, &data)
	if err != nil {
		return err
	}

	dataMutex.Lock()
	dashboardData = append(dashboardData, data)
	dataMutex.Unlock()

	return nil
}

// GetMessageChannel returns the channel for real-time messages
func GetMessageChannel() chan []byte {
	return kafka.GetMessageChannel()
}

