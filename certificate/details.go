package certificate

import (
	"fmt"
	"strings"
	"time"
)

type Type string

const (
	AppleDevelopment  Type = "Apple Development"
	AppleDistribution Type = "Apple Distribution"

	iPhoneDeveloper    Type = "iPhone Developer"
	iPhoneDistribution Type = "iPhone Distribution"

	MacDeveloper                      Type = "Mac Developer"
	ThirdPartyMacDeveloperApplication Type = "3rd Party Mac Developer Application"
	ThirdPartyMacDeveloperInstaller   Type = "3rd Party Mac Developer Installer"
	DeveloperIDApplication            Type = "Developer ID Application"
	DeveloperIDInstaller              Type = "Developer ID Installer"
)

var knownSoftwareCertificateTypes = map[Type]bool{
	AppleDevelopment:                  true,
	AppleDistribution:                 true,
	iPhoneDeveloper:                   true,
	iPhoneDistribution:                true,
	MacDeveloper:                      true,
	ThirdPartyMacDeveloperApplication: true,
	ThirdPartyMacDeveloperInstaller:   true,
	DeveloperIDApplication:            true,
	DeveloperIDInstaller:              true,
}

type Platform string

const (
	IOS   Platform = "iOS"
	MacOS Platform = "macOS"
	All   Platform = "All"
)

type Details struct {
	CommonName      string
	TeamName        string
	TeamID          string
	EndDate         time.Time
	StartDate       time.Time
	Serial          string
	SHA1Fingerprint string
}

func (d Details) Type() Type {
	split := strings.Split(d.CommonName, ":")
	if len(split) < 2 {
		// TODO: this shouldn't happen
		return ""
	}

	typeFromName := split[0]
	ok := knownSoftwareCertificateTypes[Type(typeFromName)]
	if !ok {
		// TODO: this should mean a Certificate for services (like Pass Type ID Certificate)
		return Type("")
	}

	return Type(typeFromName)
}

func (d Details) Platform() Platform {
	switch d.Type() {
	case AppleDevelopment, AppleDistribution:
		return All
	case iPhoneDeveloper, iPhoneDistribution:
		return IOS
	case MacDeveloper, ThirdPartyMacDeveloperApplication, ThirdPartyMacDeveloperInstaller, DeveloperIDApplication, DeveloperIDInstaller:
		return MacOS
	}

	// TODO: this should mean a Certificate for services (like Pass Type ID Certificate)
	return ""
}

func (d Details) String() string {
	team := fmt.Sprintf("%s (%s)", d.TeamName, d.TeamID)
	certInfo := fmt.Sprintf("Serial: %s, Name: %s, Team: %s, Expiry: %s", d.Serial, d.CommonName, team, d.EndDate)
	return certInfo
}
