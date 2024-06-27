package xcode_15_4

import (
	"fmt"

	"github.com/bitrise-io/go-plist"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pointers"
)

type Options struct {
	// Determines whether the app is exported locally or uploaded to Apple.
	// Options are export or upload.
	// The available options vary based on the selected distribution method. Defaults to export.
	Destination Destination `plist:"destination,omitempty"`

	// Reformat archive to focus on eligible target bundle identifier.
	// Defaults to top level bundle identifier.
	DistributionBundleIdentifier string `plist:"distributionBundleIdentifier,omitempty"`

	// If the app is using CloudKit, this configures the "com.apple.developer.icloud-container-environment" entitlement.
	// Available options vary depending on the type of provisioning profile used, but may include: Development and Production.
	// If not specified, this defaults to Development when development signing or Production when distribution signing.
	ICloudContainerEnvironment ICloudContainerEnvironment `plist:"iCloudContainerEnvironment,omitempty"`

	// Describes how Xcode should export the archive.
	// Available options: app-store-connect, release-testing, enterprise, debugging, developer-id, mac-application, validation, and package.
	// The list of options varies based on the type of archive.
	// Defaults to debugging.
	// Additional options include app-store (deprecated: use app-store-connect), ad-hoc (deprecated: use release-testing), and development (deprecated: use debugging).
	Method Method `plist:"method,omitempty"`

	// The signing style to use when re-signing the app for distribution.
	// Options are manual or automatic.
	// Apps that were automatically signed when archived default to automatic.
	// Apps that were manually signed when archived default to manual.
	// If your archive is manually signed and you choose to automatically sign when distributing,
	// then Xcode will create provisioning profiles and managed cloud signing certificates as necessary, but will not register devices, register app IDs, or modify app ID settings.
	SigningStyle SigningStyle `plist:"signingStyle,omitempty"`

	// Should symbols be stripped from Swift libraries in your IPA? Defaults to YES.
	StripSwiftSymbols *bool `plist:"stripSwiftSymbols,omitempty"`

	// The Developer team to use for this export. Defaults to the team used to build the archive.
	TeamID string `plist:"teamID,omitempty"`

	// ------------------------
	// For manual signing only.

	// For manual signing only.
	// Specify the provisioning profile to use for each executable in your app.
	// Keys in this dictionary are either the bundle identifiers of executables or their paths relative to the Archive's Products directory;
	// values are the provisioning profile name or UUID to use.
	ProvisioningProfiles map[string]string `plist:"provisioningProfiles,omitempty"`

	// For manual signing only.
	// Provide a certificate name, SHA-1 hash, or automatic selector to use for signing.
	// Automatic selectors allow Xcode to pick the newest installed certificate of a particular type.
	// The available automatic selectors are "Mac App Distribution", "iOS Distribution", "iOS Developer", "Developer ID Application", "Apple Distribution", "Mac Developer", and "Apple Development".
	// Defaults to an automatic certificate selector matching the current distribution method.
	SigningCertificate string `plist:"signingCertificate,omitempty"`

	// ---------------------------
	// For App Store exports only.

	// For App Store exports, should Xcode generate App Store Information for uploading with iTMSTransporter?
	// Defaults to NO.
	GenerateAppStoreInformation *bool `plist:"generateAppStoreInformation,omitempty"`

	// Should Xcode manage the app's build number when uploading to App Store Connect? Defaults to YES.
	ManageAppVersionAndBuildNumber *bool `plist:"manageAppVersionAndBuildNumber,omitempty"`

	// When enabled, this build cannot be distributed via external TestFlight or the App Store.
	// This is recommended for pull requests, development branches, and other builds that are not suitable for external distribution.
	TestFlightInternalTestingOnly *bool `plist:"testFlightInternalTestingOnly,omitempty"`

	// For App Store exports, should the package include symbols? Defaults to YES.
	UploadSymbols *bool `plist:"uploadSymbols,omitempty"`

	// ------------------------------
	// For non-App Store exports only.

	// For non-App Store exports, should Xcode re-compile the app from bitcode? Defaults to YES.
	CompileBitcode *bool `plist:"compileBitcode,omitempty"`

	// For non-App Store exports, if the app uses On Demand Resources and this is YES,
	// asset packs are embedded in the app bundle so that the app can be tested without a server to host asset packs.
	// Defaults to YES unless onDemandResourcesAssetPacksBaseURL is specified.
	EmbedOnDemandResourcesAssetPacksInBundle *bool `plist:"embedOnDemandResourcesAssetPacksInBundle,omitempty"`

	// For non-App Store exports, users can download your app over the web by opening your distribution manifest file in a web browser.
	// To generate a distribution manifest, the value of this key should be a dictionary with three sub-keys: appURL, displayImageURL, fullSizeImageURL.
	// The additional sub-key assetPackManifestURL is required when using on-demand resources.
	Manifest map[string]string `plist:"manifest,omitempty"`

	// For non-App Store exports, if the app uses On Demand Resources and embedOnDemandResourcesAssetPacksInBundle isn't YES,
	//  this should be a base URL specifying where asset packs are going to be hosted.
	//  This configures the app to download asset packs from the specified URL.
	OnDemandResourcesAssetPacksBaseURL string `plist:"onDemandResourcesAssetPacksBaseURL,omitempty"`

	// For non-App Store exports, should Xcode thin the package for one or more device variants?
	// Available options: <none> (Xcode produces a non-thinned universal app), <thin-for-all-variants> (Xcode produces a universal app and all available thinned variants),
	// or a model identifier for a specific device (e.g. "iPhone7,1"). Defaults to <none>.
	Thinning Thinning `plist:"thinning,omitempty"`
}

