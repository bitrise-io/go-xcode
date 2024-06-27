package exportoptionsgenerator

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/sliceutil"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/bitrise-io/go-xcode/export"
	"github.com/bitrise-io/go-xcode/exportoptions"
	"github.com/bitrise-io/go-xcode/plistutil"
	"github.com/bitrise-io/go-xcode/profileutil"
	xcode154Options "github.com/bitrise-io/go-xcode/v2/exportoptions/xcode_15_4"
	"github.com/bitrise-io/go-xcode/xcodeproject/serialized"
	"github.com/bitrise-io/go-xcode/xcodeproject/xcodeproj"
	"github.com/bitrise-io/go-xcode/xcodeproject/xcscheme"
)

// const for AppClipProductType and manualSigningStyle
const (
	AppClipProductType = "com.apple.product-type.application.on-demand-install-capable"
	manualSigningStyle = "manual"
)

type ExportOptionsGenerator struct {
	xcodeProj     *xcodeproj.XcodeProj
	scheme        *xcscheme.Scheme
	configuration string

	certificateProvider CodesignIdentityProvider
	profileProvider     ProvisioningProfileProvider
	targetInfoProvider  TargetInfoProvider
	logger              log.Logger
}

func New(xcodeProj *xcodeproj.XcodeProj, scheme *xcscheme.Scheme, configuration string, logger log.Logger) ExportOptionsGenerator {
	g := ExportOptionsGenerator{
		xcodeProj:     xcodeProj,
		scheme:        scheme,
		configuration: configuration,
	}
	g.certificateProvider = LocalCodesignIdentityProvider{}
	g.profileProvider = LocalProvisioningProfileProvider{}
	g.targetInfoProvider = XcodebuildTargetInfoProvider{xcodeProj: xcodeProj}
	g.logger = logger
	return g
}

type XcodeVersion struct {
	Major int
	Minor int
	Patch int
}

func (g ExportOptionsGenerator) GenerateApplicationExportOptionsAt(
	exportMethod exportoptions.Method,
	containerEnvironment string,
	teamID string,
	uploadBitcode bool,
	compileBitcode bool,
	xcodeManaged bool,
	xcodeVersion XcodeVersion,
	signingStyle string,
	pth string,
) error {
	g.logger.TDebugf("Generating application export options for: %s", exportMethod)

	//
	// Gather information for the export options
	mainTarget, err := ArchivableApplicationTarget(g.xcodeProj, g.scheme)
	if err != nil {
		return err
	}

	dependentTargets := filterApplicationBundleTargets(
		g.xcodeProj.DependentTargetsOfTarget(*mainTarget),
		exportMethod,
	)
	targets := append([]xcodeproj.Target{*mainTarget}, dependentTargets...)

	var mainTargetBundleID string
	entitlementsByBundleID := map[string]plistutil.PlistData{}
	for i, target := range targets {
		bundleID, err := g.targetInfoProvider.TargetBundleID(target.Name, g.configuration)
		if err != nil {
			return fmt.Errorf("failed to get target (%s) bundle id: %s", target.Name, err)
		}

		entitlements, err := g.targetInfoProvider.TargetCodeSignEntitlements(target.Name, g.configuration)
		if err != nil && !serialized.IsKeyNotFoundError(err) {
			return fmt.Errorf("failed to get target (%s) bundle id: %s", target.Name, err)
		}

		entitlementsByBundleID[bundleID] = plistutil.PlistData(entitlements)

		if i == 0 {
			mainTargetBundleID = bundleID
		}
	}

	// IcloudContainerEnvironment
	iCloudContainerEnvironment, err := determineIcloudContainerEnvironment(containerEnvironment, entitlementsByBundleID, exportMethod, xcodeVersion.Major)
	if err != nil {
		return err
	}

	// Profiles by Bundle ID
	exportProfileMapping := map[string]string{}
	var certificate certificateutil.CertificateInfoModel
	if signingStyle == "manual" {
		codeSignGroup, err := g.determineCodesignGroup(entitlementsByBundleID, exportMethod, teamID, xcodeManaged)
		if err != nil {
			return err
		}
		if codeSignGroup == nil {
			return fmt.Errorf("failed to determine code sign group")
		}

		for bundleID, profileInfo := range codeSignGroup.BundleIDProfileMap() {
			exportProfileMapping[bundleID] = profileInfo.Name
		}

		certificate = codeSignGroup.Certificate()
		teamID = certificate.TeamID
	}

	//
	// Generate Export options
	switch {
	case xcodeVersion.Major >= 15 && xcodeVersion.Minor >= 4:
		xcode154Options.NewGeneratorForAppStoreExports(teamID, xcode154Options.AppStoreOptionalOptions{
			OptionalOptions: xcode154Options.OptionalOptions{
				DistributionBundleIdentifier: mainTargetBundleID,
				ICloudContainerEnvironment:   iCloudContainerEnvironment,
				StripSwiftSymbols:            nil,
			},
			GenerateAppStoreInformation:    nil,
			ManageAppVersionAndBuildNumber: nil,
			TestFlightInternalTestingOnly:  nil,
			UploadSymbols:                  nil,
		})
	default:
		options, err := g.generateExportOptions(
			exportMethod,
			iCloudContainerEnvironment,
			uploadBitcode,
			compileBitcode,
			xcodeManaged,
			signingStyle,
			teamID,
			certificate.CommonName,
			exportProfileMapping,
			xcodeVersion.Major,
			mainTargetBundleID,
		)
		if err != nil {
			return err
		}

		if err := exportoptions.WritePlistToFile(options.Hash(), pth); err != nil {
			return err
		}
	}

	return nil
}

