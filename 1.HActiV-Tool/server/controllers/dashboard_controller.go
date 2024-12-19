package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"server/models"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/gorilla/websocket"
)

type DashboardController struct {
	beego.Controller
}

type InputData struct {
	EventType     string    `json:"event_type"`
	Time          time.Time `json:"timestamp"`
	ContainerName string    `json:"container_name"`
	Uid           uint32    `json:"uid"`
	Gid           uint32    `json:"gid"`
	Pid           uint32    `json:"pid"`
	Ppid          uint32    `json:"ppid"`
	ProcessName   string    `json:"process_name"`
	Filename      string    `json:"filename,omitempty"`
	Args          string    `json:"arguments,omitempty"`
	ReturnValue   int32     `json:"status,omitempty"`
	SrcIp         string    `json:"src_ip,omitempty"`
	SrcIpLabel    string    `json:"src_ip_label,omitempty"`
	DstIp         string    `json:"dst_ip,omitempty"`
	DstIpLabel    string    `json:"dst_ip_label,omitempty"`
	Protocol      string    `json:"protocol,omitempty"`
	Packets       int       `json:"packets,omitempty"`
	Size          int       `json:"size,omitempty"`
	TotalPackets  int       `json:"total_packets,omitempty"`
	TotalSize     int       `json:"total_size,omitempty"`
	Path          string    `json:"path,omitempty"`
	StartAddress  uint64    `json:"start_address,omitempty"`
	EndAddress    uint64    `json:"end_address,omitempty"`
	Type          string    `json:"type,omitempty"`
	Cpu           float32   `json:"cpu,omitempty"`
	Core          int32     `json:"core,omitempty"`
	Memory        float32   `json:"memory,omitempty"`
	Disk          float32   `json:"disk,omitempty"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (c *DashboardController) WebSocketHandler() {
	ws, err := upgrader.Upgrade(c.Ctx.ResponseWriter, c.Ctx.Request, nil)
	if err != nil {
		logs.Error("Failed to set WebSocket upgrade: %v", err)
		return
	}
	defer ws.Close()

	messageChan := models.GetMessageChannel()

	for {
		select {
		case msg := <-messageChan:
			err = ws.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				logs.Error("Failed to write WebSocket message: %v", err)
				return
			}
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (c *DashboardController) Post() {
	body, err := ioutil.ReadAll(c.Ctx.Request.Body)
	if err != nil {
		logs.Error("Failed to read request body: %v", err)
		c.handleError(400, "Failed to read request body", err)
		return
	}

	logs.Info("Raw POST body: %s", string(body))

	if len(body) == 0 {
		logs.Error("Empty request body")
		c.handleError(400, "Empty request body", nil)
		return
	}

	var inputData map[string]interface{}
	if err := json.Unmarshal(body, &inputData); err != nil {
		logs.Error("Failed to parse JSON: %v", err)
		c.handleError(400, "Invalid JSON data", err)
		return
	}

	logs.Info("Parsed data: %+v", inputData)

	eventType, ok := inputData["event_type"].(string)
	if !ok {
		c.handleError(400, "Missing or invalid event_type", nil)
		return
	}

	switch eventType {
	case "Systemcall":
		var data models.ExecveApiData
		if err := mapToStruct(inputData, &data); err != nil {
			c.handleError(400, "Invalid data for Systemcall event", err)
			return
		}
		if err := models.SaveExecveData(&data); err != nil {
			c.handleError(500, "Failed to save execve data", err)
			return
		}

	case "file_open", "delete":
		var data models.OpenApiData
		if err := mapToStruct(inputData, &data); err != nil {
			c.handleError(400, "Invalid data for file event", err)
			return
		}

		// 파일 이름에 'log'가 포함되어 있는지 확인
		if strings.Contains(strings.ToLower(data.Filename), "log") {
			if data.EventType == "file_open" {
				data.EventType = "log_file_open"
			} else if data.EventType == "delete" {
				data.EventType = "log_file_delete"
			}
		}

		if err := models.SaveOpenData(&data); err != nil {
			c.handleError(500, "Failed to save file event data", err)
			return
		}

	case "Network_traffic":
		var data models.NetworkApiData
		if err := mapToStruct(inputData, &data); err != nil {
			c.handleError(400, "Invalid data for Network_traffic event", err)
			return
		}
		if err := models.SaveNetworkData(&data); err != nil {
			c.handleError(500, "Failed to save network data", err)
			return
		}

	case "Memory":
		var data models.MemoryApiData
		if err := mapToStruct(inputData, &data); err != nil {
			c.handleError(400, "Invalid data for Memory event", err)
			return
		}
		if err := models.SaveMemoryData(&data); err != nil {
			c.handleError(500, "Failed to save memory data", err)
			return
		}


	case "ContainerMetrics":
		var data models.ContainerMetricsApiData
		if err := mapToStruct(inputData, &data); err != nil {
			c.handleError(400, "Invalid data for ContainerMetrics event", err)
			return
		}
		if err := models.SaveContainerMetricsData(&data); err != nil {
			c.handleError(500, "Failed to save container metrics data", err)
			return
		}

	case "HostMetrics":
		var data models.HostMetricsApiData
		if err := mapToStruct(inputData, &data); err != nil {
			c.handleError(400, "Invalid data for HostMetrics event", err)
			return
		}
		if err := models.SaveHostMetricsData(&data); err != nil {
			c.handleError(500, "Failed to save host metrics data", err)
			return
		}

	default:
		c.handleError(400, fmt.Sprintf("Unsupported event type: %s", eventType), nil)
		return
	}

	c.Data["json"] = map[string]string{"message": "Data received and saved successfully"}
	c.ServeJSON()
}

func (c *DashboardController) Get() {
	eventType := c.GetString("event_type", "all")
	startTimeStr := c.GetString("start_time")
	endTimeStr := c.GetString("end_time")

	var startTime, endTime time.Time
	var err error

	if startTimeStr != "" {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			c.handleError(400, "Invalid start_time format", err)
			return
		}
	}

	if endTimeStr != "" {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			c.handleError(400, "Invalid end_time format", err)
			return
		}
	}

	var data []interface{}
	if eventType == "log_file_open" || eventType == "log_file_delete" || eventType == "file_open" || eventType == "delete" {
		data, err = models.GetDashboardData(eventType, startTime, endTime)
		if err != nil {
			c.handleError(500, fmt.Sprintf("Failed to retrieve %s data", eventType), err)
			return
		}
	} else if eventType == "file_event" {
		fileOpenData, err := models.GetDashboardData("file_open", startTime, endTime)
		if err != nil {
			c.handleError(500, "Failed to retrieve file_open data", err)
			return
		}
		deleteData, err := models.GetDashboardData("delete", startTime, endTime)
		if err != nil {
			c.handleError(500, "Failed to retrieve delete data", err)
			return
		}
		logFileOpenData, err := models.GetDashboardData("log_file_open", startTime, endTime)
		if err != nil {
			c.handleError(500, "Failed to retrieve log_file_open data", err)
			return
		}
		logFileDeleteData, err := models.GetDashboardData("log_file_delete", startTime, endTime)
		if err != nil {
			c.handleError(500, "Failed to retrieve log_file_delete data", err)
			return
		}
		data = append(fileOpenData, deleteData...)
		data = append(data, logFileOpenData...)
		data = append(data, logFileDeleteData...)
	} else {
		data, err = models.GetDashboardData(eventType, startTime, endTime)
		if err != nil {
			logs.Error("Failed to retrieve data: %v", err)
			c.Data["json"] = map[string]string{"error": err.Error()}
			c.ServeJSON()
			return
		}
	}

	c.Data["json"] = data
	c.ServeJSON()
}

func (c *DashboardController) GetHistorical() {
	eventType := c.GetString("event_type", "all")
	startTimeStr := c.GetString("start_time")
	endTimeStr := c.GetString("end_time")

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		c.handleError(400, "Invalid start_time format", err)
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		c.handleError(400, "Invalid end_time format", err)
		return
	}

	var data []interface{}
	if eventType == "log_file_open" || eventType == "log_file_delete" || eventType == "file_open" || eventType == "delete" {
		data, err = models.GetDashboardData(eventType, startTime, endTime)
		if err != nil {
			c.handleError(500, fmt.Sprintf("Failed to retrieve historical %s data", eventType), err)
			return
		}
	} else if eventType == "file_event" {
		fileOpenData, err := models.GetDashboardData("file_open", startTime, endTime)
		if err != nil {
			c.handleError(500, "Failed to retrieve file_open data", err)
			return
		}
		deleteData, err := models.GetDashboardData("delete", startTime, endTime)
		if err != nil {
			c.handleError(500, "Failed to retrieve delete data", err)
			return
		}
		logFileOpenData, err := models.GetDashboardData("log_file_open", startTime, endTime)
		if err != nil {
			c.handleError(500, "Failed to retrieve log_file_open data", err)
			return
		}
		logFileDeleteData, err := models.GetDashboardData("log_file_delete", startTime, endTime)
		if err != nil {
			c.handleError(500, "Failed to retrieve log_file_delete data", err)
			return
		}
		data = append(fileOpenData, deleteData...)
		data = append(data, logFileOpenData...)
		data = append(data, logFileDeleteData...)
	} else {
		data, err = models.GetDashboardData(eventType, startTime, endTime)
		if err != nil {
			c.handleError(500, "Failed to retrieve historical data", err)
			return
		}
	}

	c.Data["json"] = data
	c.ServeJSON()
}

func (c *DashboardController) DeleteContainer() {
	containerName := c.Ctx.Input.Param(":container")

	err := models.DeleteContainerData(containerName)
	if err != nil {
		c.handleError(500, "Failed to delete container", err)
		return
	}

	c.Data["json"] = map[string]string{"message": fmt.Sprintf("Container %s deleted successfully", containerName)}
	c.ServeJSON()
}

func (c *DashboardController) SaveContainerData() {
	containerName := c.Ctx.Input.Param(":container")

	data, err := models.GetContainerEvents(containerName)
	if err != nil {
		c.handleError(500, "Failed to retrieve container data", err)
		return
	}

	c.Ctx.Output.Header("Content-Type", "text/csv")
	c.Ctx.Output.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s_events.csv", containerName))

	c.Ctx.Output.Body([]byte(data))
}

func (c *DashboardController) handleError(status int, message string, err error) {
	if err != nil {
		logs.Error("%s: %v", message, err)
	} else {
		logs.Error("%s", message)
	}
	c.Ctx.Output.SetStatus(status)
	c.Data["json"] = map[string]string{"error": message, "details": err.Error()}
	c.ServeJSON()
}

func mapToStruct(input map[string]interface{}, output interface{}) error {
	jsonData, err := json.Marshal(input)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, output)
}

