// Copyright Authors of HActiV

// network package
package network

import (
	"HActiV/pkg/docker"
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type HTTPEvent struct {
	Method     string
	Host       string
	URL        string
	Parameters string
	StatusCode int
	SrcIP      string
	DstIP      string
	Timestamp  string
	Network    string
}

var (
	httpEvents      = make([]HTTPEvent, 0)
	httpEventsMutex sync.Mutex
)

func StartHTTPMonitor(device string, bpfFilter string) {
	if device == "" {
		activeInterfaces, err := GetActiveNetworkInterfaces()
		if err != nil {
			log.Fatalf("활성화된 네트워크 인터페이스를 감지하지 못했습니다: %v", err)
		}

		for _, iface := range activeInterfaces {
			log.Printf("감지된 네트워크 인터페이스: %s", iface.Name)
			startMonitorForInterface(iface.Name, bpfFilter)
		}
		return
	}

	startMonitorForInterface(device, bpfFilter)
}

func startMonitorForInterface(device string, bpfFilter string) {
	handle, err := pcap.OpenLive(device, 65536, true, pcap.BlockForever)
	if err != nil {
		log.Fatalf("Error opening device %s: %v", device, err)
	}
	defer handle.Close()

	if bpfFilter == "" {
		bpfFilter = "tcp port 80"
	}

	if err := handle.SetBPFFilter(bpfFilter); err != nil {
		log.Fatalf("Error setting BPF filter: %v", err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	log.Printf("HTTP monitor started on interface %s, capturing HTTP/80 traffic...", device)
	for packet := range packetSource.Packets() {
		processPacket(packet, 0)
	}
}

func processPacket(packet gopacket.Packet, mntNs uint32) {
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if ipLayer == nil || tcpLayer == nil {
		return
	}

	ip, _ := ipLayer.(*layers.IPv4)
	tcp, _ := tcpLayer.(*layers.TCP)

	containerNamespaces := docker.GetContainer()

	for inode, containerInfo := range containerNamespaces {
		if containerInfo.ID != "" {
			if ip.SrcIP.String() == containerInfo.Name {
				mntNs = uint32(inode)
				break
			}
		}
	}

	if !tcp.SYN && tcp.ACK {
		applicationLayer := packet.ApplicationLayer()
		if applicationLayer != nil {
			payload := string(applicationLayer.Payload())
			if strings.Contains(payload, "HTTP/1.1") || strings.Contains(payload, "HTTP/2") {
				parseHTTPRequest(payload, ip, tcp, mntNs)
			}
		}
	}
}

func parseHTTPRequest(payload string, ip *layers.IPv4, tcp *layers.TCP, mntNs uint32) {
	reader := bufio.NewReader(bytes.NewReader([]byte(payload)))

	request, err := http.ReadRequest(reader)
	if err != nil {
		return
	}
	defer request.Body.Close()

	networkType := "Host"
	activeInterfaces, err := GetActiveNetworkInterfaces()
	if err != nil {
		log.Printf("활성화된 네트워크 인터페이스 확인 실패: %v", err)
	} else {
		for _, iface := range activeInterfaces {
			if isDockerNetwork(iface) {
				networkType = "Docker"
				break
			}
		}
	}

	event := HTTPEvent{
		Method:     request.Method,
		Host:       request.Host,
		URL:        request.URL.String(),
		Parameters: request.URL.RawQuery,
		SrcIP:      ip.SrcIP.String(),
		DstIP:      ip.DstIP.String(),
		Timestamp:  fmt.Sprintf("%d", tcp.Seq),
		Network:    networkType,
	}

	HandleHTTPEvent(event, mntNs)

	httpEventsMutex.Lock()
	httpEvents = append(httpEvents, event)
	httpEventsMutex.Unlock()
}

func GetHTTPEvents() []HTTPEvent {
	httpEventsMutex.Lock()
	defer httpEventsMutex.Unlock()

	events := make([]HTTPEvent, len(httpEvents))
	copy(events, httpEvents)
	return events
}

func GetActiveNetworkInterfaces() ([]net.Interface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("네트워크 인터페이스 가져오기 실패: %v", err)
	}

	var activeInterfaces []net.Interface
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp != 0 {
			if isDockerNetwork(iface) {
				activeInterfaces = append(activeInterfaces, iface)
			}
		}
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp != 0 && !isDockerNetwork(iface) {
			activeInterfaces = append(activeInterfaces, iface)
		}
	}
	return activeInterfaces, nil
}

func isDockerNetwork(iface net.Interface) bool {
	addrs, _ := iface.Addrs()
	for _, addr := range addrs {
		ip := strings.Split(addr.String(), "/")[0]
		if IsDockerInternalIP(ip) {
			return true
		}
	}
	return false
}
