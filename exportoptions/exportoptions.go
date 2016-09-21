package exportoptions

import (
	"fmt"
	"path/filepath"

	plist "github.com/DHowett/go-plist"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
)

// -----------------------
// --- Constants
// -----------------------

// CompileBitcodeKey ...
const CompileBitcodeKey = "compileBitcode"

// CompileBitcodeDefault ...
const CompileBitcodeDefault = true

// EmbedOnDemandResourcesAssetPacksInBundleKey ...
const EmbedOnDemandResourcesAssetPacksInBundleKey = "embedOnDemandResourcesAssetPacksInBundle"

// EmbedOnDemandResourcesAssetPacksInBundleDefault ...
const EmbedOnDemandResourcesAssetPacksInBundleDefault = true

// ICloudContainerEnvironmentKey ...
const ICloudContainerEnvironmentKey = "iCloudContainerEnvironment"
const (
	// ICloudContainerEnvironmentDevelopment ...
	ICloudContainerEnvironmentDevelopment ICloudContainerEnvironment = "Development"
	// ICloudContainerEnvironmentProduction ...
	ICloudContainerEnvironmentProduction ICloudContainerEnvironment = "Production"
	// ICloudContainerEnvironmentDefault ...
	ICloudContainerEnvironmentDefault ICloudContainerEnvironment = ICloudContainerEnvironmentDevelopment
)

// ManifestKey ...
const ManifestKey = "manifest"

// ManifestAppURLKey ...
const ManifestAppURLKey = "appURL"

// ManifestDisplayImageURLKey ...
const ManifestDisplayImageURLKey = "displayImageURL"

// ManifestFullSizeImageURLKey ...
const ManifestFullSizeImageURLKey = "fullSizeImageURL"

// ManifestAssetPackManifestURLKey ...
const ManifestAssetPackManifestURLKey = "assetPackManifestURL"

// MethodKey ...
const MethodKey = "method"
const (
	// MethodAppStore ...
	MethodAppStore Method = "app-store"
	// MethodAdHoc ...
	MethodAdHoc Method = "ad-hoc"
	// MethodPackage ...
	MethodPackage Method = "package"
	// MethodEnterprise ...
	MethodEnterprise Method = "enterprise"
	// MethodDevelopment ...
	MethodDevelopment Method = "development"
	// MethodDeveloperID ...
	MethodDeveloperID Method = "developer-id"
	// MethodDefault ...
	MethodDefault Method = MethodDevelopment
)

// OnDemandResourcesAssetPacksBaseURLKey ....
const OnDemandResourcesAssetPacksBaseURLKey = "onDemandResourcesAssetPacksBaseURL"

// TeamIDKey ...
const TeamIDKey = "teamID"

// ThinningKey ...
const ThinningKey = "thinning"
const (
	// ThinningNone ...
	ThinningNone = "none"
	// ThinningThinForAllVariants ...
	ThinningThinForAllVariants = "thin-for-all-variants"
	// ThinningDefault ...
	ThinningDefault = ThinningNone
)

// UploadBitcodeKey ....
const UploadBitcodeKey = "uploadBitcode"

// UploadBitcodeDefault ...
const UploadBitcodeDefault = true

// UploadSymbolsKey ...
const UploadSymbolsKey = "uploadSymbols"

// UploadSymbolsDefault ...
const UploadSymbolsDefault = true

// -----------------------
// --- Models
// -----------------------

// ICloudContainerEnvironment ...
type ICloudContainerEnvironment string

// Method ...
type Method string

// Manifest ...
type Manifest struct {
	AppURL               string
	DisplayImageURL      string
	FullSizeImageURL     string
	AssetPackManifestURL string
}

// IsEmpty ...
func (manifest Manifest) IsEmpty() bool {
	return (manifest.AppURL == "" && manifest.DisplayImageURL == "" && manifest.FullSizeImageURL == "" && manifest.AssetPackManifestURL == "")
}

// ToHash ...
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

// ExportOptions ...
type ExportOptions interface {
	ToHash() map[string]interface{}
	WriteToFile(pth string) error
	WriteToTmpFile() (string, error)
}

// AppStoreOptionsModel ...
type AppStoreOptionsModel struct {
	TeamID string

	// for app-store exports
	UploadBitcode bool
	UploadSymbols bool
}

// NewAppStoreOptions ...
func NewAppStoreOptions() AppStoreOptionsModel {
	return AppStoreOptionsModel{
		UploadBitcode: UploadBitcodeDefault,
		UploadSymbols: UploadSymbolsDefault,
	}
}

