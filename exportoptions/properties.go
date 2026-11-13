package exportoptions

import "fmt"

const CompileBitcodeKey = "compileBitcode"

const CompileBitcodeDefault = true

const EmbedOnDemandResourcesAssetPacksInBundleKey = "embedOnDemandResourcesAssetPacksInBundle"

const EmbedOnDemandResourcesAssetPacksInBundleDefault = true

const ICloudContainerEnvironmentKey = "iCloudContainerEnvironment"

type ICloudContainerEnvironment string

const (
	ICloudContainerEnvironmentDevelopment ICloudContainerEnvironment = "Development"
	ICloudContainerEnvironmentProduction  ICloudContainerEnvironment = "Production"
)

const DistributionBundleIdentifier = "distributionBundleIdentifier"

const ManifestKey = "manifest"

const ManifestAppURLKey = "appURL"

const ManifestDisplayImageURLKey = "displayImageURL"

const ManifestFullSizeImageURLKey = "fullSizeImageURL"

const ManifestAssetPackManifestURLKey = "assetPackManifestURL"

type Manifest struct {
	AppURL               string
	DisplayImageURL      string
	FullSizeImageURL     string
	AssetPackManifestURL string
}

func (manifest Manifest) IsEmpty() bool {
	return (manifest.AppURL == "" && manifest.DisplayImageURL == "" && manifest.FullSizeImageURL == "" && manifest.AssetPackManifestURL == "")
}

func (manifest Manifest) ToHash() map[string]string {
	hash := map[string]string{}
	if manifest.AppURL != "" {
		hash[ManifestAppURLKey] = manifest.AppURL
	}
	if manifest.DisplayImageURL != "" {
		hash[ManifestDisplayImageURLKey] = manifest.DisplayImageURL
	}
	if manifest.FullSizeImageURL != "" {
		hash[ManifestFullSizeImageURLKey] = manifest.FullSizeImageURL
	}
	if manifest.AssetPackManifestURL != "" {
		hash[ManifestAssetPackManifestURLKey] = manifest.AssetPackManifestURL
	}
	return hash
}

const MethodKey = "method"

type Method string

const (
	// MethodAppStore is deprecated since Xcode 15.3, its new name is MethodAppStoreConnect
	MethodAppStore Method = "app-store"
	// MethodAdHoc is deprecated since Xcode 15.3, its new name is MethodReleaseTesting
	MethodAdHoc      Method = "ad-hoc"
	MethodPackage    Method = "package"
	MethodEnterprise Method = "enterprise"
	// MethodDevelopment is deprecated since Xcode 15.3, its new name is MethodDebugging
	MethodDevelopment Method = "development"
	MethodDeveloperID Method = "developer-id"
	// MethodDebugging is the new name for MethodDevelopment since Xcode 15.3
	MethodDebugging Method = "debugging"
	// MethodAppStoreConnect is the new name for MethodAppStore since Xcode 15.3
	MethodAppStoreConnect Method = "app-store-connect"
	// MethodReleaseTesting is the new name for MethodAdHoc since Xcode 15.3
	MethodReleaseTesting Method = "release-testing"
	MethodDefault        Method = MethodDevelopment
)

func (m Method) IsAppStore() bool {
	return m == MethodAppStore || m == MethodAppStoreConnect
}

func (m Method) IsAdHoc() bool {
	return m == MethodAdHoc || m == MethodReleaseTesting
}

func (m Method) IsDevelopment() bool {
	return m == MethodDevelopment || m == MethodDebugging
}

func (m Method) IsEnterprise() bool {
	return m == MethodEnterprise
}

// ParseMethod parses Step input and returns the corresponding Method.
func ParseMethod(method string) (Method, error) {
	switch method {
	case "app-store":
		return MethodAppStore, nil
	case "ad-hoc":
		return MethodAdHoc, nil
	case "package":
		return MethodPackage, nil
	case "enterprise":
		return MethodEnterprise, nil
	case "development":
		return MethodDevelopment, nil
	case "developer-id":
		return MethodDeveloperID, nil
	default:
		return Method(""), fmt.Errorf("unkown method (%s)", method)
	}
}

// UpgradeToXcode15_3MethodName replaces the legacy export method strings with the ones available in Xcode 15.3 and later.
func UpgradeToXcode15_3MethodName(method Method) Method {
	switch method {
	case MethodAppStore:
		return MethodAppStoreConnect
	case MethodAdHoc:
		return MethodReleaseTesting
	case MethodDevelopment:
		return MethodDebugging
	default:
		return method
	}
}

const OnDemandResourcesAssetPacksBaseURLKey = "onDemandResourcesAssetPacksBaseURL"

const TeamIDKey = "teamID"

const ThinningKey = "thinning"

const (
	ThinningNone               = "none"
	ThinningThinForAllVariants = "thin-for-all-variants"
	ThinningDefault            = ThinningNone
)

const UploadBitcodeKey = "uploadBitcode"

const UploadBitcodeDefault = true

const UploadSymbolsKey = "uploadSymbols"

const UploadSymbolsDefault = true

const (
	manageAppVersionKey     = "manageAppVersionAndBuildNumber"
	manageAppVersionDefault = true
)

const ProvisioningProfilesKey = "provisioningProfiles"

const SigningCertificateKey = "signingCertificate"

const InstallerSigningCertificateKey = "installerSigningCertificate"

const SigningStyleKey = "signingStyle"

type SigningStyle string

const (
	SigningStyleManual    SigningStyle = "manual"
	SigningStyleAutomatic SigningStyle = "automatic"
)

const DestinationKey = "destination"

const TestFlightInternalTestingOnlyDefault = false

const TestFlightInternalTestingOnlyKey = "testFlightInternalTestingOnly"

type Destination string

const (
	DestinationExport  Destination = "export"
	DestinationDefault Destination = DestinationExport
)
