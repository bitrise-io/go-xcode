package localcodesignasset

import (
	"testing"

	"github.com/bitrise-io/go-xcode/autocodesign"
	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/bitrise-io/go-xcode/exportoptions"
	"github.com/bitrise-io/go-xcode/profileutil"
	"github.com/stretchr/testify/assert"
)

type fields struct {
	platform            autocodesign.Platform
	distributionType    autocodesign.DistributionType
	bundleID            string
	entitlements        autocodesign.Entitlements
	minProfileDaysValid int
	certIDs             []string
	deviceIDs           []string
}

func Test_GivenProfiles_WhenSearchingForAnExisting_ThenFindsIt(t *testing.T) {
	// Given
	iosDevelopmentProfile := iosDevelopmentProfile(t)
	iosAppStoreProfile := iosAppStoreProfile(t)
	tvosAdHocProfile := tvosAdHocProfile(t)
	tvosEnterpriseProfile := tvosEnterpriseProfile(t)
	profiles := []profileutil.ProvisioningProfileInfoModel{iosDevelopmentProfile, iosAppStoreProfile, tvosAdHocProfile, tvosEnterpriseProfile}
	tests := []struct {
		name            string
		fields          fields
		expectedProfile profileutil.ProvisioningProfileInfoModel
	}{
		{
			name: "iOS development profile",
			fields: fields{
				platform:            autocodesign.IOS,
				distributionType:    autocodesign.Development,
				bundleID:            "io.ios",
				entitlements:        firstSetOfEntitlements(),
				minProfileDaysValid: 0,
				certIDs:             []string{"1"},
				deviceIDs:           []string{"ios-device-1"},
			},
			expectedProfile: iosDevelopmentProfile,
		},
		{
			name: "iOS app store profile",
			fields: fields{
				platform:            autocodesign.IOS,
				distributionType:    autocodesign.AppStore,
				bundleID:            "io.ios",
				entitlements:        nil,
				minProfileDaysValid: 3,
				certIDs:             []string{"2"},
				deviceIDs:           []string{"ios-device-3"},
			},
			expectedProfile: iosAppStoreProfile,
		},
		{
			name: "tvOS ad hoc profile",
			fields: fields{
				platform:            autocodesign.TVOS,
				distributionType:    autocodesign.AdHoc,
				bundleID:            "io.tvos",
				entitlements:        nil,
				minProfileDaysValid: 5,
				certIDs:             []string{"2"},
				deviceIDs:           nil,
			},
			expectedProfile: tvosAdHocProfile,
		},
		{
			name: "tvOS enterprise profile",
			fields: fields{
				platform:            autocodesign.TVOS,
				distributionType:    autocodesign.Enterprise,
				bundleID:            "io.tvos",
				entitlements:        nil,
				minProfileDaysValid: 9,
				certIDs:             []string{"2"},
				deviceIDs:           nil,
			},
			expectedProfile: tvosEnterpriseProfile,
		},
	}

	for _, test := range tests {
		// When
		profile := findProfile(profiles, test.fields.platform, test.fields.distributionType, test.fields.bundleID, test.fields.entitlements, test.fields.minProfileDaysValid, test.fields.certIDs, test.fields.deviceIDs)

		// Then
		assert.Equal(t, test.expectedProfile, *profile)
	}
}

func Test_GivenProfiles_WhenFiltersForNonExisting_ThenItIsMissing(t *testing.T) {
	// Given
	profiles := []profileutil.ProvisioningProfileInfoModel{iosDevelopmentProfile(t), iosAppStoreProfile(t), tvosAdHocProfile(t), tvosEnterpriseProfile(t), iosXcodeManagedDevelopmentProfile(t)}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "iOS development profile with missing device id",
			fields: fields{
				platform:            autocodesign.IOS,
				distributionType:    autocodesign.Development,
				bundleID:            "io.ios",
				entitlements:        nil,
				minProfileDaysValid: 0,
				certIDs:             []string{"1"},
				deviceIDs:           []string{"ios-device-99"},
			},
		},
		{
			name: "iOS development profile with missing entitlement",
			fields: fields{
				platform:         autocodesign.IOS,
				distributionType: autocodesign.Development,
				bundleID:         "io.ios",
				entitlements: map[string]interface{}{
					"non-existing": "missing",
				},
				minProfileDaysValid: 0,
				certIDs:             []string{"1"},
				deviceIDs:           nil,
			},
		},
		{
			name: "tvOS ad hoc expired profile",
			fields: fields{
				platform:            autocodesign.TVOS,
				distributionType:    autocodesign.AdHoc,
				bundleID:            "io.tvos",
				entitlements:        nil,
				minProfileDaysValid: 14,
				certIDs:             []string{"2"},
				deviceIDs:           nil,
			},
		},
		{
			name: "tvOS non existing distribution type",
			fields: fields{
				platform:            autocodesign.TVOS,
				distributionType:    autocodesign.Development,
				bundleID:            "io.tvos",
				entitlements:        nil,
				minProfileDaysValid: 9,
				certIDs:             []string{"2"},
				deviceIDs:           nil,
			},
		},
		{
			name: "iOS app store missing certificate",
			fields: fields{
				platform:            autocodesign.IOS,
				distributionType:    autocodesign.AppStore,
				bundleID:            "io.ios",
				entitlements:        nil,
				minProfileDaysValid: 3,
				certIDs:             []string{"100"},
				deviceIDs:           []string{"ios-device-3"},
			},
		},
		{
			name: "iOS app store 1 included, 1 missing certificate",
			fields: fields{
				platform:            autocodesign.IOS,
				distributionType:    autocodesign.AppStore,
				bundleID:            "io.ios",
				entitlements:        nil,
				minProfileDaysValid: 3,
				certIDs:             []string{"2", "100"},
				deviceIDs:           []string{"ios-device-3"},
			},
		},
		{
			name: "iOS Xcode-amanged development profile",
			fields: fields{
				platform:         autocodesign.IOS,
				distributionType: autocodesign.Development,
				bundleID:         "io.ios.managed",
			},
		},
	}

	for _, test := range tests {
		// When
		profile := findProfile(profiles, test.fields.platform, test.fields.distributionType, test.fields.bundleID, test.fields.entitlements, test.fields.minProfileDaysValid, test.fields.certIDs, test.fields.deviceIDs)

		// Then
		assert.Nil(t, profile)
	}
}

