package localcodesignasset

import (
	"strings"
	"time"

	"github.com/bitrise-io/go-xcode/autocodesign"
	"github.com/bitrise-io/go-xcode/autocodesign/devportalclient/appstoreconnect"
	"github.com/bitrise-io/go-xcode/profileutil"
)

type LocalCodeSignAsset interface {
	FindMissingCodesignAssets(appLayout autocodesign.AppLayout, distrTypes []autocodesign.DistributionType, certsByType map[appstoreconnect.CertificateType][]autocodesign.Certificate, deviceIDs []string, minProfileDaysValid int) autocodesign.AppLayout
}

type manager struct {
	profileProvider ProvisioningProfileProvider
}

func New(provisioningProfileProvider ProvisioningProfileProvider) LocalCodeSignAsset {
	return manager{
		profileProvider: provisioningProfileProvider,
	}
}

func (m manager) FindMissingCodesignAssets(appLayout autocodesign.AppLayout, distrTypes []autocodesign.DistributionType, certsByType map[appstoreconnect.CertificateType][]autocodesign.Certificate, deviceIDs []string, minProfileDaysValid int) autocodesign.AppLayout {
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

/*
Check if the given profile:
- x type matches the desired distribution type
- x active (not expired)?
- x bundle id matches
- x contains the given certificate ids
- x contains the given entitlements
- x contains the given deviceIDs
- x if valid for <minProfileDaysValid>
- x platform matching
*/
func isProfileMatching(profile profileutil.ProvisioningProfileInfoModel, platform autocodesign.Platform, distributionType autocodesign.DistributionType, bundleID string, entitlements autocodesign.Entitlements, minProfileDaysValid int, certSerials []string, deviceIDs []string) bool {
	if isActive(profile, minProfileDaysValid) == false {
		return false
	}

	if hasMatchingDistributionType(profile, distributionType) == false {
		return false
	}

	if hasMatchingBundleID(profile, bundleID) == false {
		return false
	}

	if hasMatchingPlatform(profile, platform) == false {
		return false
	}

	if hasMatchingLocalCertificate(profile, certSerials) == false {
		return false
	}

	if containsAllAppEntitlements(profile, entitlements) == false {
		return false
	}

	if provisionsDevices(profile, deviceIDs) == false {
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

		//TODO: Better interface value comparison
		if profileEntitlementValue == nil || profileEntitlementValue != value {
			hasMissingEntitlement = true
			break
		}
	}

	return hasMissingEntitlement == false
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
