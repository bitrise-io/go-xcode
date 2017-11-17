package export

import (
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-tools/go-xcode/exportoptions"
	"github.com/bitrise-tools/go-xcode/plistutil"
	"github.com/bitrise-tools/go-xcode/profileutil"
)

// SelectableCodeSignGroupFilter ...
type SelectableCodeSignGroupFilter = func(group SelectableCodeSignGroup) *SelectableCodeSignGroup

// FilterSelectableCodeSignGroups ...
func FilterSelectableCodeSignGroups(groups []SelectableCodeSignGroup, filterFuncs ...SelectableCodeSignGroupFilter) []SelectableCodeSignGroup {
	log.Debugf("\n")
	log.Debugf("Filtering Codesign Groups...")

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

	if len(filteredGroups) == 0 {
		log.Debugf("Every Codesign Groups removed by the filters")
	}

	for _, group := range filteredGroups {
		log.Debugf(printableSelectableCodeSignGroup(group))
	}

	return filteredGroups
}

// CreateEntitlementsSelectableCodeSignGroupFilter ..
func CreateEntitlementsSelectableCodeSignGroupFilter(bundleIDEntitlementsMap map[string]plistutil.PlistData) SelectableCodeSignGroupFilter {
	return func(group SelectableCodeSignGroup) *SelectableCodeSignGroup {
		log.Debugf("Entitlements filter - removes profile if has missing capabilities")

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
}

// CreateExportMethodSelectableCodeSignGroupFilter ...
func CreateExportMethodSelectableCodeSignGroupFilter(exportMethod exportoptions.Method) SelectableCodeSignGroupFilter {
	return func(group SelectableCodeSignGroup) *SelectableCodeSignGroup {
		log.Debugf("Export method filter - removes profile if distribution type is not: %s", exportMethod)

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
}

// CreatTeamSelectableCodeSignGroupFilter ...
func CreatTeamSelectableCodeSignGroupFilter(teamID string) SelectableCodeSignGroupFilter {
	return func(group SelectableCodeSignGroup) *SelectableCodeSignGroup {
		log.Debugf("Development Team filter - restrict group if team is not: %s", teamID)

		if group.Certificate.TeamID == teamID {
			return &group
		}
		return nil
	}
}

// CreateNotXcodeManagedSelectableCodeSignGroupFilter ...
func CreateNotXcodeManagedSelectableCodeSignGroupFilter() SelectableCodeSignGroupFilter {
	return func(group SelectableCodeSignGroup) *SelectableCodeSignGroup {
		log.Debugf("Xcode managed filter - removes profile if xcode managed")

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
}

// CreateExcludeProfileNameSelectableCodeSignGroupFilter ...
func CreateExcludeProfileNameSelectableCodeSignGroupFilter(name string) SelectableCodeSignGroupFilter {
	return func(group SelectableCodeSignGroup) *SelectableCodeSignGroup {
		log.Debugf("Profile name filter - removes profile with name: %s", name)

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
			return &SelectableCodeSignGroup{
				Certificate:         group.Certificate,
				BundleIDProfilesMap: filteredBundleIDProfilesMap,
			}
		}
		return nil
	}
}
