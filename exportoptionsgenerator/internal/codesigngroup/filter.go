package codesigngroup

import (
	"github.com/bitrise-io/go-xcode/exportoptions"
	"github.com/bitrise-io/go-xcode/profileutil"
	"github.com/bitrise-io/go-xcode/v2/plistutil"
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
		return filterGroupProfiles(group, func(bundleID string, profile profileutil.ProvisioningProfileInfoModel) bool {
			v1BundleIDEntitlementsMap := convertToV1PlistData(bundleIDEntitlementsMap)
			missingEntitlements := profileutil.MatchTargetAndProfileEntitlements(v1BundleIDEntitlementsMap[bundleID], profile.Entitlements, profile.Type)
			return len(missingEntitlements) == 0
		})
	}
}

// CreateExportMethodFilter ...
func CreateExportMethodFilter(exportMethod exportoptions.Method) GroupMapFunc {
	return func(group Selectable) (Selectable, bool) {
		return filterGroupProfiles(group, func(bundleID string, profile profileutil.ProvisioningProfileInfoModel) bool {
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
		return filterGroupProfiles(group, func(bundleID string, profile profileutil.ProvisioningProfileInfoModel) bool {
			return !profile.IsXcodeManaged()
		})
	}
}

// CreateXcodeManagedFilter ...
func CreateXcodeManagedFilter() GroupMapFunc {
	return func(group Selectable) (Selectable, bool) {
		return filterGroupProfiles(group, func(bundleID string, profile profileutil.ProvisioningProfileInfoModel) bool {
			return profile.IsXcodeManaged()
		})
	}
}

// CreateExcludeProfileNameFilter ...
func CreateExcludeProfileNameFilter(name string) GroupMapFunc {
	return func(group Selectable) (Selectable, bool) {
		return filterGroupProfiles(group, func(bundleID string, profile profileutil.ProvisioningProfileInfoModel) bool {
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

func filterGroupProfiles(group Selectable, filterFunc func(bundleID string, profile profileutil.ProvisioningProfileInfoModel) bool) (Selectable, bool) {
	filteredBundleIDProfilesMap := map[string][]profileutil.ProvisioningProfileInfoModel{}
	for bundleID, profiles := range group.BundleIDProfilesMap {
		filteredBundleIDProfilesMap[bundleID] = filter(profiles, func(profile profileutil.ProvisioningProfileInfoModel) bool {
			return filterFunc(bundleID, profile)
		})
		if len(filteredBundleIDProfilesMap[bundleID]) == 0 {
			return group, false
		}
	}

	group.BundleIDProfilesMap = filteredBundleIDProfilesMap
	return group, true
}
