package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"server/kafka"
	_ "github.com/go-sql-driver/mysql"
)

const (
	kafkaBufferSize          = 10000
	bufferThreshold          = 6000 // 버퍼가 90% 찼을 때
	DefaultDataRetentionDays = 10
	maxRetries               = 30
	retryInterval            = 10 * time.Second
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
}

func (n *NetworkApiData) UnmarshalJSON(data []byte) error {
	type Alias NetworkApiData
	aux := &struct {
		*Alias
		Path json.RawMessage `json:"path"`
	}{
		Alias: (*Alias)(n),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	// Store the path as a string instead of trying to unmarshal it
	n.Path = string(aux.Path)
	return nil
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

type UserSettings struct {
	DataRetentionDays int
}

var (
	dashboardData []interface{}
	dataMutex     sync.RWMutex
	db            *sql.DB
	db2           *sql.DB
	kafkaBuffer   chan []byte
	userSettings  UserSettings
)

func init() {
	dashboardData = make([]interface{}, 0)
	kafkaBuffer = make(chan []byte, kafkaBufferSize)

	// Initialize both database connections
	initDB()
	initDB2()

	createTables()

	// 사용자 설정 초기화
	userSettings = UserSettings{
		DataRetentionDays: DefaultDataRetentionDays,
	}

	go processKafkaMessages()
	go periodicDataCleanup()
}

func initDB() {
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"))

	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open MySQL connection: %v", err)
	}

	for i := 0; i < maxRetries; i++ {
		if err = db.Ping(); err == nil {
			logs.Info("Successfully connected to MySQL (DB1)")
			return
		}
		logs.Warn("Failed to connect to MySQL (DB1). Retrying in %d seconds... (Attempt %d/%d)", retryInterval/time.Second, i+1, maxRetries)
		time.Sleep(retryInterval)
	}

	logs.Error("Failed to connect to MySQL (DB1) after %d attempts: %v", maxRetries, err)
}

func initDB2() {
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		os.Getenv("DB2_USER"),
		os.Getenv("DB2_PASS"),
		os.Getenv("DB2_HOST"),
		os.Getenv("DB2_PORT"),
		os.Getenv("DB2_NAME"))

	db2, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open MySQL connection (DB2): %v", err)
	}

	for i := 0; i < maxRetries; i++ {
		if err = db2.Ping(); err == nil {
			logs.Info("Successfully connected to MySQL (DB2)")
			return
		}
		logs.Warn("Failed to connect to MySQL (DB2). Retrying in %d seconds... (Attempt %d/%d)", retryInterval/time.Second, i+1, maxRetries)
		time.Sleep(retryInterval)
	}

	logs.Error("Failed to connect to MySQL (DB2) after %d attempts: %v", maxRetries, err)
}

func createTables() {
	if db == nil {
		logs.Error("Database connection is not initialized")
		return
	}

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS events (
			id INT AUTO_INCREMENT PRIMARY KEY,
			event_type VARCHAR(50),
			data JSON,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		logs.Error("Failed to create events table: %v", err)
	}
}

func SaveNetworkData(data *NetworkApiData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal NetworkApiData: %v", err)
	}

	select {
	case kafkaBuffer <- jsonData:
		// 메시지가 Kafka 버퍼에 성공적으로 추가됨
		if len(kafkaBuffer) >= bufferThreshold {
			go flushBufferToMySQL()
		}
	default:
		// Kafka 버퍼가 가득 참 (이 경우는 발생하지 않아야 함)
		log.Printf("Kafka buffer is full. This should not happen.")
		return fmt.Errorf("kafka buffer is full")
	}

	return nil
}

func flushBufferToMySQL() {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	if len(kafkaBuffer) == 0 {
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return
	}

	stmt, err := tx.Prepare("INSERT INTO events (event_type, data) VALUES (?, ?)")
	if err != nil {
		log.Printf("Failed to prepare statement: %v", err)
		tx.Rollback()
		return
	}
	defer stmt.Close()

	for len(kafkaBuffer) > 0 {
		select {
		case msg := <-kafkaBuffer:
			var data map[string]interface{}
			if err := json.Unmarshal(msg, &data); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}

			_, err = stmt.Exec(data["event_type"], string(msg))
			if err != nil {
				log.Printf("Failed to insert message: %v", err)
			}
		default:
			break
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		tx.Rollback()
	}
}

