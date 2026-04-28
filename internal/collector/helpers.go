package collector

import (
	"net"
)

// GetAllIPs returns all local IP addresses plus 0.0.0.0 for binding.
func GetAllIPs() ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	var ips []string
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok {
			ips = append(ips, ipNet.IP.String())
		}
	}
	ips = append(ips, "0.0.0.0")
	return ips, nil
}