type OptionalOptions struct {
	DistributionBundleIdentifier string
	ICloudContainerEnvironment   ICloudContainerEnvironment
	StripSwiftSymbols            *bool
}

type AppStoreOptionalOptions struct {
	OptionalOptions
	GenerateAppStoreInformation    *bool
	ManageAppVersionAndBuildNumber *bool
	TestFlightInternalTestingOnly  *bool
	UploadSymbols                  *bool
}

type NonAppStoreOptionalOptions struct {
	OptionalOptions
	CompileBitcode                           *bool
	EmbedOnDemandResourcesAssetPacksInBundle *bool
	Manifest                                 map[string]string
	OnDemandResourcesAssetPacksBaseURL       string
	Thinning                                 Thinning
}

type Generator struct {
	options Options
}

func NewGeneratorForAppStoreExports(teamID string, optionals AppStoreOptionalOptions) Generator {
	generator := newGenerator(MethodAppStoreConnect, teamID, optionals.OptionalOptions)

	if optionals.ManageAppVersionAndBuildNumber != nil {
		generator.options.ManageAppVersionAndBuildNumber = optionals.ManageAppVersionAndBuildNumber
	} else {
		generator.options.ManageAppVersionAndBuildNumber = pointers.NewBoolPtr(false)
	}

	if optionals.UploadSymbols != nil {
		generator.options.UploadSymbols = optionals.UploadSymbols
	} else {
		generator.options.UploadSymbols = pointers.NewBoolPtr(true)
	}

	if optionals.GenerateAppStoreInformation != nil {
		generator.options.GenerateAppStoreInformation = optionals.GenerateAppStoreInformation
	}
	if optionals.TestFlightInternalTestingOnly != nil {
		generator.options.TestFlightInternalTestingOnly = optionals.TestFlightInternalTestingOnly
	}
	return generator
}

func NewGeneratorForNonAppStoreExports(method Method, teamID string, optionals NonAppStoreOptionalOptions) Generator {
	generator := newGenerator(method, teamID, optionals.OptionalOptions)

	if optionals.CompileBitcode != nil {
		generator.options.CompileBitcode = optionals.CompileBitcode
	} else {
		generator.options.CompileBitcode = pointers.NewBoolPtr(false)
	}

	if optionals.Thinning != "" {
		generator.options.Thinning = optionals.Thinning
	} else {
		generator.options.Thinning = ThinningNone
	}

	if optionals.EmbedOnDemandResourcesAssetPacksInBundle != nil {
		generator.options.EmbedOnDemandResourcesAssetPacksInBundle = optionals.EmbedOnDemandResourcesAssetPacksInBundle
	}
	if optionals.Manifest != nil {
		generator.options.Manifest = optionals.Manifest
	}
	if optionals.OnDemandResourcesAssetPacksBaseURL != "" {
		generator.options.OnDemandResourcesAssetPacksBaseURL = optionals.OnDemandResourcesAssetPacksBaseURL
	}

	return generator
}

func (g Generator) GenerateForManualSigning(signingCertificate string, provisioningProfiles map[string]string) Options {
	g.options.SigningStyle = SigningStyleManual
	g.options.SigningCertificate = signingCertificate
	g.options.ProvisioningProfiles = provisioningProfiles
	return g.options
}

func (g Generator) GenerateForAutomaticSigning() Options {
	g.options.SigningStyle = SigningStyleAutomatic
	return g.options
}

func WriteExportOptionsToFile(options Options, pth string) error {
	plistBytes, err := marshalExportOptions(options)
	if err != nil {
		return err
	}
	if err := fileutil.WriteBytesToFile(pth, plistBytes); err != nil {
		return fmt.Errorf("failed to write export options to file: %s", err)
	}

	return nil
}

func newGenerator(method Method, teamID string, optionals OptionalOptions) Generator {
	options := Options{
		Destination: DestinationExport,
		Method:      method,
		TeamID:      teamID,
	}

	if optionals.StripSwiftSymbols != nil {
		options.StripSwiftSymbols = optionals.StripSwiftSymbols
	} else {
		options.StripSwiftSymbols = pointers.NewBoolPtr(true)
	}

	if optionals.DistributionBundleIdentifier != "" {
		options.DistributionBundleIdentifier = optionals.DistributionBundleIdentifier
	}
	if optionals.ICloudContainerEnvironment != "" {
		options.ICloudContainerEnvironment = optionals.ICloudContainerEnvironment
	}

	return Generator{options: options}
}

func unmarshalExportOptions(b []byte) (Options, error) {
	var options Options
	if _, err := plist.Unmarshal(b, &options); err != nil {
		return Options{}, fmt.Errorf("failed to unmarshal export options: %s", err)
	}
	return options, nil
}

func marshalExportOptions(options Options) ([]byte, error) {
	plistBytes, err := plist.MarshalIndent(options, plist.XMLFormat, "\t")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal export options: %s", err)
	}
	return plistBytes, nil
}
