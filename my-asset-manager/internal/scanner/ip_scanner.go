package scanner

import (
	"encoding/json"
	"net/http"
	"my-asset-manager/internal/models"
)

type IPScanner struct{}

func (s *IPScanner) Run(asset *models.Asset) (interface{}, error) {
	
	resp, err := http.Get("http://ip-api.com/json/" + asset.Name + "?fields=status,message,country,countryCode,city,regionName,lat,lon,isp,org,as,reverse")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	return []map[string]interface{}{{
		"ip_address": asset.Name,
		"geolocation": map[string]interface{}{
			"country":      data["country"],
			"city":         data["city"],
			"isp":          data["isp"],
			"latitude":     data["lat"],
			"longitude":    data["lon"],
		},
		"asn": map[string]interface{}{
			"name": data["as"],
		},
		"reverse_dns": data["reverse"],
	}}, nil
}
