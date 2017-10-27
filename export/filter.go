package export

import (
	"fmt"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-tools/go-xcode/profileutil"
)

// FilterSelectableCodeSignGroupsForTeam ...
func FilterSelectableCodeSignGroupsForTeam(codeSignGroups []SelectableCodeSignGroup, teamID string) []SelectableCodeSignGroup {
	filteredGroups := []SelectableCodeSignGroup{}
	for _, group := range codeSignGroups {
		if group.Certificate.TeamID == teamID {
			filteredGroups = append(filteredGroups, group)
		} else {
			log.Warnf("removing CodeSignGroup: %s", group.Certificate.CommonName)
			fmt.Println()
		}
	}
	return filteredGroups
}

// FilterSelectableCodeSignGroupsForNotXcodeManagedProfiles ...
func FilterSelectableCodeSignGroupsForNotXcodeManagedProfiles(codeSignGroups []SelectableCodeSignGroup) []SelectableCodeSignGroup {
	filteredGroups := []SelectableCodeSignGroup{}
	for _, group := range codeSignGroups {

		bundleIDNotManagedProfilesMap := map[string][]profileutil.ProvisioningProfileInfoModel{}
		for bundleID, profiles := range group.BundleIDProfilesMap {
			notManagedProfiles := []profileutil.ProvisioningProfileInfoModel{}
			for _, profile := range profiles {
				if !profile.IsXcodeManaged() {
					notManagedProfiles = append(notManagedProfiles, profile)
				}
			}
			if len(notManagedProfiles) > 0 {
				bundleIDNotManagedProfilesMap[bundleID] = profiles
			}
		}

		if len(bundleIDNotManagedProfilesMap) == len(group.BundleIDProfilesMap) {
			filteredGroups = append(filteredGroups, group)
		} else {
			log.Warnf("removing CodeSignGroup: %s", group.Certificate.CommonName)
		}
	}
	return filteredGroups
}

// FilterCodeSignGroupsForTeam ...
func FilterCodeSignGroupsForTeam(codeSignGroups []CodeSignGroup, teamID string) []CodeSignGroup {
	filteredGroups := []CodeSignGroup{}
	for _, group := range codeSignGroups {
		if group.Certificate.TeamID == teamID {
			filteredGroups = append(filteredGroups, group)
		} else {
			log.Warnf("removing CodeSignGroup: %s", group.Certificate.CommonName)
			fmt.Println()
		}
	}
	return filteredGroups
}

// FilterCodeSignGroupsForNotXcodeManagedProfiles ...
func FilterCodeSignGroupsForNotXcodeManagedProfiles(codeSignGroups []CodeSignGroup) []CodeSignGroup {
	filteredGroups := []CodeSignGroup{}
	for _, group := range codeSignGroups {
		xcodeManagedGroup := false
		for _, profile := range group.BundleIDProfileMap {
			if profile.IsXcodeManaged() {
				xcodeManagedGroup = true
				break
			}
		}
		if !xcodeManagedGroup {
			filteredGroups = append(filteredGroups, group)
		} else {
			log.Warnf("removing CodeSignGroup: %s", group.Certificate.CommonName)
		}
	}
	return filteredGroups
}
