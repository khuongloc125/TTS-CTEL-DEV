package scanner

import (
	"fmt"
	"my-asset-manager/internal/models"
	"net"
	"time"
)

type PortScanner struct{}

func (s *PortScanner) Run(asset *models.Asset) (interface{}, error) {
    targetIP := asset.Name
    
    ips, err := net.LookupIP(asset.Name)
    if err == nil && len(ips) > 0 {
        targetIP = ips[0].String()
    }

    if !isPrivateIP(targetIP) {
        return nil, fmt.Errorf("security violation: port scan only allowed on private network (%s)", targetIP)
    }
	
	commonPorts := []int{22, 80, 443, 3306, 5432}
	var openPorts []map[string]interface{}
	
	for _, port := range commonPorts {
		address := fmt.Sprintf("%s:%d", asset.Name, port)
		conn, err := net.DialTimeout("tcp", address, 500*time.Millisecond)
		if err == nil {
			openPorts = append(openPorts, map[string]interface{}{
				"port":     port,
				"protocol": "tcp",
				"state":    "open",
			})
			conn.Close()
		}
	}

	return []map[string]interface{}{{
		"ip_address": asset.Name,
		"open_ports": openPorts,
		"total_scanned": len(commonPorts),
	}}, nil
}

func isPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil { return false }
	return ip.IsPrivate() || ip.IsLoopback()
}
