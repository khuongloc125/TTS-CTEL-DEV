package scanner

import (
	"crypto/tls"
	"fmt"
	"my-asset-manager/internal/models"
	"time"
)

type SSLScanner struct{}


func (s *SSLScanner) Run(asset *models.Asset) (interface{}, error) {
	dialer := &tls.Dialer{
		Config: &tls.Config{
			InsecureSkipVerify: true, 
		},
	}

	conn, err := dialer.Dial("tcp", asset.Name+":443")
	if err != nil {
		return nil, fmt.Errorf("could not connect to port 443: %v", err)
	}
	defer conn.Close()

	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		return nil, fmt.Errorf("failed to assert tls.Conn")
	}

	state := tlsConn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return nil, fmt.Errorf("no certificates found")
	}

	cert := state.PeerCertificates[0]
	
	return map[string]interface{}{
		"domain": asset.Name,
		"certificate": map[string]interface{}{
			"subject":           cert.Subject.String(),
			"issuer":            cert.Issuer.String(),
			"serial_number":     fmt.Sprintf("%X", cert.SerialNumber),
			"valid_from":        cert.NotBefore,
			"valid_until":       cert.NotAfter,
			"days_until_expiry": int(time.Until(cert.NotAfter).Hours() / 24),
			"is_expired":        time.Now().After(cert.NotAfter),
			"is_self_signed":    cert.Issuer.String() == cert.Subject.String(),
		},
		"connection": map[string]interface{}{
			"tls_version": getTLSVersionName(state.Version),
		},
	}, nil
}

func getTLSVersionName(v uint16) string {
	switch v {
	case tls.VersionTLS10: return "TLS 1.0"
	case tls.VersionTLS11: return "TLS 1.1"
	case tls.VersionTLS12: return "TLS 1.2"
	case tls.VersionTLS13: return "TLS 1.3"
	default: return "Unknown"
	}
}
