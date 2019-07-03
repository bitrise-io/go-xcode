package certificateutil

// FilterCertificateInfoModelsByFilterFunc ...
func FilterCertificateInfoModelsByFilterFunc(certificates []CertificateInfoModel, filterFunc func(certificate CertificateInfoModel) bool) []CertificateInfoModel {
	filteredCertificates := []CertificateInfoModel{}

	for _, certificate := range certificates {
		if filterFunc(certificate) {
			filteredCertificates = append(filteredCertificates, certificate)
		}
	}

	return filteredCertificates
}

// ValidCertificateInfo contains the certificate infos filtered as valid, invalid and duplicated common name certificates
type ValidCertificateInfo struct {
	validCertificates,
	invalidCertificates,
	duplicatedCertificates []CertificateInfoModel
}

// FilterValidCertificateInfos filters out invalid and duplicated common name certificaates
func FilterValidCertificateInfos(certificateInfos []CertificateInfoModel) ValidCertificateInfo {
	certificateInfosByName := map[string]CertificateInfoModel{}

	var invalidCertificates, duplicatedCertificates []CertificateInfoModel
	for _, certificateInfo := range certificateInfos {
		if certificateInfo.CheckValidity() != nil {
			invalidCertificates = append(invalidCertificates, certificateInfo)
			continue
		}
		activeCertificate, ok := certificateInfosByName[certificateInfo.CommonName]
		if !ok {
			certificateInfosByName[certificateInfo.CommonName] = certificateInfo
		} else if certificateInfo.EndDate.After(activeCertificate.EndDate) {
			duplicatedCertificates = append(duplicatedCertificates, activeCertificate)
			certificateInfosByName[certificateInfo.CommonName] = certificateInfo
		}
	}

	validCertificates := []CertificateInfoModel{}
	for _, validCertificate := range certificateInfosByName {
		validCertificates = append(validCertificates, validCertificate)
	}

	return ValidCertificateInfo{
		validCertificates:      validCertificates,
		invalidCertificates:    invalidCertificates,
		duplicatedCertificates: duplicatedCertificates,
	}
}
