package scanner

import (
	"my-asset-manager/internal/models"
	"net"
)

type DNSScanner struct{}

func (s *DNSScanner) Run(asset *models.Asset) (interface{}, error) {
	ips, _ := net.LookupIP(asset.Name)
	var ipStrings []string
	for _, ip := range ips {
		ipStrings = append(ipStrings, ip.String())
	}

	mxs, _ := net.LookupMX(asset.Name)
	
	return map[string]interface{}{
		"domain":     asset.Name,
		"a_records":  ipStrings,
		"mx_records": mxs,
	}, nil
}
