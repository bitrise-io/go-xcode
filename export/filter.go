package export

import (
	"github.com/bitrise-io/go-xcode/v2/exportoptions"
	"github.com/bitrise-io/go-xcode/v2/plistutil"
	"github.com/bitrise-io/go-xcode/v2/profileutil"
)

// SelectableCodeSignGroupFilter ...
type SelectableCodeSignGroupFilter func(group *SelectableCodeSignGroup) bool

// FilterSelectableCodeSignGroups ...
func FilterSelectableCodeSignGroups(groups []SelectableCodeSignGroup, filterFuncs ...SelectableCodeSignGroupFilter) []SelectableCodeSignGroup {
	filteredGroups := []SelectableCodeSignGroup{}

	for _, group := range groups {
		allowed := true

		for _, filterFunc := range filterFuncs {
			if !filterFunc(&group) {
				allowed = false
				break
			}
		}

		if allowed {
			filteredGroups = append(filteredGroups, group)
		}
	}

	return filteredGroups
}

// CreateEntitlementsSelectableCodeSignGroupFilter ...
func CreateEntitlementsSelectableCodeSignGroupFilter(bundleIDEntitlementsMap map[string]plistutil.PlistData) SelectableCodeSignGroupFilter {
	return func(group *SelectableCodeSignGroup) bool {
		filteredBundleIDProfilesMap := map[string][]profileutil.ProvisioningProfileInfoModel{}

		for bundleID, profiles := range group.BundleIDProfilesMap {
			filteredProfiles := []profileutil.ProvisioningProfileInfoModel{}

			for _, profile := range profiles {
				missingEntitlements := MatchTargetAndProfileEntitlements(bundleIDEntitlementsMap[bundleID], profile.Entitlements, profile.Type)
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
			group.BundleIDProfilesMap = filteredBundleIDProfilesMap
			return true
		}

		return false
	}
}

// CreateExportMethodSelectableCodeSignGroupFilter ...
func CreateExportMethodSelectableCodeSignGroupFilter(exportMethod exportoptions.Method) SelectableCodeSignGroupFilter {
	return func(group *SelectableCodeSignGroup) bool {
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
			group.BundleIDProfilesMap = filteredBundleIDProfilesMap
			return true
		}

		return false
	}
}

// CreateTeamSelectableCodeSignGroupFilter ...
func CreateTeamSelectableCodeSignGroupFilter(teamID string) SelectableCodeSignGroupFilter {
	return func(group *SelectableCodeSignGroup) bool {
		return group.Certificate.TeamID == teamID
	}
}

// CreateNotXcodeManagedSelectableCodeSignGroupFilter ...
func CreateNotXcodeManagedSelectableCodeSignGroupFilter() SelectableCodeSignGroupFilter {
	return func(group *SelectableCodeSignGroup) bool {
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
			group.BundleIDProfilesMap = filteredBundleIDProfilesMap
			return true
		}

		return false
	}
}

// CreateXcodeManagedSelectableCodeSignGroupFilter ...
func CreateXcodeManagedSelectableCodeSignGroupFilter() SelectableCodeSignGroupFilter {
	return func(group *SelectableCodeSignGroup) bool {
		filteredBundleIDProfilesMap := map[string][]profileutil.ProvisioningProfileInfoModel{}

		for bundleID, profiles := range group.BundleIDProfilesMap {
			filteredProfiles := []profileutil.ProvisioningProfileInfoModel{}

			for _, profile := range profiles {
				if profile.IsXcodeManaged() {
					filteredProfiles = append(filteredProfiles, profile)
				}
			}

			if len(filteredProfiles) == 0 {
				break
			}

			filteredBundleIDProfilesMap[bundleID] = filteredProfiles
		}

		if len(filteredBundleIDProfilesMap) == len(group.BundleIDProfilesMap) {
			group.BundleIDProfilesMap = filteredBundleIDProfilesMap
			return true
		}

		return false
	}
}

// CreateExcludeProfileNameSelectableCodeSignGroupFilter ...
func CreateExcludeProfileNameSelectableCodeSignGroupFilter(name string) SelectableCodeSignGroupFilter {
	return func(group *SelectableCodeSignGroup) bool {
		filteredBundleIDProfilesMap := map[string][]profileutil.ProvisioningProfileInfoModel{}

		for bundleID, profiles := range group.BundleIDProfilesMap {
			filteredProfiles := []profileutil.ProvisioningProfileInfoModel{}

			for _, profile := range profiles {
				if profile.Name != name {
					filteredProfiles = append(filteredProfiles, profile)
				}
			}

			if len(filteredProfiles) == 0 {
				break
			}

			filteredBundleIDProfilesMap[bundleID] = filteredProfiles
		}

		if len(filteredBundleIDProfilesMap) == len(group.BundleIDProfilesMap) {
			group.BundleIDProfilesMap = filteredBundleIDProfilesMap
			return true
		}

		return false
	}
}

// MatchTargetAndProfileEntitlements ...
func MatchTargetAndProfileEntitlements(targetEntitlements plistutil.PlistData, profileEntitlements plistutil.PlistData, profileType profileutil.ProfileType) []string {
	missingEntitlements := []string{}

	for key := range targetEntitlements {
		_, known := profileutil.KnownProfileCapabilitiesMap[profileType][key]
		if !known {
			continue
		}
		_, found := profileEntitlements[key]
		if !found {
			missingEntitlements = append(missingEntitlements, key)
		}
	}

	return missingEntitlements
}
