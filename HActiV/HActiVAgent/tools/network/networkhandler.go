// Copyright Authors of HActiV

// network package
package network

import (
	"HActiV/configs"
	"HActiV/pkg/docker"
	"HActiV/pkg/utils"
	"encoding/json"
	"log"
	"net"
	"sync"
	"time"
	"unsafe"

	bcc "github.com/iovisor/gobpf/bcc"
)

type Event struct {
	Pid         uint32
	SrcIP       uint32
	DstIP       uint32
	Protocol    uint8
	PacketCount uint64
	MntNs       uint32
	PacketSize  uint32
	DstPort     uint16
	IsOutgoing  bool
}

type ContainerStats struct {
	PacketCount         uint64
	TotalSize           uint64
	IncomingPacketCount uint64
	IncomingTotalSize   uint64
	OutgoingPacketCount uint64
	OutgoingTotalSize   uint64
}

var (
	containerStats      = make(map[string]*ContainerStats)
	containerStatsMutex sync.RWMutex
)

func HandleEvent(data unsafe.Pointer, policies []configs.Policy, logger *utils.DualLogger) {
	event := (*Event)(data)

	direction := "incoming"
	if event.IsOutgoing {
		direction = "outgoing"
	}

	srcIP := InetNtoa(event.SrcIP)
	dstIP := InetNtoa(event.DstIP)
	protocolName := GetProtocolName(event.Protocol)

	if protocolName == "TCP" && event.DstPort == 80 {
		httpEvents := GetHTTPEvents()

		for _, httpEvent := range httpEvents {
			if httpEvent.SrcIP == srcIP && httpEvent.DstIP == dstIP {
				HandleCombinedEvent(httpEvent, event, event.MntNs, logger)
				return
			}
		}
	}

	srcType := ClassifyTraffic(srcIP)
	dstType := ClassifyTraffic(dstIP)

	if srcType == "" || dstType == "" {
		return
	}

	containerNamespaces := docker.GetContainer()
	containerInfo, exists := containerNamespaces[uint64(event.MntNs)]
	if !exists {
		if utils.HostMonitoring {
			containerInfo.Name = "H"
		} else {
			return
		}
	}


	updateContainerStats(containerInfo.Name, event.PacketSize, event.IsOutgoing)
	//matchevent Tool network -> Network_traffic 수정 Datasend와 일치 시키기 위해

	path := generateTrafficPath(srcIP, srcType, dstIP, dstType)
	pathJSON, err := json.Marshal(path)
	if err != nil {
		log.Printf("Error marshaling path to JSON: %v", err)
		return
	}

	containerStatsMutex.RLock()
	stats := containerStats[containerInfo.Name]
	containerStatsMutex.RUnlock()

	matchevent := utils.Event{
		Tool:          "Network_traffic",
		Time:          time.Now().Format(time.RFC3339),
		ContainerName: containerInfo.Name,
		SrcIp:         srcIP,
		SrcIpLabel:    srcType,
		DstIp:         dstIP,
		DstIpLabel:    dstType,
		Protocol:      protocolName,
		PacketCount:   int(stats.PacketCount),
		Direction:     direction,
		PacketSize:    int(event.PacketSize),
		PathJson:      string(pathJSON),
		TotalSize:     int(stats.TotalSize),
	}

	configs.MatchedEvent(policies, matchevent)

	logger.Log(matchevent)
	/*
		fmt.Printf("%s | CONTAINER: %s | PATH: %s | SRC_IP: %s (%s) | DST_IP: %s (%s) | PROTOCOL: %s | PACKETS: %d | SIZE: %d bytes | TOTAL_PACKETS: %d | TOTAL_SIZE: %d bytes\n",
			time.Now().Format(time.RFC3339),
			containerInfo.Name,
			string(pathJSON),
			srcIP,
			srcType,
			dstIP,
			dstType,
			protocolName,
			event.PacketCount,
			event.PacketSize,
			stats.PacketCount,
			stats.TotalSize,
		)
	*/
	if configs.DataSend {
		utils.DataSend(
			"Network_traffic",
			matchevent.Time,
			containerInfo.Name,
			srcIP,
			srcType,
			dstIP,
			dstType,
			protocolName,
			int(event.PacketSize),
			int(stats.PacketCount),
			int(stats.TotalSize),
			string(pathJSON),
			direction,
			"", "", "", "",
		)
	}
}