func processKafkaMessages() {
	for msg := range kafkaBuffer {
		var data map[string]interface{}
		if err := json.Unmarshal(msg, &data); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}

		dataMutex.Lock()
		dashboardData = append([]interface{}{data}, dashboardData...)
		if len(dashboardData) > 1000 { // 대시보드 데이터 크기 제한
			dashboardData = dashboardData[:1000]
		}
		dataMutex.Unlock()

		// 여기서 WebSocket 클라이언트에게 메시지를 보냅니다
		// sendToWebSocket(msg)
	}
}

func GetDashboardData(eventType string, startTime, endTime time.Time) ([]interface{}, error) {
	if startTime.IsZero() && endTime.IsZero() {
		// Kafka 버퍼의 실시간 데이터 반환
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

	// MySQL에서 과거 데이터 조회
	query := `
		SELECT data FROM events
		WHERE event_type = ? AND timestamp BETWEEN ? AND ?
		AND timestamp >= DATE_SUB(NOW(), INTERVAL ? DAY)
		ORDER BY timestamp DESC
		LIMIT 1000
	`
	if eventType == "all" {
		query = `
			SELECT data FROM events
			WHERE timestamp BETWEEN ? AND ?
			AND timestamp >= DATE_SUB(NOW(), INTERVAL ? DAY)
			ORDER BY timestamp DESC
			LIMIT 1000
		`
	}

	var rows *sql.Rows
	var err error
	if eventType == "all" {
		rows, err = db.Query(query, startTime, endTime, userSettings.DataRetentionDays)
	} else {
		rows, err = db.Query(query, eventType, startTime, endTime, userSettings.DataRetentionDays)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []interface{}
	for rows.Next() {
		var dataStr string
		if err := rows.Scan(&dataStr); err != nil {
			return nil, err
		}

		var data interface{}
		if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
			return nil, err
		}

		result = append(result, data)
	}

	return result, nil
}

func SaveExecveData(data *ExecveApiData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal ExecveApiData: %v", err)
	}

	select {
	case kafkaBuffer <- jsonData:
		if len(kafkaBuffer) >= bufferThreshold {
			go flushBufferToMySQL()
		}
	default:
		log.Printf("Kafka buffer is full. This should not happen.")
		return fmt.Errorf("kafka buffer is full")
	}

	return nil
}

func SaveOpenData(data *OpenApiData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal OpenApiData: %v", err)
	}

	select {
	case kafkaBuffer <- jsonData:
		if len(kafkaBuffer) >= bufferThreshold {
			go flushBufferToMySQL()
		}
	default:
		log.Printf("Kafka buffer is full. This should not happen.")
		return fmt.Errorf("kafka buffer is full")
	}

	return nil
}

func SaveMemoryData(data *MemoryApiData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal MemoryApiData: %v", err)
	}

	select {
	case kafkaBuffer <- jsonData:
		if len(kafkaBuffer) >= bufferThreshold {
			go flushBufferToMySQL()
		}
	default:
		log.Printf("Kafka buffer is full. This should not happen.")
		return fmt.Errorf("kafka buffer is full")
	}

	return nil
}

func SaveDeleteData(data *DeleteApiData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal DeleteApiData: %v", err)
	}

	select {
	case kafkaBuffer <- jsonData:
		if len(kafkaBuffer) >= bufferThreshold {
			go flushBufferToMySQL()
		}
	default:
		log.Printf("Kafka buffer is full. This should not happen.")
		return fmt.Errorf("kafka buffer is full")
	}

	return nil
}

func SaveLogAccessData(data *LogAccessApiData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal LogAccessApiData: %v", err)
	}

	select {
	case kafkaBuffer <- jsonData:
		if len(kafkaBuffer) >= bufferThreshold {
			go flushBufferToMySQL()
		}
	default:
		log.Printf("Kafka buffer is full. This should not happen.")
		return fmt.Errorf("kafka buffer is full")
	}

	return nil
}

