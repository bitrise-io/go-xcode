package export

import (
	"github.com/bitrise-tools/go-xcode/exportoptions"
	"github.com/bitrise-tools/go-xcode/plistutil"
	"github.com/bitrise-tools/go-xcode/profileutil"
)

// SelectableCodeSignGroupFilter ...
type SelectableCodeSignGroupFilter = func(group SelectableCodeSignGroup) *SelectableCodeSignGroup

// FilterSelectableCodeSignGroups ...
func FilterSelectableCodeSignGroups(groups []SelectableCodeSignGroup, filterFuncs ...SelectableCodeSignGroupFilter) []SelectableCodeSignGroup {
	filteredGroups := []SelectableCodeSignGroup{}

	for _, group := range groups {
		allowed := true
		filteredGroup := group

		for _, filterFunc := range filterFuncs {
			if newGroup := filterFunc(filteredGroup); newGroup != nil {
				filteredGroup = *newGroup
			} else {
				allowed = false
				break
			}
		}

		if allowed {
			filteredGroups = append(filteredGroups, filteredGroup)
		}
	}

	return filteredGroups
}

// CreateEntitlementsSelectableCodeSignGroupFilter ..
func CreateEntitlementsSelectableCodeSignGroupFilter(group SelectableCodeSignGroup, bundleIDEntitlementsMap map[string]plistutil.PlistData) *SelectableCodeSignGroup {
	filteredBundleIDProfilesMap := map[string][]profileutil.ProvisioningProfileInfoModel{}

	for bundleID, profiles := range group.BundleIDProfilesMap {
		filteredProfiles := []profileutil.ProvisioningProfileInfoModel{}

		for _, profile := range profiles {
			missingEntitlements := profileutil.MatchTargetAndProfileEntitlements(bundleIDEntitlementsMap[bundleID], profile.Entitlements, profile.Type)
			if len(missingEntitlements) == 0 {
				filteredProfiles = append(filteredProfiles, profile)
			}
		}

		if len(filteredProfiles) == 0 {
			break
		}

		filteredBundleIDProfilesMap[bundleID] = filteredProfiles
	}

	if len(filteredBundleIDProfilesMap) == len(group.BundleIDProfilesMap) {
		return &SelectableCodeSignGroup{
			Certificate:         group.Certificate,
			BundleIDProfilesMap: filteredBundleIDProfilesMap,
		}
	}
	return nil
}

// CreateExportMethodSelectableCodeSignGroupFilter ...
func CreateExportMethodSelectableCodeSignGroupFilter(group SelectableCodeSignGroup, exportMethod exportoptions.Method) *SelectableCodeSignGroup {
	filteredBundleIDProfilesMap := map[string][]profileutil.ProvisioningProfileInfoModel{}

	for bundleID, profiles := range group.BundleIDProfilesMap {
		filteredProfiles := []profileutil.ProvisioningProfileInfoModel{}

		for _, profile := range profiles {
			if profile.ExportType == exportMethod {
				filteredProfiles = append(filteredProfiles, profile)
			}
		}

		if len(filteredProfiles) == 0 {
			break
		}

		filteredBundleIDProfilesMap[bundleID] = filteredProfiles
	}

	if len(filteredBundleIDProfilesMap) == len(group.BundleIDProfilesMap) {
		return &SelectableCodeSignGroup{
			Certificate:         group.Certificate,
			BundleIDProfilesMap: filteredBundleIDProfilesMap,
		}
	}
	return nil
}

// CreatTeamSelectableCodeSignGroupFilter ...
func CreatTeamSelectableCodeSignGroupFilter(group SelectableCodeSignGroup, teamID string) *SelectableCodeSignGroup {
	if group.Certificate.TeamID == teamID {
		return &group
	}
	return nil
}

// CreateNotXcodeManagedSelectableCodeSignGroupFilter ...
func CreateNotXcodeManagedSelectableCodeSignGroupFilter(group SelectableCodeSignGroup) *SelectableCodeSignGroup {
	filteredBundleIDProfilesMap := map[string][]profileutil.ProvisioningProfileInfoModel{}

	for bundleID, profiles := range group.BundleIDProfilesMap {
		filteredProfiles := []profileutil.ProvisioningProfileInfoModel{}

		for _, profile := range profiles {
			if !profile.IsXcodeManaged() {
				filteredProfiles = append(filteredProfiles, profile)
			}
		}

		if len(filteredProfiles) == 0 {
			break
		}

		filteredBundleIDProfilesMap[bundleID] = filteredProfiles
	}

	if len(filteredBundleIDProfilesMap) == len(group.BundleIDProfilesMap) {
		return &SelectableCodeSignGroup{
			Certificate:         group.Certificate,
			BundleIDProfilesMap: filteredBundleIDProfilesMap,
		}
	}
	return nil
}
