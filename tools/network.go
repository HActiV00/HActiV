package tools

import (
	"HActiV/pkg/docker"
	"HActiV/pkg/utils"
	bpfcode "HActiV/tools/bpf"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	bcc "github.com/iovisor/gobpf/bcc"
)

func NetworkMonitoring() {
	fmt.Println("Starting automatic container network traffic monitoring...")

	ContainerNamespaces := docker.GetContainer()

	if len(ContainerNamespaces) < 2 {
		fmt.Println("No containers to monitor. Exiting...")
		return
	}

	// Create a context to manage graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	go MonitorContainer(ctx, ContainerNamespaces)

	// Handle graceful shutdown on Ctrl+C
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Block until a shutdown signal is received
	<-shutdown
	fmt.Println("Shutdown signal received, stopping monitoring...")

	// Cancel the context to signal goroutines to stop
	cancel()

	fmt.Println("All monitoring stopped. Exiting...")
}

type Event struct {
	Pid         uint32
	SrcIP       uint32
	DstIP       uint32
	Protocol    uint8
	PacketCount uint64
}

// Docker 서브넷 및 호스트 IP 정보
var dockerSubnets = []string{
	"172.17.0.0/16", // Docker 기본 네트워크 서브넷
	// 추가적인 Docker 서브넷이 있으면 여기에 포함
}

var hostIP = "192.168.219.170" // 호스트 IP를 정확하게 설정

// handleEvent는 eBPF에서 캡처된 네트워크 트래픽 이벤트를 처리합니다.
func handleEvent(data unsafe.Pointer) {
	event := (*Event)(data)

	// SrcIP와 DstIP의 트래픽 유형을 분류 (Canonical 트래픽은 무시)
	srcType := ClassifyTraffic(InetNtoa(event.SrcIP), dockerSubnets, hostIP)
	dstType := ClassifyTraffic(InetNtoa(event.DstIP), dockerSubnets, hostIP)

	// srcType 또는 dstType이 빈 문자열(Canonical 트래픽)일 경우 무시
	if srcType == "" || dstType == "" {
		return // Canonical 트래픽은 무시
	}
	inode, err := utils.GetNamespaceInode(event.Pid)
	if err != nil {
		fmt.Printf("failed to get namespace for PID %d: %s\n", event.Pid, err)
		return
	}
	containerNamespaces := docker.GetContainer()

	containerInfo, exists := containerNamespaces[inode]
	if !exists {
		return
	}

	// 프로토콜 이름 가져오기
	protocolName := GetProtocolName(event.Protocol)
	utils.DataSend("Network_traffic", time.Now().Format(time.RFC3339), containerInfo.Name, InetNtoa(event.SrcIP), srcType, InetNtoa(event.DstIP), dstType, protocolName, int(event.PacketCount))

	// 분류된 트래픽 정보를 출력
	fmt.Printf("%s | %s SRC_IP: %s (%s), DST_IP: %s (%s), PROTOCOL: %s, PACKETS: %d\n", time.Now().Format(time.RFC3339), containerInfo.Name,
		InetNtoa(event.SrcIP), srcType, InetNtoa(event.DstIP), dstType, protocolName, event.PacketCount)
}

