package exportoptionsgenerator

import (
	"fmt"

	"github.com/bitrise-io/go-xcode/exportoptions"
	"github.com/bitrise-io/go-xcode/plistutil"
	"github.com/bitrise-io/go-xcode/xcodeproject/serialized"
	"github.com/bitrise-io/go-xcode/xcodeproject/xcodeproj"
	"github.com/bitrise-io/go-xcode/xcodeproject/xcscheme"
)

// TargetInfoProvider can determine a target's bundle id and codesign entitlements.
type TargetInfoProvider interface {
	applicationTargetsAndEntitlements(exportMethod exportoptions.Method) (string, map[string]plistutil.PlistData, error)
}

// XcodebuildTargetInfoProvider implements TargetInfoProvider.
type XcodebuildTargetInfoProvider struct {
	xcodeProj     *xcodeproj.XcodeProj
	scheme        *xcscheme.Scheme
	configuration string
}

func NewXcodebuildTargetInfoProvider(xcodeProj *xcodeproj.XcodeProj, scheme *xcscheme.Scheme, configuration string) TargetInfoProvider {
	return &XcodebuildTargetInfoProvider{
		xcodeProj:     xcodeProj,
		scheme:        scheme,
		configuration: configuration,
	}
}

// TargetBundleID ...
func (b XcodebuildTargetInfoProvider) TargetBundleID(target, configuration string) (string, error) {
	return b.xcodeProj.TargetBundleID(target, configuration)
}

// TargetCodeSignEntitlements ...
func (b XcodebuildTargetInfoProvider) TargetCodeSignEntitlements(target, configuration string) (serialized.Object, error) {
	return b.xcodeProj.TargetCodeSignEntitlements(target, configuration)
}

func (b XcodebuildTargetInfoProvider) applicationTargetsAndEntitlements(exportMethod exportoptions.Method) (string, map[string]plistutil.PlistData, error) {
	mainTarget, err := ArchivableApplicationTarget(b.xcodeProj, b.scheme)
	if err != nil {
		return "", nil, err
	}

	dependentTargets := filterApplicationBundleTargets(
		b.xcodeProj.DependentTargetsOfTarget(*mainTarget),
		exportMethod,
	)
	targets := append([]xcodeproj.Target{*mainTarget}, dependentTargets...)

	var mainTargetBundleID string
	entitlementsByBundleID := map[string]plistutil.PlistData{}
	for i, target := range targets {
		bundleID, err := b.TargetBundleID(target.Name, b.configuration)
		if err != nil {
			return "", nil, fmt.Errorf("failed to get target (%s) bundle id: %s", target.Name, err)
		}

		entitlements, err := b.TargetCodeSignEntitlements(target.Name, b.configuration)
		if err != nil && !serialized.IsKeyNotFoundError(err) {
			return "", nil, fmt.Errorf("failed to get target (%s) bundle id: %s", target.Name, err)
		}

		entitlementsByBundleID[bundleID] = plistutil.PlistData(entitlements)

		if i == 0 {
			mainTargetBundleID = bundleID
		}
	}

	return mainTargetBundleID, entitlementsByBundleID, nil
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
		if !exportMethod.IsAppStore() && target.IsAppClipProduct() {
			continue
		}

		filteredTargets = append(filteredTargets, target)
	}

	fmt.Printf("Found %v application bundle targets", len(filteredTargets))

	return
}
