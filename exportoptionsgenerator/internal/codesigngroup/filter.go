package codesigngroup

import (
	"github.com/bitrise-io/go-xcode/exportoptions"
	"github.com/bitrise-io/go-xcode/plistutil"
	"github.com/bitrise-io/go-xcode/profileutil"
)

// GroupMapFunc ...
type GroupMapFunc func(group Selectable) (Selectable, bool)

func MapGroups(groups []Selectable, mapFunc GroupMapFunc) []Selectable {
	if mapFunc == nil {
		return groups
	}

	var mappedGroups []Selectable
	for _, group := range groups {
		groupWithFilteredProfiles, ok := mapFunc(group)
		if ok {
			mappedGroups = append(mappedGroups, groupWithFilteredProfiles)
		}
	}

	return mappedGroups
}

// CreateEntitlementsFilter ...
func CreateEntitlementsFilter(bundleIDEntitlementsMap map[string]plistutil.PlistData) GroupMapFunc {
	return func(group Selectable) (Selectable, bool) {
		filteredBundleIDProfilesMap := map[string][]profileutil.ProvisioningProfileInfoModel{}

		for bundleID, profiles := range group.BundleIDProfilesMap {
			var filteredProfiles []profileutil.ProvisioningProfileInfoModel

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
			group.BundleIDProfilesMap = filteredBundleIDProfilesMap
			return group, true
		}

		return group, false
	}
}

// CreateExportMethodFilter ...
func CreateExportMethodFilter(exportMethod exportoptions.Method) GroupMapFunc {
	return func(group Selectable) (Selectable, bool) {
		return filterGroupProfiles(group, func(profile profileutil.ProvisioningProfileInfoModel) bool {
			return profile.ExportType == exportMethod
		})
	}
}

// CreateTeamIDFilter ...
func CreateTeamIDFilter(teamID string) GroupMapFunc {
	return func(group Selectable) (Selectable, bool) {
		if group.Certificate.TeamID == teamID {
			return group, true
		}
		return group, false
	}
}

// CreateNonXcodeManagedFilter ...
func CreateNonXcodeManagedFilter() GroupMapFunc {
	return func(group Selectable) (Selectable, bool) {
		return filterGroupProfiles(group, func(profile profileutil.ProvisioningProfileInfoModel) bool {
			return !profile.IsXcodeManaged()
		})
	}
}

// CreateXcodeManagedFilter ...
func CreateXcodeManagedFilter() GroupMapFunc {
	return func(group Selectable) (Selectable, bool) {
		return filterGroupProfiles(group, func(profile profileutil.ProvisioningProfileInfoModel) bool {
			return profile.IsXcodeManaged()
		})
	}
}

// CreateExcludeProfileNameFilter ...
func CreateExcludeProfileNameFilter(name string) GroupMapFunc {
	return func(group Selectable) (Selectable, bool) {
		return filterGroupProfiles(group, func(profile profileutil.ProvisioningProfileInfoModel) bool {
			return profile.Name != name
		})
	}
}

// filter - returns a slice containing only the elements
// that satisfy the predicate function filterFunc.
func filter[T any](slice []T, filterFunc func(T) bool) []T {
	if filterFunc == nil {
		return slice
	}

	filtered := make([]T, 0, len(slice))
	for _, item := range slice {
		if filterFunc(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func filterGroupProfiles(group Selectable, filterFunc func(profile profileutil.ProvisioningProfileInfoModel) bool) (Selectable, bool) {
	filteredBundleIDProfilesMap := map[string][]profileutil.ProvisioningProfileInfoModel{}
	for bundleID, profiles := range group.BundleIDProfilesMap {
		filteredBundleIDProfilesMap[bundleID] = filter(profiles, filterFunc)
		if len(filteredBundleIDProfilesMap[bundleID]) == 0 {
			return group, false
		}
	}

	group.BundleIDProfilesMap = filteredBundleIDProfilesMap
	return group, true
}
