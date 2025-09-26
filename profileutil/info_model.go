package profileutil

import (
	"fmt"
	"strings"
	"time"

	"github.com/bitrise-io/go-xcode/v2/certificateutil"
	"github.com/bitrise-io/go-xcode/v2/exportoptions"
	"github.com/bitrise-io/go-xcode/v2/plistutil"
)

// ProvisioningProfileInfoModel ...
type ProvisioningProfileInfoModel struct {
	UUID                  string
	Name                  string
	TeamName              string
	TeamID                string
	BundleID              string
	ExportType            exportoptions.Method
	ProvisionedDevices    []string
	DeveloperCertificates []certificateutil.CertificateInfoModel
	CreationDate          time.Time
	ExpirationDate        time.Time
	Entitlements          plistutil.PlistData
	ProvisionsAllDevices  bool
	Type                  ProfileType
}

func (info ProvisioningProfileInfoModel) String(installedCertificates ...certificateutil.CertificateInfoModel) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Name: %s (%s)\n", info.Name, info.UUID))
	builder.WriteString(fmt.Sprintf("Export Type: %s\n", info.ExportType))
	builder.WriteString(fmt.Sprintf("Team: %s (%s)\n", info.TeamName, info.TeamID))
	builder.WriteString(fmt.Sprintf("Bundle ID: %s\n", info.BundleID))
	builder.WriteString(fmt.Sprintf("Expiry: %s\n", info.ExpirationDate))
	builder.WriteString(fmt.Sprintf("Is Xcode Managed: %t\n", info.IsXcodeManaged()))

	builder.WriteString("Capabilities:\n")
	for key, value := range collectCapabilitiesPrintableInfo(info.Entitlements) {
		builder.WriteString(fmt.Sprintf("  - %s: %v\n", key, value))
	}

	if info.ProvisionedDevices != nil {
		builder.WriteString("Devices:\n")
		for _, device := range info.ProvisionedDevices {
			builder.WriteString(fmt.Sprintf("  - %s\n", device))
		}
	}

	builder.WriteString("Certificates:\n")
	for _, certificateInfo := range info.DeveloperCertificates {
		builder.WriteString(fmt.Sprintf("  - Name: %s, Serial: %s, Team ID: %s\n",
			certificateInfo.CommonName, certificateInfo.Serial, certificateInfo.TeamID))
	}

	var errors []string
	if installedCertificates != nil && !info.hasInstalledCertificate(installedCertificates) {
		errors = append(errors, "None of the profile's certificates are installed")
	}
	if err := info.CheckValidity(); err != nil {
		errors = append(errors, err.Error())
	}
	if len(errors) > 0 {
		builder.WriteString("Errors:\n")
		for _, err := range errors {
			builder.WriteString(fmt.Sprintf("  - %s\n", err))
		}
	}

	return builder.String()
}

func (info ProvisioningProfileInfoModel) IsXcodeManaged() bool {
	return isXcodeManaged(info.Name)
}

// CheckValidity ...
func (info ProvisioningProfileInfoModel) CheckValidity() error {
	// TODO: directly using time.Now() makes testing difficult
	timeNow := time.Now()
	if !timeNow.Before(info.ExpirationDate) {
		return fmt.Errorf("provisioning profile expired at: %s", info.ExpirationDate)
	}
	return nil
}

// hasInstalledCertificate ...
func (info ProvisioningProfileInfoModel) hasInstalledCertificate(installedCertificates []certificateutil.CertificateInfoModel) bool {
	has := false
	for _, certificate := range info.DeveloperCertificates {
		for _, installedCertificate := range installedCertificates {
			if certificate.Serial == installedCertificate.Serial {
				has = true
				break
			}
		}
	}
	return has
}

func isXcodeManaged(profileName string) bool {
	if strings.HasPrefix(profileName, "XC") {
		return true
	}
	if strings.Contains(profileName, "Provisioning Profile") {
		if strings.HasPrefix(profileName, "iOS Team") ||
			strings.HasPrefix(profileName, "Mac Catalyst Team") ||
			strings.HasPrefix(profileName, "tvOS Team") ||
			strings.HasPrefix(profileName, "Mac Team") {
			return true
		}
	}
	return false
}

func collectCapabilitiesPrintableInfo(entitlements plistutil.PlistData) map[string]interface{} {
	capabilities := map[string]interface{}{}

	for key, value := range entitlements {
		if KnownProfileCapabilitiesMap[ProfileTypeIos][key] ||
			KnownProfileCapabilitiesMap[ProfileTypeMacOs][key] {
			capabilities[key] = value
		}
	}

	return capabilities
}
