package localcodesignasset

import (
	"reflect"
	"strings"
	"time"

	"github.com/bitrise-io/go-xcode/autocodesign"
	"github.com/bitrise-io/go-xcode/autocodesign/devportalclient/appstoreconnect"
	"github.com/bitrise-io/go-xcode/profileutil"
)

type Manager struct {
	profileProvider ProvisioningProfileProvider
}

func NewManager(provisioningProfileProvider ProvisioningProfileProvider) Manager {
	return Manager{
		profileProvider: provisioningProfileProvider,
	}
}

/*
// AppCodesignAssets is the result of ensuring codesigning assets
type AppCodesignAssets struct {
	ArchivableTargetProfilesByBundleID map[string]Profile
	UITestTargetProfilesByBundleID     map[string]Profile
	Certificate                        certificateutil.CertificateInfoModel
}
*/
//func (m Manager) FindCodesignAssets(appLayout autocodesign.AppLayout, distrTypes []autocodesign.DistributionType, certsByType map[appstoreconnect.CertificateType][]autocodesign.Certificate, deviceIDs []string, minProfileDaysValid int) map[autocodesign.DistributionType]autocodesign.AppCodesignAssets {
//	profiles, err := m.profileProvider.ListProvisioningProfiles()
//	if err != nil {
//		return appLayout
//	}
//
//	for _, distrType := range distrTypes {
//		certSerials := certificateSerials(certsByType, distrType)
//
//		for bundleID, entitlements := range appLayout.EntitlementsByArchivableTargetBundleID {
//			profile := findProfile(profiles, appLayout.Platform, distrType, bundleID, entitlements, minProfileDaysValid, certSerials, deviceIDs)
//
//			if profile == nil {
//				continue
//			}
//
//			delete(appLayout.EntitlementsByArchivableTargetBundleID, bundleID)
//		}
//	}
//
//	for i, bundleID := range appLayout.UITestTargetBundleIDs {
//		for _, profile := range profiles {
//			//TODO: Is bundle id check enough or should we do the full profile id check?
//			if profile.BundleID == bundleID {
//				remove(appLayout.UITestTargetBundleIDs, i)
//			}
//		}
//	}
//
//	return appLayout
//}

func (m Manager) FindMissingCodesignAssets(appLayout autocodesign.AppLayout, distrTypes []autocodesign.DistributionType, certsByType map[appstoreconnect.CertificateType][]autocodesign.Certificate, deviceIDs []string, minProfileDaysValid int) autocodesign.AppLayout {
	profiles, err := m.profileProvider.ListProvisioningProfiles()
	if err != nil {
		return appLayout
	}

	for _, distrType := range distrTypes {
		certSerials := certificateSerials(certsByType, distrType)

		for bundleID, entitlements := range appLayout.EntitlementsByArchivableTargetBundleID {
			profile := findProfile(profiles, appLayout.Platform, distrType, bundleID, entitlements, minProfileDaysValid, certSerials, deviceIDs)

			if profile == nil {
				continue
			}

			delete(appLayout.EntitlementsByArchivableTargetBundleID, bundleID)
		}
	}

	for i, bundleID := range appLayout.UITestTargetBundleIDs {
		for _, profile := range profiles {
			//TODO: Is bundle id check enough or should we do the full profile id check?
			if profile.BundleID == bundleID {
				remove(appLayout.UITestTargetBundleIDs, i)
			}
		}
	}

	return appLayout
}

func findProfile(localProfiles []profileutil.ProvisioningProfileInfoModel, platform autocodesign.Platform, distributionType autocodesign.DistributionType, bundleID string, entitlements autocodesign.Entitlements, minProfileDaysValid int, certSerials []string, deviceIDs []string) *profileutil.ProvisioningProfileInfoModel {
	for _, profile := range localProfiles {
		if isProfileMatching(profile, platform, distributionType, bundleID, entitlements, minProfileDaysValid, certSerials, deviceIDs) {
			return &profile
		}
	}

	return nil
}

func isProfileMatching(profile profileutil.ProvisioningProfileInfoModel, platform autocodesign.Platform, distributionType autocodesign.DistributionType, bundleID string, entitlements autocodesign.Entitlements, minProfileDaysValid int, certSerials []string, deviceIDs []string) bool {
	if !isActive(profile, minProfileDaysValid) {
		return false
	}

	if !hasMatchingDistributionType(profile, distributionType) {
		return false
	}

	if !hasMatchingBundleID(profile, bundleID) {
		return false
	}

	if !hasMatchingPlatform(profile, platform) {
		return false
	}

	if !hasMatchingLocalCertificate(profile, certSerials) {
		return false
	}

	if !containsAllAppEntitlements(profile, entitlements) {
		return false
	}

	if !provisionsDevices(profile, deviceIDs) {
		return false
	}

	return true
}

//TODO: Kept these here while developing the feature. Move it to utils.go.

func hasMatchingBundleID(profile profileutil.ProvisioningProfileInfoModel, bundleID string) bool {
	return profile.BundleID == bundleID
}

func hasMatchingLocalCertificate(profile profileutil.ProvisioningProfileInfoModel, localCertificateSerials []string) bool {
	var profileCertificateSerials []string
	for _, certificate := range profile.DeveloperCertificates {
		profileCertificateSerials = append(profileCertificateSerials, certificate.Serial)
	}

	return 0 < len(intersection(localCertificateSerials, profileCertificateSerials))
}

func containsAllAppEntitlements(profile profileutil.ProvisioningProfileInfoModel, appEntitlements autocodesign.Entitlements) bool {
	profileEntitlements := autocodesign.Entitlements(profile.Entitlements)
	hasMissingEntitlement := false

	for key, value := range appEntitlements {
		profileEntitlementValue := profileEntitlements[key]

		// TODO: Better entitlements comparison
		if reflect.DeepEqual(profileEntitlementValue, value) {
			hasMissingEntitlement = true
			break
		}
	}

	return !hasMissingEntitlement
}

func hasMatchingDistributionType(profile profileutil.ProvisioningProfileInfoModel, distributionType autocodesign.DistributionType) bool {
	return autocodesign.DistributionType(profile.ExportType) == distributionType
}

func isActive(profile profileutil.ProvisioningProfileInfoModel, minProfileDaysValid int) bool {
	expiration := time.Now()
	if minProfileDaysValid > 0 {
		expiration = expiration.AddDate(0, 0, minProfileDaysValid)
	}

	return expiration.Before(profile.ExpirationDate)
}

func hasMatchingPlatform(profile profileutil.ProvisioningProfileInfoModel, platform autocodesign.Platform) bool {
	//TODO: Remove lowercasing?
	return strings.ToLower(string(platform)) == string(profile.Type)
}

func provisionsDevices(profile profileutil.ProvisioningProfileInfoModel, deviceIDs []string) bool {
	if profile.ProvisionsAllDevices || len(deviceIDs) == 0 {
		return true
	}

	if len(profile.ProvisionedDevices) == 0 {
		return false
	}

	for _, deviceID := range deviceIDs {
		if contains(profile.ProvisionedDevices, deviceID) {
			continue
		}
		return false
	}

	return true
}
