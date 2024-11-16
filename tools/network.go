// Copyright Authors of HActiV

package tools

import (
	"HActiV/configs"
	"HActiV/pkg/docker"
	"HActiV/pkg/utils"
	bpfcode "HActiV/tools/bpf"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
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
}

type IPInfo struct {
	Organization string    `json:"organization"`
	LastUpdated  time.Time `json:"last_updated"`
}

type GeoJSResponse struct {
	Organization string `json:"organization"`
	IP           string `json:"ip"`
}

type PathNode struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type PathLink struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type TrafficPath struct {
	Nodes []PathNode `json:"nodes"`
	Links []PathLink `json:"links"`
}

type ContainerStats struct {
	PacketCount uint64
	TotalSize   uint64
}

var (
	dockerSubnets            []string
	hostIP                   string
	gatewayIP                string
	dnsServers               []string
	verifiedIPRanges         = []string{"91.189.88.0/21"}
	ipInfoMap                = make(map[string]IPInfo)
	ipInfoFile               = "ip_info.json"
	ipInfoMutex              sync.RWMutex
	monitoredContainers      = make(map[uint64]context.CancelFunc)
	monitoredContainersMutex sync.Mutex
	containerStats           = make(map[string]*ContainerStats)
	containerStatsMutex      sync.RWMutex
)

const (
	cacheValidityPeriod = 24 * time.Hour
)

func init() {
	loadIPInfoFromFile()
	detectNetworkInfo()
}

func loadIPInfoFromFile() {
	data, err := ioutil.ReadFile(ipInfoFile)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Error reading IP info file: %v", err)
		}
		return
	}

	ipInfoMutex.Lock()
	defer ipInfoMutex.Unlock()

	err = json.Unmarshal(data, &ipInfoMap)
	if err != nil {
		log.Printf("Error unmarshaling IP info: %v", err)
	}
}

func saveIPInfoToFile() {
	ipInfoMutex.RLock()
	defer ipInfoMutex.RUnlock()

	data, err := json.MarshalIndent(ipInfoMap, "", "  ")
	if err != nil {
		log.Printf("Error marshaling IP info: %v", err)
		return
	}

	err = ioutil.WriteFile(ipInfoFile, data, 0644)
	if err != nil {
		log.Printf("Error writing IP info file: %v", err)
	}
}

func detectNetworkInfo() {
	detectHostIP()
	detectDockerSubnets()
	detectGatewayIP()
	detectDNSServers()
}

func detectHostIP() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatalf("Error getting interface addresses: %v", err)
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				hostIP = ipnet.IP.String()
				log.Printf("Detected host IP: %s", hostIP)
				return
			}
		}
	}

	log.Fatalf("Could not detect host IP")
}

func detectDockerSubnets() {
	cmd := exec.Command("docker", "network", "ls", "--format", "{{.Name}}")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error listing Docker networks: %v", err)
		return
	}

	networks := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, network := range networks {
		if network == "bridge" {
			inspectCmd := exec.Command("docker", "network", "inspect", network, "--format", "{{range .IPAM.Config}}{{.Subnet}}{{end}}")
			subnetOutput, err := inspectCmd.Output()
			if err != nil {
				log.Printf("Error inspecting Docker network %s: %v", network, err)
				continue
			}
			subnet := strings.TrimSpace(string(subnetOutput))
			if subnet != "" {
				dockerSubnets = append(dockerSubnets, subnet)
			}
		}
	}

	log.Printf("Detected Docker subnets: %v", dockerSubnets)
}

func detectGatewayIP() {
	cmd := exec.Command("ip", "route", "show", "default")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error detecting gateway IP: %v", err)
		return
	}

	fields := strings.Fields(string(output))
	if len(fields) > 2 {
		gatewayIP = fields[2]
		log.Printf("Detected gateway IP: %s", gatewayIP)
	}
}

func detectDNSServers() {
	content, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		log.Printf("Error reading /etc/resolv.conf: %v", err)
		return
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 1 && fields[0] == "nameserver" {
			dnsServers = append(dnsServers, fields[1])
		}
	}

	log.Printf("Detected DNS servers: %v", dnsServers)
}

