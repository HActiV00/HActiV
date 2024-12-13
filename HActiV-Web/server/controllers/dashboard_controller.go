package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"server/models"
	"time"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/gorilla/websocket"
)

type DashboardController struct {
	beego.Controller
}

type InputData struct {
	EventType     string      `json:"event_type"`
	Time          time.Time   `json:"timestamp"`
	ContainerName string      `json:"container_name"`
	Uid           uint32      `json:"uid"`
	Gid           uint32      `json:"gid"`
	Pid           uint32      `json:"pid"`
	Ppid          uint32      `json:"ppid"`
	ProcessName   string      `json:"process_name"`
	Filename      string      `json:"filename,omitempty"`
	Args          string      `json:"arguments,omitempty"`
	ReturnValue   int32       `json:"status,omitempty"`
	SrcIp         string      `json:"src_ip,omitempty"`
	SrcIpLabel    string      `json:"src_ip_label,omitempty"`
	DstIp         string      `json:"dst_ip,omitempty"`
	DstIpLabel    string      `json:"dst_ip_label,omitempty"`
	Protocol      string      `json:"protocol,omitempty"`
	Packets       int         `json:"packets,omitempty"`
	Size          int         `json:"size,omitempty"`
	TotalPackets  int         `json:"total_packets,omitempty"`
	TotalSize     int         `json:"total_size,omitempty"`
	Path          models.Path `json:"path,omitempty"`
	StartAddress  uint64      `json:"start_address,omitempty"`
	EndAddress    uint64      `json:"end_address,omitempty"`
	Type          string      `json:"type,omitempty"`
	Cpu           float32     `json:"cpu,omitempty"`
	Core          int32       `json:"core,omitempty"`
	Memory        float32     `json:"memory,omitempty"`
	Disk          float32     `json:"disk,omitempty"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	clients    = make(map[*websocket.Conn]bool)
	broadcast  = make(chan []byte)
	register   = make(chan *websocket.Conn)
	unregister = make(chan *websocket.Conn)
)

func init() {
	go handleMessages()
}

func handleMessages() {
	for {
		select {
		case conn := <-register:
			clients[conn] = true
		case conn := <-unregister:
			if _, ok := clients[conn]; ok {
				delete(clients, conn)
				conn.Close()
			}
		case message := <-broadcast:
			for conn := range clients {
				err := conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					logs.Error("Error broadcasting message: %v", err)
					conn.Close()
					delete(clients, conn)
				}
			}
		}
	}
}

func (c *DashboardController) WebSocketHandler() {
	ws, err := upgrader.Upgrade(c.Ctx.ResponseWriter, c.Ctx.Request, nil)
	if err != nil {
		logs.Error("Failed to set WebSocket upgrade: %v", err)
		return
	}

	register <- ws

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			logs.Error("WebSocket read error: %v", err)
			unregister <- ws
			break
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

	case "file_open":
		var data models.OpenApiData
		if err := mapToStruct(inputData, &data); err != nil {
			c.handleError(400, "Invalid data for file_open event", err)
			return
		}
		if err := models.SaveOpenData(&data); err != nil {
			c.handleError(500, "Failed to save open data", err)
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

	case "delete":
		var data models.DeleteApiData
		if err := mapToStruct(inputData, &data); err != nil {
			c.handleError(400, "Invalid data for delete event", err)
			return
		}
		if err := models.SaveDeleteData(&data); err != nil {
			c.handleError(500, "Failed to save delete data", err)
			return
		}

	case "log_file_access":
		var data models.LogAccessApiData
		if err := mapToStruct(inputData, &data); err != nil {
			c.handleError(400, "Invalid data for log_file_access event", err)
			return
		}
		if err := models.SaveLogAccessData(&data); err != nil {
			c.handleError(500, "Failed to save log access data", err)
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

	broadcast <- body

	c.Data["json"] = map[string]string{"message": "Data received and saved successfully"}
	c.ServeJSON()
}

func (c *DashboardController) Get() {
	eventType := c.GetString("event_type", "all")
	logs.Debug("Requested event type: %s", eventType)

	data, err := models.GetDashboardData(eventType)
	if err != nil {
		logs.Error("Failed to retrieve data: %v", err)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	c.Data["json"] = data
	c.ServeJSON()
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