//// GenerateApplicationExportOptions generates exportOptions for an application export.
//func (g ExportOptionsGenerator) GenerateApplicationExportOptions(exportMethod exportoptions.Method, containerEnvironment string, teamID string, uploadBitcode bool, compileBitcode bool, xcodeManaged bool,
//	xcodeMajorVersion int64) (exportoptions.Options, error) {
//
//	g.logger.TDebugf("Generating application export options for: %s", exportMethod)
//
//	mainTarget, err := ArchivableApplicationTarget(g.xcodeProj, g.scheme)
//	if err != nil {
//		return nil, err
//	}
//
//	dependentTargets := filterApplicationBundleTargets(
//		g.xcodeProj.DependentTargetsOfTarget(*mainTarget),
//		exportMethod,
//	)
//	targets := append([]xcodeproj.Target{*mainTarget}, dependentTargets...)
//
//	var mainTargetBundleID string
//	entitlementsByBundleID := map[string]plistutil.PlistData{}
//	for i, target := range targets {
//		bundleID, err := g.targetInfoProvider.TargetBundleID(target.Name, g.configuration)
//		if err != nil {
//			return exportoptions.Options{}, fmt.Errorf("failed to get target (%s) bundle id: %s", target.Name, err)
//		}
//
//		entitlements, err := g.targetInfoProvider.TargetCodeSignEntitlements(target.Name, g.configuration)
//		if err != nil && !serialized.IsKeyNotFoundError(err) {
//			return exportoptions.Options{}, fmt.Errorf("failed to get target (%s) bundle id: %s", target.Name, err)
//		}
//
//		entitlementsByBundleID[bundleID] = plistutil.PlistData(entitlements)
//
//		if i == 0 {
//			mainTargetBundleID = bundleID
//		}
//	}
//
//	g.logger.TDebugf("Generated application export options plist for: %s", exportMethod)
//
//	return g.generateExportOptions(exportMethod, containerEnvironment, teamID, uploadBitcode, compileBitcode,
//		xcodeManaged, entitlementsByBundleID, xcodeMajorVersion, mainTargetBundleID)
//}

