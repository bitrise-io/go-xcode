package profileutil

import (
	"fmt"
	"strings"
	"time"

	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/bitrise-io/go-xcode/exportoptions"
	"github.com/bitrise-io/go-xcode/v2/plistutil"
)

type PlistData plistutil.PlistData

func (profile PlistData) GetProfileType() (ProfileType, error) {
	data := plistutil.PlistData(profile)
	platforms, _ := data.GetStringArray("Platform")
	if len(platforms) == 0 {
		return "", fmt.Errorf("missing Platform array in profile")
	}

	platform := strings.ToLower(platforms[0])
	var profileType ProfileType

	switch platform {
	case string(ProfileTypeIos):
		profileType = ProfileTypeIos
	case string(ProfileTypeMacOs):
		profileType = ProfileTypeMacOs
	case string(ProfileTypeTvOs):
		profileType = ProfileTypeTvOs
	default:
		return "", fmt.Errorf("unknown platform type: %s", platform)
	}

	return profileType, nil
}

func (profile PlistData) GetUUID() string {
	data := plistutil.PlistData(profile)
	uuid, _ := data.GetString("UUID")
	return uuid
}

func (profile PlistData) GetName() string {
	data := plistutil.PlistData(profile)
	name, _ := data.GetString("Name")
	return name
}

func (profile PlistData) GetApplicationIdentifier() string {
	data := plistutil.PlistData(profile)
	entitlements, ok := data.GetMapStringInterface("Entitlements")
	if !ok {
		return ""
	}

	applicationID, ok := entitlements.GetString("application-identifier")
	if !ok {
		applicationID, ok = entitlements.GetString("com.apple.application-identifier")
		if !ok {
			return ""
		}
	}
	return applicationID
}

func (profile PlistData) GetBundleIdentifier() string {
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

func (profile PlistData) GetExportMethod() exportoptions.Method {
	data := plistutil.PlistData(profile)
	entitlements, _ := data.GetMapStringInterface("Entitlements")
	platform, _ := data.GetStringArray("Platform")

	if len(platform) != 0 {
		switch strings.ToLower(platform[0]) {
		case "osx":
			_, ok := data.GetStringArray("ProvisionedDevices")
			if !ok {
				if allDevices, ok := data.GetBool("ProvisionsAllDevices"); ok && allDevices {
					return exportoptions.MethodDeveloperID
				}
				return exportoptions.MethodAppStore
			}
			return exportoptions.MethodDevelopment
		case "ios", "tvos":
			_, ok := data.GetStringArray("ProvisionedDevices")
			if !ok {
				if allDevices, ok := data.GetBool("ProvisionsAllDevices"); ok && allDevices {
					return exportoptions.MethodEnterprise
				}
				return exportoptions.MethodAppStore
			}
			if allow, ok := entitlements.GetBool("get-task-allow"); ok && allow {
				return exportoptions.MethodDevelopment
			}
			return exportoptions.MethodAdHoc
		}
	}

	return exportoptions.MethodDefault
}

func (profile PlistData) GetEntitlements() plistutil.PlistData {
	data := plistutil.PlistData(profile)
	entitlements, _ := data.GetMapStringInterface("Entitlements")
	return entitlements
}

func (profile PlistData) GetTeamID() string {
	data := plistutil.PlistData(profile)
	entitlements, ok := data.GetMapStringInterface("Entitlements")
	if ok {
		teamID, _ := entitlements.GetString("com.apple.developer.team-identifier")
		return teamID
	}
	return ""
}

func (profile PlistData) GetExpirationDate() time.Time {
	data := plistutil.PlistData(profile)
	expiry, _ := data.GetTime("ExpirationDate")
	return expiry
}

func (profile PlistData) GetProvisionedDevices() []string {
	data := plistutil.PlistData(profile)
	devices, _ := data.GetStringArray("ProvisionedDevices")
	return devices
}

func (profile PlistData) GetDeveloperCertificates() [][]byte {
	data := plistutil.PlistData(profile)
	developerCertificates, _ := data.GetByteArrayArray("DeveloperCertificates")
	return developerCertificates
}

func (profile PlistData) GetDeveloperCertificateInfo() []certificateutil.CertificateInfoModel {
	certificateBytesList := profile.GetDeveloperCertificates()

	var certificateInfos []certificateutil.CertificateInfoModel
	for _, certificateBytes := range certificateBytesList {
		certificate, err := certificateutil.CertificateFromDERContent(certificateBytes)
		if err != nil || certificate == nil {
			continue
		}

		certificateInfo := certificateutil.NewCertificateInfo(*certificate, nil)
		certificateInfos = append(certificateInfos, certificateInfo)
	}

	return certificateInfos
}

func (profile PlistData) GetTeamName() string {
	data := plistutil.PlistData(profile)
	teamName, _ := data.GetString("TeamName")
	return teamName
}

func (profile PlistData) GetCreationDate() time.Time {
	data := plistutil.PlistData(profile)
	creationDate, _ := data.GetTime("CreationDate")
	return creationDate
}

func (profile PlistData) GetProvisionsAllDevices() bool {
	data := plistutil.PlistData(profile)
	provisionsAlldevices, _ := data.GetBool("ProvisionsAllDevices")
	return provisionsAlldevices
}
