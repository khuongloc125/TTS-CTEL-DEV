package scanner

import (
	"my-asset-manager/internal/models"
	"net/http"
	"time"
)

type TechScanner struct{}

func (s *TechScanner) Run(asset *models.Asset) (interface{}, error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	
	resp, err := client.Get("http://" + asset.Name)
	if err != nil {
		resp, err = client.Get("https://" + asset.Name)
		if err != nil {
			return nil, err
		}
	}
	defer resp.Body.Close()

	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	return map[string]interface{}{
		"domain":  asset.Name,
		"headers": headers,
	}, nil
}