func SaveContainerMetricsData(data *ContainerMetricsApiData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal ContainerMetricsApiData: %v", err)
	}

	select {
	case kafkaBuffer <- jsonData:
		if len(kafkaBuffer) >= bufferThreshold {
			go flushBufferToMySQL()
		}
	default:
		log.Printf("Kafka buffer is full. This should not happen.")
		return fmt.Errorf("kafka buffer is full")
	}

	return nil
}

func SaveHostMetricsData(data *HostMetricsApiData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal HostMetricsApiData: %v", err)
	}

	select {
	case kafkaBuffer <- jsonData:
		if len(kafkaBuffer) >= bufferThreshold {
			go flushBufferToMySQL()
		}
	default:
		log.Printf("Kafka buffer is full. This should not happen.")
		return fmt.Errorf("kafka buffer is full")
	}

	return nil
}

func DeleteContainerData(containerName string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}

	_, err = tx.Exec("DELETE FROM events WHERE JSON_EXTRACT(data, '$.container_name') = ?", containerName)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete container data: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	dataMutex.Lock()
	newData := make([]interface{}, 0, len(dashboardData))
	for _, item := range dashboardData {
		if itemMap, ok := item.(map[string]interface{}); ok {
			if itemMap["container_name"] != containerName {
				newData = append(newData, item)
			}
		}
	}
	dashboardData = newData
	dataMutex.Unlock()

	return nil
}

func GetContainerEvents(containerName string) (string, error) {
	query := `
		SELECT data
		FROM events
		WHERE JSON_EXTRACT(data, '$.container_name') = ?
		ORDER BY timestamp DESC
	`

	rows, err := db.Query(query, containerName)
	if err != nil {
		return "", fmt.Errorf("failed to query container events: %v", err)
	}
	defer rows.Close()

	var events []map[string]interface{}
	for rows.Next() {
		var dataStr string
		if err := rows.Scan(&dataStr); err != nil {
			return "", fmt.Errorf("failed to scan row: %v", err)
		}

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
			return "", fmt.Errorf("failed to unmarshal data: %v", err)
		}

		events = append(events, data)
	}

	csvData := "Event Type,Timestamp,UID,GID,PID,PPID,Command,Process Name,Arguments\n"
	for _, event := range events {
		csvData += fmt.Sprintf("%s,%s,%d,%d,%d,%d,%s,%s,%s\n",
			event["event_type"],
			event["timestamp"],
			event["uid"],
			event["gid"],
			event["pid"],
			event["ppid"],
			event["command"],
			event["process_name"],
			event["arguments"])
	}

	return csvData, nil
}

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

func handleKafkaMessage(msg []byte) error {
	select {
	case kafkaBuffer <- msg:
		// Message added to Kafka buffer successfully
		if len(kafkaBuffer) >= bufferThreshold {
			go flushBufferToMySQL()
		}
	default:
		// Kafka buffer is full, store in MySQL
		if err := storeMessageInMySQL(msg); err != nil {
			log.Printf("Failed to store message in MySQL: %v", err)
			return err
		}
	}
	return nil
}

func storeMessageInMySQL(message []byte) error {
	var data map[string]interface{}
	if err := json.Unmarshal(message, &data); err != nil {
		return err
	}

	_, err := db.Exec("INSERT INTO events (event_type, data) VALUES (?, ?)",
		data["event_type"], string(message))
	return err
}

func GetKafkaChannel() chan []byte {
	return kafkaBuffer
}

func GetMessageChannel() chan []byte {
	return kafkaBuffer
}

func periodicDataCleanup() {
	for {
		time.Sleep(24 * time.Hour) // 매일 실행
		cleanupOldData()
	}
}

func cleanupOldData() {
	if db == nil {
		logs.Error("Database connection is not initialized")
		return
	}

	retentionPeriod := time.Now().AddDate(0, 0, -userSettings.DataRetentionDays)
	_, err := db.Exec("DELETE FROM events WHERE timestamp < ?", retentionPeriod)
	if err != nil {
		logs.Error("Failed to cleanup old data: %v", err)
	}
}

func UpdateUserSettings(retentionDays int) error {
	if retentionDays < DefaultDataRetentionDays {
		return fmt.Errorf("retention period must be at least %d days", DefaultDataRetentionDays)
	}
	userSettings.DataRetentionDays = retentionDays
	return nil
}

