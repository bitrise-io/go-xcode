package localcodesignasset

import (
	"github.com/bitrise-io/go-xcode/v2/autocodesign"
	"github.com/bitrise-io/go-xcode/v2/autocodesign/devportalclient/appstoreconnect"
)

func certificateSerials(certsByType map[appstoreconnect.CertificateType][]autocodesign.Certificate, distrType autocodesign.DistributionType) []string {
	certType := autocodesign.CertificateTypeByDistribution[distrType]
	certs := certsByType[certType]

	var serials []string
	for _, cert := range certs {
		serials = append(serials, cert.CertificateInfo.Serial)
	}

	return serials
}

func contains(array []string, element string) bool {
	for _, item := range array {
		if item == element {
			return true
		}
	}
	return false
}
