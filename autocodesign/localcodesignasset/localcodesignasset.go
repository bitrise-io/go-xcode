package localcodesignasset

import (
	"fmt"
	"io/ioutil"
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

// Profile ...
type Profile struct {
	attributes     appstoreconnect.ProfileAttributes
	id             string
	bundleID       string
	deviceIDs      []string
	certificateIDs []string
}

// ID ...
func (p Profile) ID() string {
	return p.id
}

// Attributes ...
func (p Profile) Attributes() appstoreconnect.ProfileAttributes {
	return p.attributes
}

// CertificateIDs ...
func (p Profile) CertificateIDs() ([]string, error) {
	return p.certificateIDs, nil
}

// DeviceIDs ...
func (p Profile) DeviceIDs() ([]string, error) {
	return p.deviceIDs, nil
}

// BundleID ...
func (p Profile) BundleID() (appstoreconnect.BundleID, error) {
	return appstoreconnect.BundleID{
		ID: p.id,
		Attributes: appstoreconnect.BundleIDAttributes{
			Identifier: p.bundleID,
			Name:       p.attributes.Name,
		},
	}, nil
}

// Entitlements ...
func (p Profile) Entitlements() (autocodesign.Entitlements, error) {
	return autocodesign.ParseRawProfileEntitlements(p.attributes.ProfileContent)
}

/*
	IOS       BundleIDPlatform = "IOS"
	MacOS     BundleIDPlatform = "MAC_OS"
	Universal BundleIDPlatform = "UNIVERSAL"
*/

func getBundleIDPlatform(profileType profileutil.ProfileType) appstoreconnect.BundleIDPlatform {
	switch profileType {
	case profileutil.ProfileTypeIos, profileutil.ProfileTypeTvOs:
		return appstoreconnect.IOS
	case profileutil.ProfileTypeMacOs:
		return appstoreconnect.MacOS
	}

	return ""
}

func ProfileInfoToProfile(info profileutil.ProvisioningProfileInfoModel) (autocodesign.Profile, error) {
	_, pth, err := profileutil.FindProvisioningProfile(info.UUID)
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadFile(pth)
	if err != nil {
		return nil, err
	}

	return Profile{
		attributes: appstoreconnect.ProfileAttributes{
			Name:           info.Name,
			UUID:           info.UUID,
			ProfileContent: content,
			Platform:       getBundleIDPlatform(info.Type),
			ExpirationDate: appstoreconnect.Time(info.ExpirationDate),
		},
		id:             "", // only in case of Developer Portal Profiles
		bundleID:       info.BundleID,
		certificateIDs: nil, // only in case of Developer Portal Profiles
		deviceIDs:      nil, // only in case of Developer Portal Profiles
	}, nil
}

func (m Manager) FindCodesignAssets(appLayout autocodesign.AppLayout, distrTypes []autocodesign.DistributionType, certsByType map[appstoreconnect.CertificateType][]autocodesign.Certificate, deviceIDs []string, minProfileDaysValid int) (map[autocodesign.DistributionType]autocodesign.AppCodesignAssets, *autocodesign.AppLayout, error) {
	profiles, err := m.profileProvider.ListProvisioningProfiles()
	if err != nil {
		return nil, nil, err
	}

	assetsByDistribution := map[autocodesign.DistributionType]autocodesign.AppCodesignAssets{}

	for _, distrType := range distrTypes {
		certSerials := certificateSerials(certsByType, distrType)

		var asset *autocodesign.AppCodesignAssets
		for bundleID, entitlements := range appLayout.EntitlementsByArchivableTargetBundleID {
			profileInfo := findProfile(profiles, appLayout.Platform, distrType, bundleID, entitlements, minProfileDaysValid, certSerials, deviceIDs)
			if profileInfo == nil {
				continue
			}

			profile, err := ProfileInfoToProfile(*profileInfo)
			if err != nil {
				return nil, nil, err
			}

			if asset == nil {
				asset = &autocodesign.AppCodesignAssets{
					ArchivableTargetProfilesByBundleID: map[string]autocodesign.Profile{
						bundleID: profile,
					},
				}
			} else {
				profileByArchivableTargetBundleID := asset.ArchivableTargetProfilesByBundleID
				if profileByArchivableTargetBundleID == nil {
					profileByArchivableTargetBundleID = map[string]autocodesign.Profile{}
				}

				profileByArchivableTargetBundleID[bundleID] = profile
				asset.ArchivableTargetProfilesByBundleID = profileByArchivableTargetBundleID
			}

			delete(appLayout.EntitlementsByArchivableTargetBundleID, bundleID)
		}

		if distrType == autocodesign.Development {
			for i, bundleID := range appLayout.UITestTargetBundleIDs {
				wildcardBundleID, err := autocodesign.CreateWildcardBundleID(bundleID)
				if err != nil {
					return nil, nil, fmt.Errorf("could not create wildcard bundle id: %s", err)
				}

				// Capabilities are not supported for UITest targets.
				profileInfo := findProfile(profiles, appLayout.Platform, distrType, wildcardBundleID, nil, minProfileDaysValid, certSerials, deviceIDs)
				if profileInfo == nil {
					continue
				}

				profile, err := ProfileInfoToProfile(*profileInfo)
				if err != nil {
					return nil, nil, err
				}

				if asset == nil {
					asset = &autocodesign.AppCodesignAssets{
						UITestTargetProfilesByBundleID: map[string]autocodesign.Profile{
							bundleID: profile,
						},
					}
				} else {
					profileByUITestTargetBundleID := asset.UITestTargetProfilesByBundleID
					if profileByUITestTargetBundleID == nil {
						profileByUITestTargetBundleID = map[string]autocodesign.Profile{}
					}

					profileByUITestTargetBundleID[bundleID] = profile
					asset.UITestTargetProfilesByBundleID = profileByUITestTargetBundleID
				}

				remove(appLayout.UITestTargetBundleIDs, i)
			}
		}

		if asset != nil {
			certType := autocodesign.CertificateTypeByDistribution[distrType]
			certs := certsByType[certType]
			cert := certs[0]

			// TODO: This and the certificate part of the ensureProfile function call should be extracted
			asset.Certificate = cert.CertificateInfo

			assetsByDistribution[distrType] = *asset
		}
	}

	if len(appLayout.EntitlementsByArchivableTargetBundleID) == 0 && len(appLayout.UITestTargetBundleIDs) == 0 {
		return assetsByDistribution, nil, nil
	}

	return assetsByDistribution, &appLayout, nil
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
