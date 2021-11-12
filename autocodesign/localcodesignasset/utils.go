package localcodesignasset

import (
	"github.com/bitrise-io/go-xcode/autocodesign"
	"github.com/bitrise-io/go-xcode/autocodesign/devportalclient/appstoreconnect"
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

func intersection(a, b []string) (c []string) {
	m := make(map[string]bool)

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		if _, ok := m[item]; ok {
			c = append(c, item)
		}
	}

	return
}

func remove(slice []string, i int) []string {
	copy(slice[i:], slice[i+1:])
	return slice[:len(slice)-1]
}

func contains(array []string, element string) bool {
	for _, item := range array {
		if item == element {
			return true
		}
	}
	return false
}