// generateExportOptions generates an exportOptions based on the provided conditions.
func (g ExportOptionsGenerator) generateExportOptions(
	exportMethod exportoptions.Method,
	iCloudContainerEnvironment string,
	uploadBitcode bool,
	compileBitcode bool,
	xcodeManaged bool,
	exportCodeSignStyle string,
	teamID string,
	certificateCommonName string,
	exportProfileMapping map[string]string,
	xcodeMajorVersion int,
	distributionBundleIdentifier string,
) (exportoptions.ExportOptions, error) {
	exportOpts := generateBaseExportOptions(exportMethod, teamID, uploadBitcode, compileBitcode, iCloudContainerEnvironment)
	if xcodeMajorVersion >= 12 {
		exportOpts = addDistributionBundleIdentifierFromXcode12(exportOpts, distributionBundleIdentifier)
	}
	if xcodeMajorVersion >= 13 {
		exportOpts = disableManagedBuildNumberFromXcode13(exportOpts)
	}

	shouldSetManualSigning := xcodeManaged && exportCodeSignStyle == manualSigningStyle
	if shouldSetManualSigning {
		g.logger.Warnf("App was signed with Xcode managed profile when archiving,")
		g.logger.Warnf("ipa export uses manual code signing.")
		g.logger.Warnf(`Setting "signingStyle" to "manual".`)
	}

	g.logger.TDebugf("Determined code signing style")

	switch options := exportOpts.(type) {
	case exportoptions.AppStoreOptionsModel:
		options.BundleIDProvisioningProfileMapping = exportProfileMapping
		options.SigningCertificate = certificateCommonName
		if shouldSetManualSigning {
			options.SigningStyle = manualSigningStyle
		}
		exportOpts = options
	case exportoptions.NonAppStoreOptionsModel:
		options.BundleIDProvisioningProfileMapping = exportProfileMapping
		options.SigningCertificate = certificateCommonName
		if shouldSetManualSigning {
			options.SigningStyle = manualSigningStyle
		}
		exportOpts = options
	}

	return exportOpts, nil
}

// GetDefaultProvisioningProfile ...
func (g ExportOptionsGenerator) GetDefaultProvisioningProfile() (profileutil.ProvisioningProfileInfoModel, error) {
	defaultProfileURL := os.Getenv("BITRISE_DEFAULT_PROVISION_URL")
	if defaultProfileURL == "" {
		return profileutil.ProvisioningProfileInfoModel{}, nil
	}

	tmpDir, err := pathutil.NormalizedOSTempDirPath("tmp_default_profile")
	if err != nil {
		return profileutil.ProvisioningProfileInfoModel{}, err
	}

	tmpDst := filepath.Join(tmpDir, "default.mobileprovision")
	tmpDstFile, err := os.Create(tmpDst)
	if err != nil {
		return profileutil.ProvisioningProfileInfoModel{}, err
	}
	defer func() {
		if err := tmpDstFile.Close(); err != nil {
			g.logger.Errorf("Failed to close file (%s), error: %s", tmpDst, err)
		}
	}()

	response, err := http.Get(defaultProfileURL)
	if err != nil {
		return profileutil.ProvisioningProfileInfoModel{}, err
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			g.logger.Errorf("Failed to close response body, error: %s", err)
		}
	}()

	if _, err := io.Copy(tmpDstFile, response.Body); err != nil {
		return profileutil.ProvisioningProfileInfoModel{}, err
	}

	defaultProfile, err := profileutil.NewProvisioningProfileInfoFromFile(tmpDst)
	if err != nil {
		return profileutil.ProvisioningProfileInfoModel{}, err
	}

	return defaultProfile, nil
}