func NetworkMonitoring() {
	fmt.Println("Starting network monitoring. Waiting for containers...")
	fmt.Println("Starting automatic container network traffic monitoring...")

	// Network 규칙 초기화 및 설정
	if err := configs.SetupRules("network"); err != nil {
		log.Fatalf("Failed to setup network rules: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go monitorTraffic(ctx)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown
	fmt.Println("Shutdown signal received, stopping monitoring...")
	cancel()
	fmt.Println("All monitoring stopped. Exiting...")
	saveIPInfoToFile()
}

// func monitorContainers(ctx context.Context) {
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return
// 		default:
// 			containerNamespaces := docker.GetContainer()
// 			if len(containerNamespaces) == 0 {
// 				log.Println("No containers detected. Continuing to monitor...")
// 				time.Sleep(5 * time.Second)
// 				continue
// 			}

// 			monitoredContainersMutex.Lock()
// 			// Start monitoring for new containers
// 			for namespace, containerInfo := range containerNamespaces {
// 				if _, exists := monitoredContainers[namespace]; !exists {
// 					containerCtx, containerCancel := context.WithCancel(ctx)
// 					monitoredContainers[namespace] = containerCancel
// 					go MonitorContainer(containerCtx, containerInfo)
// 				}
// 			}

// 			// Stop monitoring for removed containers
// 			for namespace, cancelFunc := range monitoredContainers {
// 				if _, exists := containerNamespaces[namespace]; !exists {
// 					cancelFunc()
// 					delete(monitoredContainers, namespace)
// 				}
// 			}
// 			monitoredContainersMutex.Unlock()

// 			time.Sleep(5 * time.Second)
// 		}
// 	}
// }

func handleEvent(data unsafe.Pointer) {
	event := (*Event)(data)

	srcIP := InetNtoa(event.SrcIP)
	dstIP := InetNtoa(event.DstIP)
	srcType := ClassifyTraffic(srcIP, dockerSubnets, hostIP)
	dstType := ClassifyTraffic(dstIP, dockerSubnets, hostIP)

	if srcType == "" || dstType == "" {
		return
	}

	containerNamespaces := docker.GetContainer()
	containerInfo, exists := containerNamespaces[uint64(event.MntNs)]
	if !exists {
		return
	}

	protocolName := GetProtocolName(event.Protocol)

	// 이벤트 데이터 생성
	eventData := map[string]interface{}{
		"src_ip":       srcIP,
		"dst_ip":       dstIP,
		"protocol":     protocolName,
		"packet_count": int(event.PacketCount),
	}

	// 규칙 적용을 위한 매칭 검사
	if !configs.ParseRules("Network_traffic", eventData, "network") {
		// "ignore" action인 경우 출력하지 않음
		return
	}

	path := generateTrafficPath(srcIP, srcType, dstIP, dstType)
	pathJSON, err := json.Marshal(path)
	if err != nil {
		log.Printf("Error marshaling path to JSON: %v", err)
		return
	}

	updateContainerStats(containerInfo.Name, event.PacketSize)

	containerStatsMutex.RLock()
	stats := containerStats[containerInfo.Name]
	containerStatsMutex.RUnlock()

	utils.DataSend("Network_traffic", time.Now().Format(time.RFC3339), containerInfo.Name, srcIP, srcType, dstIP, dstType, protocolName, int(event.PacketCount), string(pathJSON))

	fmt.Printf("%s | %s PATH: %s, SRC_IP: %s (%s), DST_IP: %s (%s), PROTOCOL: %s, PACKETS: %d, SIZE: %d bytes, TOTAL_PACKETS: %d, TOTAL_SIZE: %d bytes\n",
		time.Now().Format(time.RFC3339), containerInfo.Name, string(pathJSON),
		srcIP, srcType, dstIP, dstType, protocolName, event.PacketCount, event.PacketSize,
		stats.PacketCount, stats.TotalSize)
}

func updateContainerStats(containerName string, packetSize uint32) {
	containerStatsMutex.Lock()
	defer containerStatsMutex.Unlock()

	if stats, exists := containerStats[containerName]; exists {
		stats.PacketCount++
		stats.TotalSize += uint64(packetSize)
	} else {
		containerStats[containerName] = &ContainerStats{
			PacketCount: 1,
			TotalSize:   uint64(packetSize),
		}
	}
}

func MonitorContainer(ctx context.Context) {

	<-ctx.Done()
	// log.Printf("Stopping monitoring for container: %s", container.ID)
}

func initPerfMap(module *bcc.Module, channel chan []byte) *bcc.PerfMap {
	table := bcc.NewTable(module.TableId("events"), module)
	lost := make(chan uint64)
	perfMap, err := bcc.InitPerfMap(table, channel, lost)
	if err != nil {
		log.Fatalf("Failed to init perf map: %v", err)
	}
	return perfMap
}

func monitorTraffic(ctx context.Context) {
	bpfModule := utils.LoadBPFModule(bpfcode.NetworkCcode)
	defer bpfModule.Close()

	kprobercv, err := bpfModule.LoadKprobe("kprobe__ip_rcv")
	if err != nil {
		log.Fatalf("Failed to load kprobe: %v", err)
	}
	err = bpfModule.AttachKprobe("ip_rcv", kprobercv, -1)
	if err != nil {
		log.Fatalf("Failed to attach kprobe: %v", err)
	}
	// kprobeout, err := bpfModule.LoadKprobe("kprobe__ip_output")
	// if err != nil {
	// 	log.Fatalf("Failed to load kprobe: %v", err)
	// }
	// err = bpfModule.AttachKprobe("ip_output", kprobeout, -1)
	// if err != nil {
	// 	log.Fatalf("Failed to attach kprobe: %v", err)
	// }

	channel := make(chan []byte)
	perfMap := initPerfMap(bpfModule, channel)
	perfMap.Start()
	fmt.Println("Network Monitoring Start")
	defer perfMap.Stop()
	for {
		select {
		case data := <-channel:
			handleEvent(unsafe.Pointer(&data[0]))
		case <-ctx.Done():
			log.Println("Stopping traffic monitoring")
			return
		}
	}
}

func InetNtoa(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d", byte(ip), byte(ip>>8), byte(ip>>16), byte(ip>>24))
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

func isVerifiedIP(ip string) bool {
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

func ClassifyTraffic(ip string, dockerSubnets []string, hostIP string) string {
	if isVerifiedIP(ip) {
		return "Ignore"
	}

	if IsDockerInternalIP(ip, dockerSubnets) {
		return "Docker internal"
	} else if ip == hostIP {
		return "Host internal"
	} else if ip == gatewayIP {
		return "Gateway"
	} else if isDNSServer(ip) {
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

func isDNSServer(ip string) bool {
	for _, dnsIP := range dnsServers {
		if ip == dnsIP {
			return true
		}
	}
	return false
}

func IsDockerInternalIP(ip string, dockerSubnets []string) bool {
	parsedIP := net.ParseIP(ip)
	for _, subnet := range dockerSubnets {
		_, ipNet, _ := net.ParseCIDR(subnet)
		if ipNet.Contains(parsedIP) {
			return true
		}
	}
	return false
}

func IsLocalNetworkIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	if parsedIP.IsLoopback() || parsedIP.IsLinkLocalUnicast() || parsedIP.IsLinkLocalMulticast() {
		return true
	}

	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.Contains(parsedIP) {
				return true
			}
		}
	}
	return false
}

func GetServiceNameFromFile(ip string, ipType string) string {
	if isMulticastIP(ip) {
		return "Multicast"
	}

	if ipType != "External" {
		return ""
	}

	ipInfoMutex.RLock()
	info, exists := ipInfoMap[ip]
	ipInfoMutex.RUnlock()

	if exists && time.Since(info.LastUpdated) < cacheValidityPeriod {
		return info.Organization
	}

	serviceName := fetchIPInfo(ip)
	if serviceName == "" {
		serviceName = "Unknown"
	}

	ipInfoMutex.Lock()
	ipInfoMap[ip] = IPInfo{
		Organization: serviceName,
		LastUpdated:  time.Now(),
	}
	ipInfoMutex.Unlock()

	go saveIPInfoToFile()

	return serviceName
}

func fetchIPInfo(ip string) string {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(fmt.Sprintf("https://get.geojs.io/v1/ip/geo/%s.json", ip))
	if err != nil {
		log.Printf("Error fetching IP info: %v", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return ""
	}

	var geoInfo GeoJSResponse
	err = json.Unmarshal(body, &geoInfo)
	if err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		return ""
	}

	if geoInfo.Organization != "" {
		return geoInfo.Organization
	}

	return "Unknown"
}

func isMulticastIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.IsMulticast()
}

func IsSpecialIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	if parsedIP.IsMulticast() {
		return true
	}

	if isBroadcastIP(parsedIP) {
		return true
	}

	return false
}

func isBroadcastIP(ip net.IP) bool {
	if ip.To4() == nil {
		return false
	}
	return ip.Equal(net.IPv4bcast) || ip.Equal(net.ParseIP("255.255.255.255"))
}

func generateTrafficPath(srcIP, srcType, dstIP, dstType string) TrafficPath {
	path := TrafficPath{}

	// Add source node
	path.Nodes = append(path.Nodes, PathNode{ID: srcIP, Type: srcType})

	// Add intermediate nodes if necessary
	if srcType == "Docker internal" && dstType == "External" {
		path.Nodes = append(path.Nodes,
			PathNode{ID: hostIP, Type: "Host internal"},
			PathNode{ID: gatewayIP, Type: "Gateway"},
			PathNode{ID: dnsServers[0], Type: "DNS"})
	}

	// Add destination node
	path.Nodes = append(path.Nodes, PathNode{ID: dstIP, Type: dstType})

	// Create links
	for i := 0; i < len(path.Nodes)-1; i++ {
		path.Links = append(path.Links, PathLink{
			Source: path.Nodes[i].ID,
			Target: path.Nodes[i+1].ID,
		})
	}

	return path
}