// Helpers

func iosXcodeManagedDevelopmentProfile(t *testing.T) profileutil.ProvisioningProfileInfoModel {
	return profileutil.ProvisioningProfileInfoModel{
		UUID:                  "uuid-1",
		Name:                  "iOS Team Provisioning Profile: io.ios.managed",
		TeamName:              "TeamName",
		TeamID:                "TeamID",
		BundleID:              "io.ios.managed",
		ExportType:            exportoptions.MethodDevelopment,
		ProvisionedDevices:    []string{"ios-device-1", "ios-device-2", "ios-device-3"},
		DeveloperCertificates: []certificateutil.CertificateInfoModel{devCert(t, dateRelativeToNow(1, 0, 0))},
		CreationDate:          dateRelativeToNow(0, 0, -1),
		ExpirationDate:        dateRelativeToNow(0, 0, 5),
		Entitlements:          firstSetOfEntitlements(),
		ProvisionsAllDevices:  false,
		Type:                  profileutil.ProfileTypeIos,
	}
}

func iosDevelopmentProfile(t *testing.T) profileutil.ProvisioningProfileInfoModel {
	return profileutil.ProvisioningProfileInfoModel{
		UUID:                  "uuid-1",
		Name:                  "iOS development profile",
		TeamName:              "TeamName",
		TeamID:                "TeamID",
		BundleID:              "io.ios",
		ExportType:            exportoptions.MethodDevelopment,
		ProvisionedDevices:    []string{"ios-device-1", "ios-device-2", "ios-device-3"},
		DeveloperCertificates: []certificateutil.CertificateInfoModel{devCert(t, dateRelativeToNow(1, 0, 0))},
		CreationDate:          dateRelativeToNow(0, 0, -1),
		ExpirationDate:        dateRelativeToNow(0, 0, 5),
		Entitlements:          firstSetOfEntitlements(),
		ProvisionsAllDevices:  false,
		Type:                  profileutil.ProfileTypeIos,
	}
}

func iosAppStoreProfile(t *testing.T) profileutil.ProvisioningProfileInfoModel {
	return profileutil.ProvisioningProfileInfoModel{
		UUID:                  "uuid-2",
		Name:                  "iOS app store profile",
		TeamName:              "TeamName",
		TeamID:                "TeamID",
		BundleID:              "io.ios",
		ExportType:            exportoptions.MethodAppStore,
		ProvisionedDevices:    nil,
		DeveloperCertificates: []certificateutil.CertificateInfoModel{distCert(t, dateRelativeToNow(1, 0, 0))},
		CreationDate:          dateRelativeToNow(0, 0, -1),
		ExpirationDate:        dateRelativeToNow(0, 0, 5),
		Entitlements:          nil,
		ProvisionsAllDevices:  true,
		Type:                  profileutil.ProfileTypeIos,
	}
}

func tvosAdHocProfile(t *testing.T) profileutil.ProvisioningProfileInfoModel {
	return profileutil.ProvisioningProfileInfoModel{
		UUID:                  "uuid-3",
		Name:                  "tvOS ad hoc profile",
		TeamName:              "TeamName",
		TeamID:                "TeamID",
		BundleID:              "io.tvos",
		ExportType:            exportoptions.MethodAdHoc,
		ProvisionedDevices:    []string{"tvos-device-1", "tvos-device-2", "tvos-device-3"},
		DeveloperCertificates: []certificateutil.CertificateInfoModel{distCert(t, dateRelativeToNow(1, 0, 0))},
		CreationDate:          dateRelativeToNow(0, 0, -1),
		ExpirationDate:        dateRelativeToNow(0, 0, 10),
		Entitlements:          nil,
		ProvisionsAllDevices:  false,
		Type:                  profileutil.ProfileTypeTvOs,
	}
}

func tvosEnterpriseProfile(t *testing.T) profileutil.ProvisioningProfileInfoModel {
	return profileutil.ProvisioningProfileInfoModel{
		UUID:                  "uuid-4",
		Name:                  "tvOS enterprise profile",
		TeamName:              "TeamName",
		TeamID:                "TeamID",
		BundleID:              "io.tvos",
		ExportType:            exportoptions.MethodEnterprise,
		ProvisionedDevices:    nil,
		DeveloperCertificates: []certificateutil.CertificateInfoModel{distCert(t, dateRelativeToNow(1, 0, 0))},
		CreationDate:          dateRelativeToNow(0, 0, -1),
		ExpirationDate:        dateRelativeToNow(0, 0, 10),
		Entitlements:          secondSetOfEntitlements(),
		ProvisionsAllDevices:  true,
		Type:                  profileutil.ProfileTypeTvOs,
	}
}

func firstSetOfEntitlements() map[string]interface{} {
	return map[string]interface{}{
		"first": []interface{}{
			"first-value",
		},
		"second": "second-value",
	}
}

func secondSetOfEntitlements() map[string]interface{} {
	return map[string]interface{}{
		"third": "third-value",
	}
}