func HandleCombinedEvent(httpEvent HTTPEvent, networkEvent *Event, mntNs uint32, logger *utils.DualLogger) {
	formattedTimestamp := time.Now().UTC().Format(time.RFC3339)

	srcType := ClassifyTraffic(httpEvent.SrcIP)
	dstType := ClassifyTraffic(httpEvent.DstIP)

	if srcType == "" || dstType == "" {
		return
	}

	containerNamespaces := docker.GetContainer()
	containerInfo, exists := containerNamespaces[uint64(mntNs)]
	containerName := "Unknown"
	if exists {
		containerName = containerInfo.Name
	} else if utils.HostMonitoring{
		containerName = "H"
	}

	if containerName == "Unknown" {
		return
	}

	containerStatsMutex.RLock()
	stats, statsExists := containerStats[containerName]
	containerStatsMutex.RUnlock()

	if !statsExists {
		log.Printf("Container stats not found for %s", containerName)
		return
	}

	direction := "incoming"
	if networkEvent.IsOutgoing {
		direction = "outgoing"
	}

	pathJSON, err := generateTrafficPathJSON(httpEvent.SrcIP, srcType, httpEvent.DstIP, dstType)
	if err != nil {
		log.Printf("Error generating traffic path JSON: %v", err)
		return
	}

	//matchevent Tool network -> Network_traffic 수정 Datasend와 일치 시키기 위해
	matchevent := utils.Event{
		Tool:          "Network_traffic",
		Time:          time.Now().Format(time.RFC3339),
		ContainerName: containerInfo.Name,
		SrcIp:         httpEvent.SrcIP,
		SrcIpLabel:    srcType,
		DstIp:         httpEvent.DstIP,
		DstIpLabel:    dstType,
		Direction:     direction,
		PacketSize:    int(networkEvent.PacketSize),
		PacketCount:   int(stats.PacketCount),
		TotalSize:     int(stats.TotalSize),
		PathJson:      pathJSON,
		Method:        httpEvent.Method,
		Host:          httpEvent.Host,
		URL:           httpEvent.URL,
		Parameters:    httpEvent.Parameters,
	}
	logger.Log(matchevent)

	/*
		fmt.Printf(
			"%s | CONTAINER: %s | HTTP | SRC_IP: %s (%s) | DST_IP: %s (%s) | METHOD: %s | HOST: %s | URL: %s | PARAMETERS: %s | PROTOCOL: %s | SIZE: %d bytes | TOTAL_PACKETS: %d | TOTAL_SIZE: %d bytes | PATH: %s | DIRECTION: %s\n",
			formattedTimestamp,
			containerName,
			httpEvent.SrcIP, srcType,
			httpEvent.DstIP, dstType,
			httpEvent.Method, httpEvent.Host, httpEvent.URL, httpEvent.Parameters,
			GetProtocolName(uint8(networkEvent.Protocol)),
			networkEvent.PacketSize,
			stats.PacketCount,
			stats.TotalSize,
			pathJSON,
			direction,
		)
	*/

	if configs.DataSend {
		utils.DataSend(
			"Network_traffic",
			formattedTimestamp,
			containerName,
			httpEvent.SrcIP,
			srcType,
			httpEvent.DstIP,
			dstType,
			GetProtocolName(uint8(networkEvent.Protocol)),
			int(networkEvent.PacketSize),
			int(stats.PacketCount),
			int(stats.TotalSize),
			pathJSON,
			direction,
			httpEvent.Method,
			httpEvent.Host,
			httpEvent.URL,
			httpEvent.Parameters,
		)
	}
}

