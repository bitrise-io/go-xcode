package certificateutil

import "sort"

// CertificateValidityGroups partitions certificate infos into valid, invalid and
// duplicated (same common name) groups.
type CertificateValidityGroups struct {
	ValidCertificates,
	InvalidCertificates,
	DuplicatedCertificates []CertificateInfoModel
}

// GroupCertificatesByValidity partitions the certificate infos into valid, invalid
// and duplicated common name groups.
func GroupCertificatesByValidity(certificateInfos []CertificateInfoModel) CertificateValidityGroups {
	var invalidCertificates []CertificateInfoModel
	nameToCerts := map[string][]CertificateInfoModel{}
	for _, certificateInfo := range certificateInfos {
		if certificateInfo.CheckValidity() != nil {
			invalidCertificates = append(invalidCertificates, certificateInfo)
			continue
		}

		nameToCerts[certificateInfo.CommonName] = append(nameToCerts[certificateInfo.CommonName], certificateInfo)
	}

	var validCertificates, duplicatedCertificates []CertificateInfoModel
	for _, certs := range nameToCerts {
		if len(certs) == 0 {
			continue
		}

		sort.Slice(certs, func(i, j int) bool {
			return certs[i].EndDate.After(certs[j].EndDate)
		})
		validCertificates = append(validCertificates, certs[0])
		if len(certs) > 1 {
			duplicatedCertificates = append(duplicatedCertificates, certs[1:]...)
		}
	}

	return CertificateValidityGroups{
		ValidCertificates:      validCertificates,
		InvalidCertificates:    invalidCertificates,
		DuplicatedCertificates: duplicatedCertificates,
	}
}
