// Copyright Authors of HActiV

// network package
package network

import (
	"fmt"
	"net"
)

func InetNtoa(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d", byte(ip), byte(ip>>8), byte(ip>>16), byte(ip>>24))
}

func IsDockerInternalIP(ip string) bool {
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

func IsMulticastIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.IsMulticast()
}
