package certificateutil

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

func NewCertificateFromPemContent(content []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(content)
	if block == nil || block.Bytes == nil || len(block.Bytes) == 0 {
		return nil, fmt.Errorf("failed to parse profile from: %s", string(content))
	}
	return NewCertificateFromDERContent(block.Bytes)
}

func NewCertificateFromDERContent(content []byte) (*x509.Certificate, error) {
	return x509.ParseCertificate(content)
}