func HandleHTTPEvent(event HTTPEvent, mntNs uint32) {
	srcType := ClassifyTraffic(event.SrcIP)
	dstType := ClassifyTraffic(event.DstIP)

	if srcType == "" || dstType == "" {
		return
	}

	containerNamespaces := docker.GetContainer()
	_, exists := containerNamespaces[uint64(mntNs)]
	if !exists {
		return
	}

	_, err := GenerateTrafficPathJSON(event.SrcIP, srcType, event.DstIP, dstType)
	if err != nil {
		log.Printf("Error generating traffic path JSON: %v", err)
		return
	}
}

func generateTrafficPathJSON(srcIP, srcType, dstIP, dstType string) (string, error) {
	path := utils.Path{
		Nodes: []utils.Node{
			{ID: srcIP, Type: srcType},
			{ID: dstIP, Type: dstType},
		},
		Links: []utils.Link{
			{Source: srcIP, Target: dstIP},
		},
	}

	pathJSON, err := json.Marshal(path)
	if err != nil {
		return "", err
	}
	return string(pathJSON), nil
}

func updateContainerStats(containerName string, packetSize uint32, isOutgoing bool) {
	containerStatsMutex.Lock()
	defer containerStatsMutex.Unlock()

	if stats, exists := containerStats[containerName]; exists {
		stats.PacketCount++
		stats.TotalSize += uint64(packetSize)
		if isOutgoing {
			stats.OutgoingPacketCount++
			stats.OutgoingTotalSize += uint64(packetSize)
		} else {
			stats.IncomingPacketCount++
			stats.IncomingTotalSize += uint64(packetSize)
		}
	} else {
		stats := &ContainerStats{
			PacketCount: 1,
			TotalSize:   uint64(packetSize),
		}
		if isOutgoing {
			stats.OutgoingPacketCount = 1
			stats.OutgoingTotalSize = uint64(packetSize)
		} else {
			stats.IncomingPacketCount = 1
			stats.IncomingTotalSize = uint64(packetSize)
		}
		containerStats[containerName] = stats
	}
}

func InitPerfMap(module *bcc.Module, channel chan []byte) *bcc.PerfMap {
	table := bcc.NewTable(module.TableId("events"), module)
	lost := make(chan uint64)
	//perfMap, err := bcc.InitPerfMap(table, channel, lost)
	perfMap, err := bcc.InitPerfMapWithPageCnt(table, channel, lost, 512)
	if err != nil {
		log.Fatalf("Failed to init perf map: %v", err)
	}
	return perfMap
}

func GetProtocolName(protocol uint8) string {
	switch protocol {
	case 1:
		return "ICMP"
	case 6:
		return "TCP"
	case 17:
		return "UDP"
	default:
		return "UNKNOWN"
	}
}

func ClassifyTraffic(ip string) string {
	if IsVerifiedIP(ip) {
		return "Ignore"
	}

	if IsDockerInternalIP(ip) {
		return "Docker internal"
	} else if ip == hostIP {
		return "Host internal"
	} else if ip == gatewayIP {
		return "Gateway"
	} else if IsDNSServer(ip) {
		return "DNS"
	} else if IsLocalNetworkIP(ip) {
		return "Local Network"
	}

	serviceName := GetServiceNameFromFile(ip, "External")
	if serviceName != "" {
		return "External (" + serviceName + ")"
	}
	return "External"
}

func IsDNSServer(ip string) bool {
	for _, dnsIP := range dnsServers {
		if ip == dnsIP {
			return true
		}
	}
	return false
}

func IsVerifiedIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	for _, cidr := range verifiedIPRanges {
		_, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}

		if ipnet.Contains(parsedIP) {
			return true
		}
	}
	return false
}
