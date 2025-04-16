package exportoptionsgenerator

import (
	"fmt"

	"github.com/bitrise-io/go-xcode/plistutil"
	"github.com/bitrise-io/go-xcode/xcodeproject/serialized"
	"github.com/bitrise-io/go-xcode/xcodeproject/xcodeproj"
	"github.com/bitrise-io/go-xcode/xcodeproject/xcscheme"
)

type ArchiveInfo struct {
	MainBundleID           string
	AppClipBundleID        string
	EntitlementsByBundleID map[string]plistutil.PlistData
}

// InfoProvider can determine the exportable bundle ID(s) and codesign entitlements.
type InfoProvider interface {
	Read() (ArchiveInfo, error)
}

// XcodebuildInfoProvider implements TargetInfoProvider.
type XcodebuildInfoProvider struct {
	xcodeProj     *xcodeproj.XcodeProj
	scheme        *xcscheme.Scheme
	configuration string
}

func NewXcodebuildTargetInfoProvider(xcodeProj *xcodeproj.XcodeProj, scheme *xcscheme.Scheme, configuration string) InfoProvider {
	return &XcodebuildInfoProvider{
		xcodeProj:     xcodeProj,
		scheme:        scheme,
		configuration: configuration,
	}
}

// Read returns the main target's bundle ID and the entitlements of all dependent targets.
func (b XcodebuildInfoProvider) Read() (ArchiveInfo, error) {
	mainTarget, err := ArchivableApplicationTarget(b.xcodeProj, b.scheme)
	if err != nil {
		return ArchiveInfo{}, err
	}

	dependentTargets := filterApplicationBundleTargets(b.xcodeProj.DependentTargetsOfTarget(*mainTarget))
	targets := append([]xcodeproj.Target{*mainTarget}, dependentTargets...)

	mainTargetBundleID := ""
	appClipBundleID := ""
	entitlementsByBundleID := map[string]plistutil.PlistData{}
	for i, target := range targets {
		bundleID, err := b.xcodeProj.TargetBundleID(target.Name, b.configuration)
		if err != nil {
			return ArchiveInfo{}, fmt.Errorf("failed to get target (%s) bundle id: %s", target.Name, err)
		}

		entitlements, err := b.xcodeProj.TargetCodeSignEntitlements(target.Name, b.configuration)
		if err != nil && !serialized.IsKeyNotFoundError(err) {
			return ArchiveInfo{}, fmt.Errorf("failed to get target (%s) bundle id: %s", target.Name, err)
		}

		entitlementsByBundleID[bundleID] = plistutil.PlistData(entitlements)

		if target.IsAppClipProduct() {
			appClipBundleID = bundleID
		}
		if i == 0 {
			mainTargetBundleID = bundleID
		}
	}

	return ArchiveInfo{
		MainBundleID:           mainTargetBundleID,
		AppClipBundleID:        appClipBundleID,
		EntitlementsByBundleID: entitlementsByBundleID,
	}, nil
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

func filterApplicationBundleTargets(targets []xcodeproj.Target) (filteredTargets []xcodeproj.Target) {
	for _, target := range targets {
		if !target.IsExecutableProduct() {
			continue
		}

		filteredTargets = append(filteredTargets, target)
	}

	return
}
