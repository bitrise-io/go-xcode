package exportoptions

import (
	"fmt"

	"howett.net/plist"
)

// AppStoreOptionsModel ...
type AppStoreOptionsModel struct {
	TeamID                             string
	BundleIDProvisioningProfileMapping map[string]string
	SigningCertificate                 string
	InstallerSigningCertificate        string
	SigningStyle                       string
	Destination                        string
	ICloudContainerEnvironment         ICloudContainerEnvironment
	DistributionBundleIdentifier       string

	// for app-store exports
	UploadBitcode bool
	UploadSymbols bool
	// Should Xcode manage the app's build number when uploading to App Store Connect? Defaults to YES.
	ManageAppVersion bool
}

// NewAppStoreOptions ...
func NewAppStoreOptions() AppStoreOptionsModel {
	return AppStoreOptionsModel{
		UploadBitcode:    UploadBitcodeDefault,
		UploadSymbols:    UploadSymbolsDefault,
		ManageAppVersion: manageAppVersionDefault,
	}
}

// Hash ...
func (options AppStoreOptionsModel) Hash() map[string]interface{} {
	hash := map[string]interface{}{}
	hash[MethodKey] = MethodAppStore
	if options.TeamID != "" {
		hash[TeamIDKey] = options.TeamID
	}
	if UploadBitcodeDefault != options.UploadBitcode {
		hash[UploadBitcodeKey] = options.UploadBitcode
	}
	if UploadSymbolsDefault != options.UploadSymbols {
		hash[UploadSymbolsKey] = options.UploadSymbols
	}
	if manageAppVersionDefault != options.ManageAppVersion {
		hash[manageAppVersionKey] = options.ManageAppVersion
	}
	if options.ICloudContainerEnvironment != "" {
		hash[ICloudContainerEnvironmentKey] = options.ICloudContainerEnvironment
	}
	if options.DistributionBundleIdentifier != "" {
		hash[DistributionBundleIdentifier] = options.DistributionBundleIdentifier
	}
	if len(options.BundleIDProvisioningProfileMapping) > 0 {
		hash[ProvisioningProfilesKey] = options.BundleIDProvisioningProfileMapping
	}
	if options.SigningCertificate != "" {
		hash[SigningCertificateKey] = options.SigningCertificate
	}
	if options.InstallerSigningCertificate != "" {
		hash[InstallerSigningCertificateKey] = options.InstallerSigningCertificate
	}
	if options.SigningStyle != "" {
		hash[SigningStyleKey] = options.SigningStyle
	}
	if options.Destination != "" {
		hash[DestinationKey] = options.Destination
	}
	return hash
}

// String ...
func (options AppStoreOptionsModel) String() (string, error) {
	hash := options.Hash()
	plistBytes, err := plist.MarshalIndent(hash, plist.XMLFormat, "\t")
	if err != nil {
		return "", fmt.Errorf("failed to marshal export options model, error: %s", err)
	}
	return string(plistBytes), err
}

// WriteToFile ...
func (options AppStoreOptionsModel) WriteToFile(pth string) error {
	return WritePlistToFile(options.Hash(), pth)
}

// WriteToTmpFile ...
func (options AppStoreOptionsModel) WriteToTmpFile() (string, error) {
	return WritePlistToTmpFile(options.Hash())
}
