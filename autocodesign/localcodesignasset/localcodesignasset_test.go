package localcodesignasset

import (
	"testing"
	"time"

	devportaltime "github.com/bitrise-io/go-xcode/v2/autocodesign/devportalclient/time"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/go-xcode/v2/autocodesign"
	"github.com/bitrise-io/go-xcode/v2/autocodesign/devportalclient/appstoreconnect"
	"github.com/bitrise-io/go-xcode/v2/autocodesign/localcodesignasset/mocks"
	"github.com/bitrise-io/go-xcode/v2/certificateutil"
	"github.com/bitrise-io/go-xcode/v2/exportoptions"
	"github.com/bitrise-io/go-xcode/v2/profileutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	teamID   = "team-id"
	teamName = "Testing team"
)

func Test_GiveniOSAppLayoutWithUITestTargets_WhenExistingProfile_ThenFindsIt(t *testing.T) {
	// Given
	certsByType := certsByType(t)
	manager, profiles := createTestObjects(t)

	appLayout := autocodesign.AppLayout{
		Platform: autocodesign.IOS,
		EntitlementsByArchivableTargetBundleID: map[string]autocodesign.Entitlements{
			"io.ios.valid": entitlements(),
		},
		UITestTargetBundleIDs: []string{"io.ios.valid", "io.ios.valid"},
	}

	expectedAssets := autocodesign.AppCodesignAssets{
		ArchivableTargetProfilesByBundleID: map[string]autocodesign.Profile{
			"io.ios.valid": findProvProfile(t, profiles, "uuid-1"),
		},
		UITestTargetProfilesByBundleID: map[string]autocodesign.Profile{
			"io.ios.valid": findProvProfile(t, profiles, "uuid-4"),
		},
		Certificate: findCert(t, certsByType, "1"),
	}

	// When
	assets, missingAppLayout, err := manager.FindCodesignAssets(appLayout, autocodesign.Development, certsByType, []string{}, 0)

	// Then
	assert.NoError(t, err)
	assert.Nil(t, missingAppLayout)
	assert.Equal(t, expectedAssets, *assets)
}

func Test_GiveniOSAppLayoutWithEntitlements_WhenExistingProfile_ThenFindsIt(t *testing.T) {
	// Given
	certsByType := certsByType(t)
	manager, profiles := createTestObjects(t)

	appLayout := autocodesign.AppLayout{
		Platform: autocodesign.IOS,
		EntitlementsByArchivableTargetBundleID: map[string]autocodesign.Entitlements{
			"io.ios.valid": entitlements(),
		},
	}

	expectedAssets := autocodesign.AppCodesignAssets{
		ArchivableTargetProfilesByBundleID: map[string]autocodesign.Profile{
			"io.ios.valid": findProvProfile(t, profiles, "uuid-1"),
		},
		UITestTargetProfilesByBundleID: nil,
		Certificate:                    findCert(t, certsByType, "1"),
	}

	// When
	assets, missingAppLayout, err := manager.FindCodesignAssets(appLayout, autocodesign.Development, certsByType, []string{}, 0)

	// Then
	assert.NoError(t, err)
	assert.Nil(t, missingAppLayout)
	assert.Equal(t, expectedAssets, *assets)
}

func Test_GiventvOSAppLayout_WhenExistingProfile_ThenFindsIt(t *testing.T) {
	// Given
	certsByType := certsByType(t)
	manager, profiles := createTestObjects(t)

	appLayout := autocodesign.AppLayout{
		Platform: autocodesign.TVOS,
		EntitlementsByArchivableTargetBundleID: map[string]autocodesign.Entitlements{
			"io.tvos.valid": nil,
		},
	}

	expectedAssets := autocodesign.AppCodesignAssets{
		ArchivableTargetProfilesByBundleID: map[string]autocodesign.Profile{
			"io.tvos.valid": findProvProfile(t, profiles, "uuid-2"),
		},
		UITestTargetProfilesByBundleID: nil,
		Certificate:                    findCert(t, certsByType, "2"),
	}

	// When
	assets, missingAppLayout, err := manager.FindCodesignAssets(appLayout, autocodesign.AppStore, certsByType, []string{}, 0)

	// Then
	assert.NoError(t, err)
	assert.Nil(t, missingAppLayout)
	assert.Equal(t, expectedAssets, *assets)
}

