package scanner

import (
	"fmt"
	"my-asset-manager/internal/models"
	"time"

	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
)

type WhoisScanner struct{}

func (s *WhoisScanner) Run(asset *models.Asset) (interface{}, error) {
	raw, err := whois.Whois(asset.Name)
	if err != nil {
		return nil, fmt.Errorf("whois query failed: %v", err)
	}

	result, err := whoisparser.Parse(raw)
	if err != nil {
		return map[string]interface{}{
			"domain": asset.Name,
			"raw":    raw,
		}, nil
	}

	return map[string]interface{}{
		"domain":           asset.Name,
		"registrar":        result.Registrar.Name,
		"created_date":     result.Domain.CreatedDate,
		"expiration_date":  result.Domain.ExpirationDate,
		"name_servers":     result.Domain.NameServers,
		"status":           result.Domain.Status,
		"registrant_org":   result.Registrant.Organization,
		"last_updated":     time.Now().Format(time.RFC3339), // Chuẩn hóa thời gian
	}, nil
}