// determineCodesignGroup finds the best codesign group (certificate + profiles)
// based on the installed Provisioning Profiles and Codesign Certificates.
func (g ExportOptionsGenerator) determineCodesignGroup(bundleIDEntitlementsMap map[string]plistutil.PlistData, exportMethod exportoptions.Method, teamID string, allowXcodeManagedProfiles bool) (*export.IosCodeSignGroup, error) {
	fmt.Println()
	g.logger.Printf("Target Bundle ID - Entitlements map")
	var bundleIDs []string
	for bundleID, entitlements := range bundleIDEntitlementsMap {
		bundleIDs = append(bundleIDs, bundleID)

		var entitlementKeys []string
		for key := range entitlements {
			entitlementKeys = append(entitlementKeys, key)
		}
		g.logger.Printf("%s: %s", bundleID, entitlementKeys)
	}

	fmt.Println()
	g.logger.Printf("Resolving CodeSignGroups...")

	certs, err := g.certificateProvider.ListCodesignIdentities()
	if err != nil {
		return nil, fmt.Errorf("Failed to get installed certificates, error: %s", err)
	}

	g.logger.Debugf("Installed certificates:")
	for _, certInfo := range certs {
		g.logger.Debugf(certInfo.String())
	}

	profs, err := g.profileProvider.ListProvisioningProfiles()
	if err != nil {
		return nil, fmt.Errorf("Failed to get installed provisioning profiles, error: %s", err)
	}

	g.logger.Debugf("Installed profiles:")
	for _, profileInfo := range profs {
		g.logger.Debugf(profileInfo.String(certs...))
	}

	g.logger.Printf("Resolving CodeSignGroups...")
	codeSignGroups := export.CreateSelectableCodeSignGroups(certs, profs, bundleIDs)
	if len(codeSignGroups) == 0 {
		g.logger.Errorf("Failed to find code signing groups for specified export method (%s)", exportMethod)
	}

	g.logger.Debugf("\nGroups:")
	for _, group := range codeSignGroups {
		g.logger.Debugf(group.String())
	}

	if len(bundleIDEntitlementsMap) > 0 {
		g.logger.Warnf("Filtering CodeSignInfo groups for target capabilities")

		codeSignGroups = export.FilterSelectableCodeSignGroups(codeSignGroups, export.CreateEntitlementsSelectableCodeSignGroupFilter(bundleIDEntitlementsMap))

		g.logger.Debugf("\nGroups after filtering for target capabilities:")
		for _, group := range codeSignGroups {
			g.logger.Debugf(group.String())
		}
	}

	g.logger.Warnf("Filtering CodeSignInfo groups for export method")

	codeSignGroups = export.FilterSelectableCodeSignGroups(codeSignGroups, export.CreateExportMethodSelectableCodeSignGroupFilter(exportMethod))

	g.logger.Debugf("\nGroups after filtering for export method:")
	for _, group := range codeSignGroups {
		g.logger.Debugf(group.String())
	}

	if teamID != "" {
		g.logger.Warnf("ExportDevelopmentTeam specified: %s, filtering CodeSignInfo groups...", teamID)

		codeSignGroups = export.FilterSelectableCodeSignGroups(codeSignGroups, export.CreateTeamSelectableCodeSignGroupFilter(teamID))

		g.logger.Debugf("\nGroups after filtering for team ID:")
		for _, group := range codeSignGroups {
			g.logger.Debugf(group.String())
		}
	}

	if !allowXcodeManagedProfiles {
		g.logger.Warnf("App was signed with NON Xcode managed profile when archiving,\n" +
			"only NOT Xcode managed profiles are allowed to sign when exporting the archive.\n" +
			"Removing Xcode managed CodeSignInfo groups")

		codeSignGroups = export.FilterSelectableCodeSignGroups(codeSignGroups, export.CreateNotXcodeManagedSelectableCodeSignGroupFilter())

		g.logger.Debugf("\nGroups after filtering for NOT Xcode managed profiles:")
		for _, group := range codeSignGroups {
			g.logger.Debugf(group.String())
		}
	}

	defaultProfileURL := os.Getenv("BITRISE_DEFAULT_PROVISION_URL")
	if teamID == "" && defaultProfileURL != "" {
		if defaultProfile, err := g.GetDefaultProvisioningProfile(); err == nil {
			g.logger.Debugf("\ndefault profile: %v\n", defaultProfile)
			filteredCodeSignGroups := export.FilterSelectableCodeSignGroups(codeSignGroups,
				export.CreateExcludeProfileNameSelectableCodeSignGroupFilter(defaultProfile.Name))
			if len(filteredCodeSignGroups) > 0 {
				codeSignGroups = filteredCodeSignGroups

				g.logger.Debugf("\nGroups after removing default profile:")
				for _, group := range codeSignGroups {
					g.logger.Debugf(group.String())
				}
			}
		}
	}

	var iosCodeSignGroups []export.IosCodeSignGroup

	for _, selectable := range codeSignGroups {
		bundleIDProfileMap := map[string]profileutil.ProvisioningProfileInfoModel{}
		for bundleID, profiles := range selectable.BundleIDProfilesMap {
			if len(profiles) > 0 {
				bundleIDProfileMap[bundleID] = profiles[0]
			} else {
				g.logger.Warnf("No profile available to sign (%s) target!", bundleID)
			}
		}

		iosCodeSignGroups = append(iosCodeSignGroups, *export.NewIOSGroup(selectable.Certificate, bundleIDProfileMap))
	}

	g.logger.Debugf("\nFiltered groups:")
	for i, group := range iosCodeSignGroups {
		g.logger.Debugf("Group #%d:", i)
		for bundleID, profile := range group.BundleIDProfileMap() {
			g.logger.Debugf(" - %s: %s (%s)", bundleID, profile.Name, profile.UUID)
		}
	}

	if len(iosCodeSignGroups) < 1 {
		g.logger.Errorf("Failed to find Codesign Groups")
		return nil, nil
	}

	if len(iosCodeSignGroups) > 1 {
		g.logger.Warnf("Multiple code signing groups found! Using the first code signing group")
	}

	return &iosCodeSignGroups[0], nil
}