func Test_GiveniOSAppLayout_WhenExpiredProfile_ThenDoesNotFindIt(t *testing.T) {
	// Given
	certsByType := certsByType(t)
	manager, _ := createTestObjects(t)

	appLayout := autocodesign.AppLayout{
		Platform: autocodesign.TVOS,
		EntitlementsByArchivableTargetBundleID: map[string]autocodesign.Entitlements{
			"io.tvos.expired": nil,
		},
	}

	// When
	assets, missingAppLayout, err := manager.FindCodesignAssets(appLayout, autocodesign.AppStore, certsByType, []string{}, 0)

	// Then
	assert.NoError(t, err)
	assert.Nil(t, assets)
	assert.Equal(t, appLayout, *missingAppLayout)
}

func Test_GiveniOSAppLayoutWithEntitlements_WhenProfileHasMissingEntitlements_ThenDoesNotFindIt(t *testing.T) {
	// Given
	certsByType := certsByType(t)
	manager, _ := createTestObjects(t)

	entitlements := entitlements()
	entitlements["main"] = "$(build-setting)"

	appLayout := autocodesign.AppLayout{
		Platform: autocodesign.IOS,
		EntitlementsByArchivableTargetBundleID: map[string]autocodesign.Entitlements{
			"io.ios.valid": entitlements,
		},
	}

	// When
	assets, missingAppLayout, err := manager.FindCodesignAssets(appLayout, autocodesign.Development, certsByType, []string{}, 0)

	// Then
	assert.NoError(t, err)
	assert.Nil(t, assets)
	assert.Equal(t, appLayout, *missingAppLayout)
}

// Helpers

func createTestObjects(t *testing.T) (Manager, []profileutil.ProvisioningProfileInfoModel) {
	profiles := profiles(t)

	mockProvider := new(mocks.ProvisioningProfileProvider)
	mockProvider.On("ListProvisioningProfiles", mock.Anything).Return(profiles, nil)

	mockConverter := new(mocks.ProvisioningProfileConverter)
	call := mockConverter.On("ProfileInfoToProfile", mock.Anything)
	call.RunFn = func(args mock.Arguments) {
		profileInfo, ok := args[0].(profileutil.ProvisioningProfileInfoModel)
		if !ok {
			t.Fatalf("Failed to cast arg to ProvisioningProfileInfoModel")
		}
		call.ReturnArguments = mock.Arguments{profileFromModel(profileInfo), nil}
	}

	return NewManager(mockProvider, mockConverter), profiles
}

func findCert(t *testing.T, certsByType map[appstoreconnect.CertificateType][]autocodesign.Certificate, serial string) certificateutil.CertificateInfo {
	for _, certs := range certsByType {
		for _, cert := range certs {
			if cert.CertificateInfo.Serial == serial {
				return cert.CertificateInfo
			}
		}
	}

	t.Fatalf("missing certificate")

	return certificateutil.CertificateInfo{}
}

func findProvProfile(t *testing.T, profiles []profileutil.ProvisioningProfileInfoModel, uuid string) autocodesign.Profile {
	for _, profile := range profiles {
		if profile.UUID == uuid {
			return profileFromModel(profile)
		}
	}

	t.Fatalf("missing profile")

	return Profile{}
}

func profileFromModel(profileInfo profileutil.ProvisioningProfileInfoModel) autocodesign.Profile {
	return Profile{
		attributes: appstoreconnect.ProfileAttributes{
			Name:           profileInfo.Name,
			UUID:           profileInfo.UUID,
			ProfileContent: []byte{},
			Platform:       getBundleIDPlatform(profileInfo.Type),
			ExpirationDate: devportaltime.Time(profileInfo.ExpirationDate),
		},
		id:             "",
		bundleID:       profileInfo.BundleID,
		certificateIDs: nil,
		deviceUDIDs:    nil,
	}
}

