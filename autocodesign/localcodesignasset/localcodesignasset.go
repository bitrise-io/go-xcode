package localcodesignasset

import (
	"fmt"
	"time"

	"github.com/bitrise-io/go-xcode/autocodesign"
	"github.com/bitrise-io/go-xcode/autocodesign/devportalclient/appstoreconnect"
	"github.com/bitrise-io/go-xcode/profileutil"
)

type LocalCodeSignAsset interface {
	FindMissingCodesingAssets(appLayout autocodesign.AppLayout, distrTypes []autocodesign.DistributionType, certsByType map[appstoreconnect.CertificateType][]autocodesign.Certificate) autocodesign.AppLayout
}

type manager struct {
	profileProvider ProvisioningProfileProvider
}

func New(provisioningProfileProvider ProvisioningProfileProvider) LocalCodeSignAsset {
	return manager{
		profileProvider: provisioningProfileProvider,
	}
}

/*
Check if the given profile:
- x type matches the desired distribution type
- x active (not expired)?
- x bundle id matches
- x contains the given certificate ids
- x contains the given entitlements
- contains the given deviceIDs
- x if valid for <minProfileDaysValid>
- platform matching
*/

func isProfileMatching(profile profileutil.ProvisioningProfileInfoModel, platform autocodesign.Platform, distributionType autocodesign.DistributionType, bundleID string, entitlements autocodesign.Entitlements, minProfileDaysValid int, certSerials []string) bool {
	if profile.BundleID != bundleID {
		return false
	}

	var devCertSerials []string
	for _, devCert := range profile.DeveloperCertificates {
		devCertSerials = append(devCertSerials, devCert.Serial)
	}

	if len(intersection(certSerials, devCertSerials)) == 0 {
		return false
	}

	profileEntitlements := autocodesign.Entitlements(profile.Entitlements)
	hasMissingEntitlement := false

	for key, value := range entitlements {
		profileEntitlementValue := profileEntitlements[key]

		if profileEntitlementValue == nil || profileEntitlementValue != value {
			hasMissingEntitlement = true
			break
		}
	}

	if hasMissingEntitlement {
		return false
	}

	if autocodesign.DistributionType(profile.ExportType) != distributionType {
		return false
	}

	expiration := time.Now()
	if minProfileDaysValid > 0 {
		expiration = expiration.AddDate(0, 0, minProfileDaysValid)
	}
	if !expiration.Before(profile.ExpirationDate) {
		return false
	}

	profile.Type

	return true
}

func findProfile(localProfiles []profileutil.ProvisioningProfileInfoModel, platform autocodesign.Platform, distributionType autocodesign.DistributionType, bundleID string, entitlements autocodesign.Entitlements) (*profileutil.ProvisioningProfileInfoModel, error) {
	platformProfileTypes, ok := autocodesign.PlatformToProfileTypeByDistribution[platform]
	if !ok {
		return nil, fmt.Errorf("no profiles for platform: %s", platform)
	}

	profileType := platformProfileTypes[distributionType]

	profileName := autocodesign.ProfileName(profileType, bundleID)

	for _, profile := range localProfiles {
	}

	return nil
}

func (m manager) FindMissingCodesingAssets(appLayout autocodesign.AppLayout, distrTypes []autocodesign.DistributionType, certsByType map[appstoreconnect.CertificateType][]autocodesign.Certificate) autocodesign.AppLayout {
	profiles, err := m.profileProvider.ListProvisioningProfiles()
	if err != nil {
		return appLayout
	}

	var localAppLayout = appLayout

	for _, distrType := range distrTypes {
		certType := autocodesign.CertificateTypeByDistribution[distrType]
		certs := certsByType[certType]

		// The other place was using the cert ID but the profiles returned by the profileProvider only have the CertificateInfoModel type which only has a serial field.
		var certSerials []string
		for _, cert := range certs {
			certSerials = append(certSerials, cert.CertificateInfo.Serial)
		}

		// The other place also does this. Not sure if it is needed.
		//platformProfileTypes, _ := autocodesign.PlatformToProfileTypeByDistribution[appLayout.Platform]
		//profileType := platformProfileTypes[distrType]

		for bundleID, entitlements := range appLayout.EntitlementsByArchivableTargetBundleID {
			for _, profile := range profiles {
				if profile.BundleID != bundleID {
					continue
				}

				var devCertSerials []string
				for _, devCert := range profile.DeveloperCertificates {
					devCertSerials = append(devCertSerials, devCert.Serial)
				}

				if len(intersection(certSerials, devCertSerials)) == 0 {
					continue
				}

				profileEntitlements := autocodesign.Entitlements(profile.Entitlements)
				hasMissingEntitlement := false

				for key, value := range entitlements {
					profileEntitlementValue := profileEntitlements[key]

					if profileEntitlementValue == nil || profileEntitlementValue != value {
						hasMissingEntitlement = true
						break
					}
				}

				if hasMissingEntitlement {
					continue
				}

				delete(localAppLayout.EntitlementsByArchivableTargetBundleID, bundleID)
			}
		}

		for i, bundleID := range appLayout.UITestTargetBundleIDs {
			for _, profile := range profiles {
				if profile.BundleID == bundleID {
					remove(localAppLayout.UITestTargetBundleIDs, i)
				}
			}
		}
	}

	return localAppLayout
}

func remove(slice []string, i int) []string {
	copy(slice[i:], slice[i+1:])
	return slice[:len(slice)-1]
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
