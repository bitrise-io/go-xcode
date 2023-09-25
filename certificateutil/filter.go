package certificateutil

import (
	"sort"

	"github.com/bitrise-io/go-xcode/v2/certificate"
)

func FilterCertificateInfoModelsByFilterFunc(certificates []certificate.Certificate, filterFunc func(certificate certificate.Certificate) bool) []certificate.Certificate {
	var filteredCertificates []certificate.Certificate

	for _, cert := range certificates {
		if filterFunc(cert) {
			filteredCertificates = append(filteredCertificates, cert)
		}
	}

	return filteredCertificates
}

type FilterValidResult struct {
	ValidCertificates,
	InvalidCertificates,
	DuplicatedCertificates []certificate.Certificate
}

func FilterValidCertificateInfos(certificateInfos []certificate.Certificate) FilterValidResult {
	var invalidCertificates []certificate.Certificate
	nameToCerts := map[string][]certificate.Certificate{}
	for _, certificateInfo := range certificateInfos {
		if certificateInfo.CheckValidity() != nil {
			invalidCertificates = append(invalidCertificates, certificateInfo)
			continue
		}

		certDetails := certificateInfo.Details()
		nameToCerts[certDetails.CommonName] = append(nameToCerts[certDetails.CommonName], certificateInfo)
	}

	var validCertificates, duplicatedCertificates []certificate.Certificate
	for _, certs := range nameToCerts {
		if len(certs) == 0 {
			continue
		}

		sort.Slice(certs, func(i, j int) bool {
			return certs[i].X509Certificate.NotAfter.Before(certs[j].X509Certificate.NotAfter)
		})
		validCertificates = append(validCertificates, certs[0])
		if len(certs) > 1 {
			duplicatedCertificates = append(duplicatedCertificates, certs[1:]...)
		}
	}

	return FilterValidResult{
		ValidCertificates:      validCertificates,
		InvalidCertificates:    invalidCertificates,
		DuplicatedCertificates: duplicatedCertificates,
	}
}