// ToHash ...
func (options AppStoreOptionsModel) ToHash() map[string]interface{} {
	hash := map[string]interface{}{}
	hash[MethodKey] = MethodAppStore
	if options.TeamID != "" {
		hash[TeamIDKey] = options.TeamID
	}
	if options.UploadBitcode != UploadBitcodeDefault {
		hash[UploadBitcodeKey] = options.UploadBitcode
	}
	if options.UploadSymbols != UploadSymbolsDefault {
		hash[UploadSymbolsKey] = options.UploadSymbols
	}
	return hash
}

// WriteToFile ...
func (options AppStoreOptionsModel) WriteToFile(pth string) error {
	return WritePlistToFile(options.ToHash(), pth)
}

// WriteToTmpFile ...
func (options AppStoreOptionsModel) WriteToTmpFile() (string, error) {
	return WritePlistToTmpFile(options.ToHash())
}

// NonAppStoreOptionsModel ...
type NonAppStoreOptionsModel struct {
	Method Method
	TeamID string

	// for non app-store exports
	CompileBitcode                           bool
	EmbedOnDemandResourcesAssetPacksInBundle bool
	ICloudContainerEnvironment               ICloudContainerEnvironment
	Manifest                                 Manifest
	OnDemandResourcesAssetPacksBaseURL       string
	Thinning                                 string
}

// NewNonAppStoreOptions ...
func NewNonAppStoreOptions(method Method) NonAppStoreOptionsModel {
	return NonAppStoreOptionsModel{
		Method:                                   method,
		CompileBitcode:                           CompileBitcodeDefault,
		EmbedOnDemandResourcesAssetPacksInBundle: EmbedOnDemandResourcesAssetPacksInBundleDefault,
		ICloudContainerEnvironment:               ICloudContainerEnvironmentDefault,
		Thinning:                                 ThinningDefault,
	}
}

// ToHash ...
func (options NonAppStoreOptionsModel) ToHash() map[string]interface{} {
	hash := map[string]interface{}{}
	if options.Method != "" {
		hash[MethodKey] = options.Method
	}
	if options.TeamID != "" {
		hash[TeamIDKey] = options.TeamID
	}
	if options.CompileBitcode != CompileBitcodeDefault {
		hash[CompileBitcodeKey] = options.CompileBitcode
	}
	if options.EmbedOnDemandResourcesAssetPacksInBundle != EmbedOnDemandResourcesAssetPacksInBundleDefault {
		hash[EmbedOnDemandResourcesAssetPacksInBundleKey] = options.EmbedOnDemandResourcesAssetPacksInBundle
	}
	if options.ICloudContainerEnvironment != ICloudContainerEnvironmentDefault {
		hash[ICloudContainerEnvironmentKey] = options.ICloudContainerEnvironment
	}
	if !options.Manifest.IsEmpty() {
		hash[ManifestKey] = options.Manifest.ToHash()
	}
	if options.OnDemandResourcesAssetPacksBaseURL != "" {
		hash[OnDemandResourcesAssetPacksBaseURLKey] = options.OnDemandResourcesAssetPacksBaseURL
	}
	if options.Thinning != ThinningDefault {
		hash[ThinningKey] = options.Thinning
	}
	return hash
}

// WriteToFile ...
func (options NonAppStoreOptionsModel) WriteToFile(pth string) error {
	return WritePlistToFile(options.ToHash(), pth)
}

// WriteToTmpFile ...
func (options NonAppStoreOptionsModel) WriteToTmpFile() (string, error) {
	return WritePlistToTmpFile(options.ToHash())
}

// WritePlistToFile ...
func WritePlistToFile(options map[string]interface{}, pth string) error {
	plistBytes, err := plist.MarshalIndent(options, plist.XMLFormat, "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal export options model, error: %s", err)
	}
	if err := fileutil.WriteBytesToFile(pth, plistBytes); err != nil {
		return fmt.Errorf("failed to write export options, error: %s", err)
	}

	return nil
}

// WritePlistToTmpFile ...
func WritePlistToTmpFile(options map[string]interface{}) (string, error) {
	tmpDir, err := pathutil.NormalizedOSTempDirPath("output")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir, error: %s", err)
	}
	pth := filepath.Join(tmpDir, "exportOptions.plist")

	if err := WritePlistToFile(options, pth); err != nil {
		return "", fmt.Errorf("failed to write to file options, error: %s", err)
	}

	return pth, nil
}