// ArchivableApplicationTarget locate archivable app target from a given project and scheme
func ArchivableApplicationTarget(xcodeProj *xcodeproj.XcodeProj, scheme *xcscheme.Scheme) (*xcodeproj.Target, error) {
	archiveEntry, ok := scheme.AppBuildActionEntry()
	if !ok {
		return nil, fmt.Errorf("archivable entry not found in project: %s for scheme: %s", xcodeProj.Path, scheme.Name)
	}

	mainTarget, ok := xcodeProj.Proj.Target(archiveEntry.BuildableReference.BlueprintIdentifier)
	if !ok {
		return nil, fmt.Errorf("target not found: %s", archiveEntry.BuildableReference.BlueprintIdentifier)
	}

	return &mainTarget, nil
}

func addDistributionBundleIdentifierFromXcode12(exportOpts exportoptions.ExportOptions, distributionBundleIdentifier string) exportoptions.ExportOptions {
	switch options := exportOpts.(type) {
	case exportoptions.AppStoreOptionsModel:
		// Export option plist with App store export method (Xcode 12.0.1) do not contain distribution bundle identifier.
		// Probably due to App store IPAs containing App Clips also, which are executable targets with a separate bundle ID.
		return exportOpts
	case exportoptions.NonAppStoreOptionsModel:
		options.DistributionBundleIdentifier = distributionBundleIdentifier
		return options
	}
	return nil
}

func disableManagedBuildNumberFromXcode13(exportOpts exportoptions.ExportOptions) exportoptions.ExportOptions {
	switch options := exportOpts.(type) {
	case exportoptions.AppStoreOptionsModel:
		options.ManageAppVersion = false // Only available for app-store exports

		return options
	}

	return exportOpts
}