func profiles(t *testing.T) []profileutil.ProvisioningProfileInfoModel {
	now := time.Now()
	iosDevProfile := profileutil.ProvisioningProfileInfoModel{
		UUID:                  "uuid-1",
		Name:                  "Valid development profile",
		TeamName:              teamName,
		TeamID:                teamID,
		BundleID:              "io.ios.valid",
		ExportType:            exportoptions.MethodDevelopment,
		ProvisionedDevices:    []string{"device-1", "device-2", "device-3"},
		DeveloperCertificates: []certificateutil.CertificateInfo{devCert(t, now, now.AddDate(1, 0, 0))},
		CreationDate:          now.AddDate(0, -1, 0),
		ExpirationDate:        now.AddDate(0, 1, 0),
		Entitlements:          entitlements(),
		ProvisionsAllDevices:  false,
		Type:                  profileutil.ProfileTypeIos,
	}
	tvosDistProfile := profileutil.ProvisioningProfileInfoModel{
		UUID:                  "uuid-2",
		Name:                  "Valid distribution profile",
		TeamName:              teamName,
		TeamID:                teamID,
		BundleID:              "io.tvos.valid",
		ExportType:            exportoptions.MethodAppStore,
		ProvisionedDevices:    nil,
		DeveloperCertificates: []certificateutil.CertificateInfo{distCert(t, now, now.AddDate(1, 0, 0))},
		CreationDate:          now.AddDate(0, -1, 0),
		ExpirationDate:        now.AddDate(0, 1, 0),
		Entitlements:          nil,
		ProvisionsAllDevices:  true,
		Type:                  profileutil.ProfileTypeTvOs,
	}
	iosExpiredProfile := profileutil.ProvisioningProfileInfoModel{
		UUID:                  "uuid-3",
		Name:                  "Expired distribution profile",
		TeamName:              teamName,
		TeamID:                teamID,
		BundleID:              "io.tvos.expired",
		ExportType:            exportoptions.MethodAppStore,
		ProvisionedDevices:    nil,
		DeveloperCertificates: []certificateutil.CertificateInfo{distCert(t, now, now.AddDate(1, 0, 0))},
		CreationDate:          now.AddDate(0, -1, 0),
		ExpirationDate:        now.AddDate(0, 0, -1),
		Entitlements:          nil,
		ProvisionsAllDevices:  true,
		Type:                  profileutil.ProfileTypeIos,
	}
	iosWildcardDevProfile := profileutil.ProvisioningProfileInfoModel{
		UUID:                  "uuid-4",
		Name:                  "Valid wildcard development profile",
		TeamName:              teamName,
		TeamID:                teamID,
		BundleID:              "io.ios.*",
		ExportType:            exportoptions.MethodDevelopment,
		ProvisionedDevices:    []string{"device-1", "device-2", "device-3"},
		DeveloperCertificates: []certificateutil.CertificateInfo{devCert(t, now, now.AddDate(1, 0, 0))},
		CreationDate:          now.AddDate(0, -1, 0),
		ExpirationDate:        now.AddDate(0, 1, 0),
		Entitlements:          entitlements(),
		ProvisionsAllDevices:  false,
		Type:                  profileutil.ProfileTypeIos,
	}

	return []profileutil.ProvisioningProfileInfoModel{iosDevProfile, tvosDistProfile, iosExpiredProfile, iosWildcardDevProfile}
}

func entitlements() map[string]interface{} {
	return map[string]interface{}{
		"main": []interface{}{
			"this-is-the-main-value",
		},
		"sub": "test",
	}
}

func certsByType(t *testing.T) map[appstoreconnect.CertificateType][]autocodesign.Certificate {
	notBefore := time.Now()
	expiry := notBefore.AddDate(1, 0, 0)
	devCert := autocodesign.Certificate{
		CertificateInfo: devCert(t, notBefore, expiry),
		ID:              "dev",
	}
	distCert := autocodesign.Certificate{
		CertificateInfo: distCert(t, notBefore, expiry),
		ID:              "dist",
	}

	return map[appstoreconnect.CertificateType][]autocodesign.Certificate{
		appstoreconnect.IOSDevelopment:  {devCert},
		appstoreconnect.IOSDistribution: {distCert},
	}
}

func devCert(t *testing.T, notBefore, expiry time.Time) certificateutil.CertificateInfo {
	certInfo, err := certificateutil.GenerateTestCertificateInfo(1, teamID, teamName, "Development certificate", notBefore, expiry)
	require.NoError(t, err)
	return certInfo
}

func distCert(t *testing.T, notBefore, expiry time.Time) certificateutil.CertificateInfo {
	certInfo, err := certificateutil.GenerateTestCertificateInfo(1, teamID, teamName, "Distribution certificate", notBefore, expiry)
	require.NoError(t, err)
	return certInfo
}
