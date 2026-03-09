package profileutil

import (
	certificateutilv1 "github.com/bitrise-io/go-xcode/certificateutil"
	profileutilv1 "github.com/bitrise-io/go-xcode/profileutil"
	"github.com/bitrise-io/go-xcode/v2/certificateutil"
)

// V2Profile ...
func V2Profile(model profileutilv1.ProvisioningProfileInfoModel) ProvisioningProfileInfoModel {
	return ProvisioningProfileInfoModel{
		UUID:                  model.UUID,
		Name:                  model.Name,
		TeamName:              model.TeamName,
		TeamID:                model.TeamID,
		BundleID:              model.BundleID,
		ExportType:            model.ExportType,
		ProvisionedDevices:    copySlice(model.ProvisionedDevices),
		DeveloperCertificates: copyV1Certs(model.DeveloperCertificates),
		CreationDate:          model.CreationDate,
		ExpirationDate:        model.ExpirationDate,
		Entitlements:          copyMap(model.Entitlements),
		ProvisionsAllDevices:  model.ProvisionsAllDevices,
		Type:                  ProfileType(model.Type),
	}
}

// V1Profiles ...
func V1Profiles(models []ProvisioningProfileInfoModel) []profileutilv1.ProvisioningProfileInfoModel {
	profiles := make([]profileutilv1.ProvisioningProfileInfoModel, len(models))
	for i, model := range models {
		profiles[i] = V1Profile(model)
	}
	return profiles
}

// V1Profile ...
func V1Profile(model ProvisioningProfileInfoModel) profileutilv1.ProvisioningProfileInfoModel {
	return profileutilv1.ProvisioningProfileInfoModel{
		UUID:                  model.UUID,
		Name:                  model.Name,
		TeamName:              model.TeamName,
		TeamID:                model.TeamID,
		BundleID:              model.BundleID,
		ExportType:            model.ExportType,
		ProvisionedDevices:    copySlice(model.ProvisionedDevices),
		DeveloperCertificates: copyV2Certs(model.DeveloperCertificates),
		CreationDate:          model.CreationDate,
		ExpirationDate:        model.ExpirationDate,
		Entitlements:          copyMap(model.Entitlements),
		ProvisionsAllDevices:  model.ProvisionsAllDevices,
		Type:                  profileutilv1.ProfileType(model.Type),
	}
}

func copySlice[T any](src []T) []T {
	if src == nil {
		return nil
	}
	dst := make([]T, len(src))
	copy(dst, src)
	return dst
}

func copyMap[K comparable, V any](src map[K]V) map[K]V {
	if src == nil {
		return nil
	}
	dst := make(map[K]V, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func copyV1Certs(src []certificateutilv1.CertificateInfoModel) []certificateutil.CertificateInfoModel {
	if src == nil {
		return nil
	}
	dst := make([]certificateutil.CertificateInfoModel, len(src))
	for i, cert := range src {
		dst[i] = certificateutil.CertificateInfoModel(cert)
	}
	return dst
}

func copyV2Certs(src []certificateutil.CertificateInfoModel) []certificateutilv1.CertificateInfoModel {
	if src == nil {
		return nil
	}
	dst := make([]certificateutilv1.CertificateInfoModel, len(src))
	for i, cert := range src {
		dst[i] = certificateutilv1.CertificateInfoModel(cert)
	}
	return dst
}
