package profileutil

import (
	profileutilv1 "github.com/bitrise-io/go-xcode/profileutil"
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
		DeveloperCertificates: copySlice(model.DeveloperCertificates),
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
		DeveloperCertificates: copySlice(model.DeveloperCertificates),
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
