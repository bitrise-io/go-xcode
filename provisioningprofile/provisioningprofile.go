package provisioningprofile

import (
	"strings"
	"time"

	"github.com/bitrise-io/steps-certificate-and-profile-installer/profileutil"
	"github.com/bitrise-tools/go-xcode/exportoptions"
	"github.com/bitrise-tools/go-xcode/plistutil"
	"howett.net/plist"
)

const (
	notValidParameterErrorMessage = "security: SecPolicySetValue: One or more parameters passed to a function were not valid."
)

// Profile ...
type Profile plistutil.PlistData

// NewProfileFromFile ...
func NewProfileFromFile(provisioningProfilePth string) (Profile, error) {
	pkcs7, err := profileutil.ProvisioningProfileFromFile(provisioningProfilePth)
	if err != nil {
		return Profile{}, err
	}

	var plistData plistutil.PlistData
	if _, err := plist.Unmarshal(pkcs7.Content, &plistData); err != nil {
		return Profile{}, err
	}

	return Profile(plistData), nil
}

// GetUUID ...
func (profile Profile) GetUUID() string {
	data := plistutil.PlistData(profile)
	uuid, _ := data.GetString("UUID")
	return uuid
}

// GetName ...
func (profile Profile) GetName() string {
	data := plistutil.PlistData(profile)
	uuid, _ := data.GetString("Name")
	return uuid
}

// GetApplicationIdentifier ...
func (profile Profile) GetApplicationIdentifier() string {
	data := plistutil.PlistData(profile)
	entitlements, ok := data.GetMapStringInterface("Entitlements")
	if !ok {
		return ""
	}

	applicationID, ok := entitlements.GetString("application-identifier")
	if !ok {
		return ""
	}
	return applicationID
}

// GetBundleIdentifier ...
func (profile Profile) GetBundleIdentifier() string {
	applicationID := profile.GetApplicationIdentifier()

	plistData := plistutil.PlistData(profile)
	prefixes, found := plistData.GetStringArray("ApplicationIdentifierPrefix")
	if found {
		for _, prefix := range prefixes {
			applicationID = strings.TrimPrefix(applicationID, prefix+".")
		}
	}

	teamID := profile.GetTeamID()
	return strings.TrimPrefix(applicationID, teamID+".")
}

// GetExportMethod ...
func (profile Profile) GetExportMethod() exportoptions.Method {
	data := plistutil.PlistData(profile)
	_, ok := data.GetStringArray("ProvisionedDevices")
	if !ok {
		if allDevices, ok := data.GetBool("ProvisionsAllDevices"); ok && allDevices {
			return exportoptions.MethodEnterprise
		}
		return exportoptions.MethodAppStore
	}

	entitlements, ok := data.GetMapStringInterface("Entitlements")
	if ok {
		if allow, ok := entitlements.GetBool("get-task-allow"); ok && allow {
			return exportoptions.MethodDevelopment
		}
		return exportoptions.MethodAdHoc
	}

	return exportoptions.MethodDefault
}

// GetEntitlements ...
func (profile Profile) GetEntitlements() plistutil.PlistData {
	data := plistutil.PlistData(profile)
	entitlements, _ := data.GetMapStringInterface("Entitlements")
	return entitlements
}

// GetTeamID ...
func (profile Profile) GetTeamID() string {
	data := plistutil.PlistData(profile)
	entitlements, ok := data.GetMapStringInterface("Entitlements")
	if ok {
		teamID, _ := entitlements.GetString("com.apple.developer.team-identifier")
		return teamID
	}
	return ""
}

// GetExpirationDate ...
func (profile Profile) GetExpirationDate() time.Time {
	data := plistutil.PlistData(profile)
	expiry, _ := data.GetTime("ExpirationDate")
	return expiry
}

// GetProvisionedDevices ...
func (profile Profile) GetProvisionedDevices() []string {
	data := plistutil.PlistData(profile)
	devices, _ := data.GetStringArray("ProvisionedDevices")
	return devices
}

// GetDeveloperCertificates ...
func (profile Profile) GetDeveloperCertificates() [][]byte {
	data := plistutil.PlistData(profile)
	developerCertificates, _ := data.GetByteArrayArray("DeveloperCertificates")
	return developerCertificates
}