func filterApplicationBundleTargets(targets []xcodeproj.Target, exportMethod exportoptions.Method) (filteredTargets []xcodeproj.Target) {
	fmt.Printf("Filtering %v application bundle targets", len(targets))

	for _, target := range targets {
		if !target.IsExecutableProduct() {
			continue
		}

		// App store exports contain App Clip too. App Clip provisioning profile has to be included in export options:
		// ..
		// <key>provisioningProfiles</key>
		// <dict>
		// 	<key>io.bundle.id</key>
		// 	<string>Development Application Profile</string>
		// 	<key>io.bundle.id.AppClipID</key>
		// 	<string>Development App Clip Profile</string>
		// </dict>
		// ..,
		if exportMethod != exportoptions.MethodAppStore && target.IsAppClipProduct() {
			continue
		}

		filteredTargets = append(filteredTargets, target)
	}

	fmt.Printf("Found %v application bundle targets", len(filteredTargets))

	return
}

// projectUsesCloudKit determines whether the project uses any CloudKit capability or not.
func projectUsesCloudKit(bundleIDEntitlementsMap map[string]plistutil.PlistData) bool {
	fmt.Printf("Checking if project uses CloudKit")

	for _, entitlements := range bundleIDEntitlementsMap {
		if entitlements == nil {
			continue
		}

		services, ok := entitlements.GetStringArray("com.apple.developer.icloud-services")
		if !ok {
			continue
		}

		if sliceutil.IsStringInSlice("CloudKit", services) || sliceutil.IsStringInSlice("CloudDocuments", services) {
			fmt.Printf("Project uses CloudKit")

			return true
		}
	}
	return false
}

// determineIcloudContainerEnvironment calculates the value of iCloudContainerEnvironment.
func determineIcloudContainerEnvironment(desiredIcloudContainerEnvironment string, bundleIDEntitlementsMap map[string]plistutil.PlistData, exportMethod exportoptions.Method, xcodeMajorVersion int) (string, error) {
	// iCloudContainerEnvironment: If the app is using CloudKit, this configures the "com.apple.developer.icloud-container-environment" entitlement.
	// Available options vary depending on the type of provisioning profile used, but may include: Development and Production.
	usesCloudKit := projectUsesCloudKit(bundleIDEntitlementsMap)
	if !usesCloudKit {
		return "", nil
	}

	// From Xcode 9 iCloudContainerEnvironment is required for every export method, before that version only for non app-store exports.
	if xcodeMajorVersion < 9 && exportMethod == exportoptions.MethodAppStore {
		return "", nil
	}

	if exportMethod == exportoptions.MethodAppStore {
		return "Production", nil
	}

	if desiredIcloudContainerEnvironment == "" {
		return "", fmt.Errorf("Your project uses CloudKit but \"iCloud container environment\" input not specified.\n"+
			"Export method is: %s (For app-store export method Production container environment is implied.)", exportMethod)
	}

	return desiredIcloudContainerEnvironment, nil
}

// generateBaseExportOptions creates a default exportOptions introduced in Xcode 7.
func generateBaseExportOptions(exportMethod exportoptions.Method, teamID string, uploadBitcode, compileBitcode bool, iCloudContainerEnvironment string) exportoptions.ExportOptions {
	if exportMethod == exportoptions.MethodAppStore {
		appStoreOptions := exportoptions.NewAppStoreOptions()
		appStoreOptions.TeamID = teamID
		appStoreOptions.UploadBitcode = uploadBitcode
		if iCloudContainerEnvironment != "" {
			appStoreOptions.ICloudContainerEnvironment = exportoptions.ICloudContainerEnvironment(iCloudContainerEnvironment)
		}
		return appStoreOptions
	}

	nonAppStoreOptions := exportoptions.NewNonAppStoreOptions(exportMethod)
	nonAppStoreOptions.CompileBitcode = compileBitcode

	if iCloudContainerEnvironment != "" {
		nonAppStoreOptions.ICloudContainerEnvironment = exportoptions.ICloudContainerEnvironment(iCloudContainerEnvironment)
	}

	return nonAppStoreOptions
}
