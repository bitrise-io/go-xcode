package certificateutil

import "strings"

// Apple assigns code signing certificates a common name with a well-known prefix
// (developer/distribution) or substring (installer). These predicates classify a
// certificate by that common name so every consumer shares one definition instead
// of re-implementing the string matching.

// IsDevelopment reports whether the certificate is an Apple development signing
// certificate (e.g. "Apple Development", "iPhone Developer", "iOS Developer").
func (info CertificateInfoModel) IsDevelopment() bool {
	name := strings.ToLower(info.CommonName)
	return strings.HasPrefix(name, "apple development") ||
		strings.HasPrefix(name, "iphone developer") ||
		strings.HasPrefix(name, "ios developer")
}

// IsDistribution reports whether the certificate is an Apple distribution signing
// certificate (e.g. "Apple Distribution", "iPhone Distribution").
// See https://help.apple.com/xcode/mac/current/#/dev80c6204ec
func (info CertificateInfoModel) IsDistribution() bool {
	name := strings.ToLower(info.CommonName)
	return strings.HasPrefix(name, "apple distribution") ||
		strings.HasPrefix(name, "iphone distribution")
}

// IsInstaller reports whether the certificate is a macOS installer (.pkg) signing
// certificate (e.g. "3rd Party Mac Developer Installer", "Mac Installer Distribution").
func (info CertificateInfoModel) IsInstaller() bool {
	return strings.Contains(info.CommonName, "Installer")
}
