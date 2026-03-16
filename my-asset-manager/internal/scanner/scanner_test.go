package scanner

import (
	"my-asset-manager/internal/models"
	"testing"
)

func TestAllScanners(t *testing.T) {
	domainAsset := &models.Asset{Name: "google.com", Type: models.TypeDomain}
	
	(&DNSScanner{}).Run(domainAsset)
	
	(&PortScanner{}).Run(&models.Asset{Name: "127.0.0.1", Type: models.TypeIP})
	
	(&SSLScanner{}).Run(domainAsset)
	
	(&WhoisScanner{}).Run(&models.Asset{Name: "localhost", Type: models.TypeDomain})
    
	(&TechScanner{}).Run(domainAsset)

	(&IPScanner{}).Run(&models.Asset{Name: "127.0.0.1", Type: models.TypeIP})

	t.Run("SSL Scanner - TLS Versions", func(t *testing.T) {
		s := &SSLScanner{}
		s.Run(&models.Asset{Name: "google.com", Type: models.TypeDomain})
	})
}