func MonitorContainer(ctx context.Context, Container map[uint64]utils.ContainerInfo) {

	bpfModule := utils.LoadBPFModule(bpfcode.NetworkCcode)
	defer bpfModule.Close()

	kprobe, err := bpfModule.LoadKprobe("kprobe__ip_rcv")
	if err != nil {
		log.Fatalf("Failed to load kprobe: %v", err)
	}
	err = bpfModule.AttachKprobe("ip_rcv", kprobe, -1)
	if err != nil {
		log.Fatalf("Failed to attach kprobe: %v", err)
	}

	channel := make(chan []byte)
	perfMap := initPerfMap(bpfModule, channel)
	go monitorTraffic(ctx, channel)
	perfMap.Start()
	defer perfMap.Stop()

	<-ctx.Done()
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

func monitorTraffic(ctx context.Context, channel chan []byte) {
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

// InetNtoa converts an IP address from its 32-bit integer representation to string format (e.g., "192.168.0.1").
func InetNtoa(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d", byte(ip), byte(ip>>8), byte(ip>>16), byte(ip>>24))
}

// GetProtocolName returns the protocol name (e.g., "TCP", "UDP") given its protocol number.
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

// ｖｅｒｉｆｅｄ ｉｄ ranges definition
var verifiedIPRanges = []string{
	"91.189.88.0/21", // Canonical's IP range
}

// IP 캐시와 뮤텍스를 사용해 동시성을 처리
var ipCache = make(map[string]string)
var mu sync.Mutex

// isVerifiedIP checks if an IP is part of the verified IP ranges
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

// ClassifyTraffic classifies the type of traffic based on IP addresses
func ClassifyTraffic(ip string, dockerSubnets []string, hostIP string) string {
	// Canonical IP 대역을 무시
	if isVerifiedIP(ip) {
		return "Ignore" // Canonical 트래픽은 무시
	}

	// IP가 Docker 내부 네트워크에 속하는지 확인
	if IsDockerInternalIP(ip, dockerSubnets) {
		return "Docker internal" // Docker 내부 트래픽

		// IP가 로컬 네트워크에 속하는지 확인
	} else if IsLocalNetworkIP(ip) {
		return "Local Network"

		// IP가 호스트 IP와 일치하는지 확인
	} else if ip == hostIP {
		return "Host internal"
	}

	// 그 외의 트래픽은 외부로 간주
	serviceName := GetServiceNameFromWhois(ip, "External")
	if serviceName != "" {
		return "External (" + serviceName + ")"
	}
	return "External"
}

// IsDockerInternalIP checks if the IP is within Docker's internal subnets
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

// IsLocalNetworkIP checks if the IP belongs to the local network
func IsLocalNetworkIP(ip string) bool {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.Contains(net.ParseIP(ip)) {
				return true
			}
		}
	}
	return false
}

// GetServiceNameFromWhois uses the whois command to retrieve the service name for an IP, with caching and error handling
func GetServiceNameFromWhois(ip string, ipType string) string {
	// IP가 멀티캐스트 주소인지 확인
	if isMulticastIP(ip) {
		return "Multicast"
	}

	// IP가 외부인지 확인. 그렇지 않으면 whois 명령을 건너뜀
	if ipType != "External" {
		return ""
	}

	// IP가 캐시에 있는지 확인
	mu.Lock()
	if service, exists := ipCache[ip]; exists {
		mu.Unlock()
		return service
	}
	mu.Unlock()

	// whois 명령 실행
	cmd := exec.Command("whois", ip)
	output, err := cmd.Output()
	if err != nil {
		// 에러가 발생하면 "Unknown" 반환
		return "Unknown"
	}

	// whois 출력에서 서비스/조직 이름을 추출
	serviceName := parseWhoisOutput(string(output))
	if serviceName == "" {
		// 서비스 이름을 찾지 못하면 "Unknown" 반환
		return "Unknown"
	}

	// 결과를 캐시에 저장
	mu.Lock()
	ipCache[ip] = serviceName
	mu.Unlock()

	return serviceName
}

// parseWhoisOutput extracts the service or organization name from whois output
func parseWhoisOutput(output string) string {
	// 관심 있는 필드를 위한 정규 표현식 정의
	orgNameRegex := regexp.MustCompile(`(?i)OrgName:\s*(.*)`)
	netNameRegex := regexp.MustCompile(`(?i)NetName:\s*(.*)`)
	organizationRegex := regexp.MustCompile(`(?i)Organization:\s*(.*)`)

	// 첫 번째로 일치하는 필드를 찾아 반환
	if match := orgNameRegex.FindStringSubmatch(output); len(match) > 1 {
		return strings.TrimSpace(match[1])
	} else if match := netNameRegex.FindStringSubmatch(output); len(match) > 1 {
		return strings.TrimSpace(match[1])
	} else if match := organizationRegex.FindStringSubmatch(output); len(match) > 1 {
		return strings.TrimSpace(match[1])
	}

	// 필드를 찾지 못한 경우 빈 문자열 반환
	return ""
}

// isMulticastIP checks if the IP belongs to the multicast range (224.0.0.0/4)
func isMulticastIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.IsMulticast()
}

// IsSpecialIP checks if the IP is either a broadcast or multicast address
func IsSpecialIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// IP가 멀티캐스트인지 확인
	if parsedIP.IsMulticast() {
		return true
	}

	// IP가 브로드캐스트 주소인지 확인 (IPv4: 서브넷 내 마지막 주소)
	if isBroadcastIP(parsedIP) {
		return true
	}

	return false
}

// isBroadcastIP는 IP가 브로드캐스트 IP인지 확인
func isBroadcastIP(ip net.IP) bool {
	if ip.To4() == nil {
		return false // IPv4 주소가 아님
	}
	// IPv4에서 브로드캐스트 IP는 서브넷의 마지막 주소
	return ip.Equal(net.IPv4bcast) || ip.Equal(net.ParseIP("255.255.255.255"))
}
