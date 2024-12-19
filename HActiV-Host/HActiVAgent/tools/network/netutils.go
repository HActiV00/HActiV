// Copyright Authors of HActiV

// network package
package network

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var (
	dockerSubnets    []string
	hostIP           string
	gatewayIP        string
	dnsServers       []string
	verifiedIPRanges = []string{"91.189.88.0/21"}
	ipInfoMap        = make(map[string]IPInfo)
	ipInfoFile       = "ip_info.json"
	ipInfoMutex      sync.RWMutex
)

const (
	cacheValidityPeriod = 24 * time.Hour
)

type IPInfo struct {
	Organization string    `json:"organization"`
	LastUpdated  time.Time `json:"last_updated"`
}

type GeoJSResponse struct {
	Organization string `json:"organization"`
	IP           string `json:"ip"`
}

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

func SaveIPInfoToFile() {
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

func GetServiceNameFromFile(ip string, ipType string) string {
	if IsMulticastIP(ip) {
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

	go SaveIPInfoToFile()

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

func DetectHostIPAndInterface() (string, string) {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Fatalf("Error getting network interfaces: %v", err)
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			log.Printf("Error getting addresses for interface %s: %v", iface.Name, err)
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String(), iface.Name
				}
			}
		}
	}

	log.Fatalf("Could not detect host IP and interface")
	return "", ""
}
