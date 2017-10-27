package export

import (
	"fmt"
	"sort"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-tools/go-xcode/certificateutil"
	"github.com/bitrise-tools/go-xcode/exportoptions"
	"github.com/bitrise-tools/go-xcode/plistutil"
	"github.com/bitrise-tools/go-xcode/profileutil"
	"github.com/ryanuber/go-glob"
)

func isCertificateInstalled(installedCertificates []certificateutil.CertificateInfoModel, certificate certificateutil.CertificateInfoModel) bool {
	installedMap := map[string]bool{}
	for _, certificate := range installedCertificates {
		installedMap[certificate.Serial] = true
	}
	return installedMap[certificate.Serial]
}

// CertificateProfilesGroup ...
type CertificateProfilesGroup struct {
	Certificate certificateutil.CertificateInfoModel
	Profiles    []profileutil.ProvisioningProfileInfoModel
}

// PrintCertificateProfilesGroup ...
func PrintCertificateProfilesGroup(group CertificateProfilesGroup) {
	log.Printf(group.Certificate.CommonName)
	for _, profile := range group.Profiles {
		log.Printf("- %s:", profile.Name)
	}
}

func createCertificateProfilesGroups(certificates []certificateutil.CertificateInfoModel, profiles []profileutil.ProvisioningProfileInfoModel) []CertificateProfilesGroup {
	serialProfilesMap := map[string][]profileutil.ProvisioningProfileInfoModel{}
	serialCertificateMap := map[string]certificateutil.CertificateInfoModel{}
	for _, profile := range profiles {
		for _, certificate := range profile.DeveloperCertificates {
			if !isCertificateInstalled(certificates, certificate) {
				continue
			}

			certificateProfiles := serialProfilesMap[certificate.Serial]
			if certificateProfiles == nil {
				certificateProfiles = []profileutil.ProvisioningProfileInfoModel{}
			}
			certificateProfiles = append(certificateProfiles, profile)
			serialProfilesMap[certificate.Serial] = certificateProfiles
			serialCertificateMap[certificate.Serial] = certificate
		}
	}

	groups := []CertificateProfilesGroup{}
	for serial, profiles := range serialProfilesMap {
		certificate := serialCertificateMap[serial]
		group := CertificateProfilesGroup{
			Certificate: certificate,
			Profiles:    profiles,
		}
		PrintCertificateProfilesGroup(group)
		groups = append(groups, group)
	}

	return groups
}

// SelectableCodeSignGroup ..
type SelectableCodeSignGroup struct {
	Certificate         certificateutil.CertificateInfoModel
	BundleIDProfilesMap map[string][]profileutil.ProvisioningProfileInfoModel
}

// PrintSelectableCodeSignGroup ...
func PrintSelectableCodeSignGroup(group SelectableCodeSignGroup) {
	log.Printf(group.Certificate.CommonName)
	for bundleID, profiles := range group.BundleIDProfilesMap {
		log.Printf("%s:", bundleID)
		for _, profile := range profiles {
			log.Printf("- %s", profile.Name)
		}
	}
}

func createSelectableCodeSignGroups(certificateProfilesGroups []CertificateProfilesGroup, bundleIDCapabilitiesMap map[string]plistutil.PlistData, exportMethod exportoptions.Method) []SelectableCodeSignGroup {
	groups := []SelectableCodeSignGroup{}

	for _, certificateProfilesGroup := range certificateProfilesGroups {
		certificate := certificateProfilesGroup.Certificate
		profiles := certificateProfilesGroup.Profiles

		log.Printf("Checking certificate - profiles group: %s", certificate.CommonName)

		bundleIDProfilesMap := map[string][]profileutil.ProvisioningProfileInfoModel{}
		for bundleID, capabilities := range bundleIDCapabilitiesMap {

			matchingProfiles := []profileutil.ProvisioningProfileInfoModel{}
			for _, profile := range profiles {
				if profile.ExportType != exportMethod {
					log.Printf("Profile (%s) export type (%s) does not match: %s", profile.Name, profile.ExportType, exportMethod)
					continue
				}

				if !glob.Glob(profile.BundleID, bundleID) {
					log.Printf("Profile (%s) bundle id (%s) does not match: %s", profile.Name, profile.BundleID, bundleID)
					continue
				}

				if missingCapabilities := profileutil.MatchTargetAndProfileEntitlements(capabilities, profile.Entitlements); len(missingCapabilities) > 0 {
					log.Printf("Profile (%s) does not have capabilities: %v", profile.Name, missingCapabilities)
					continue
				}

				log.Printf("Profile (%s) matches", profile.Name)

				matchingProfiles = append(matchingProfiles, profile)
			}

			if len(matchingProfiles) > 0 {
				sort.Sort(ByBundleIDLength(matchingProfiles))
				bundleIDProfilesMap[bundleID] = matchingProfiles
			}
		}

		if len(bundleIDProfilesMap) == len(bundleIDCapabilitiesMap) {
			group := SelectableCodeSignGroup{
				Certificate:         certificate,
				BundleIDProfilesMap: bundleIDProfilesMap,
			}
			groups = append(groups, group)

			log.Printf("Valid code sign group:")
			PrintSelectableCodeSignGroup(group)

		} else {
			log.Printf("Removing certificate - profiles group: %s", certificate.CommonName)
		}
	}

	return groups
}

// ResolveSelectableCodeSignGroups ...
func ResolveSelectableCodeSignGroups(certificates []certificateutil.CertificateInfoModel, profiles []profileutil.ProvisioningProfileInfoModel, bundleIDCapabilities map[string]plistutil.PlistData, exportMethod exportoptions.Method) []SelectableCodeSignGroup {
	log.Printf("Creating certificate profiles groups...")
	certificateProfilesGroups := createCertificateProfilesGroups(certificates, profiles)

	log.Printf("Creating selectable code sign groups...")
	return createSelectableCodeSignGroups(certificateProfilesGroups, bundleIDCapabilities, exportMethod)
}

// CodeSignGroup ...
type CodeSignGroup struct {
	Certificate        certificateutil.CertificateInfoModel
	BundleIDProfileMap map[string]profileutil.ProvisioningProfileInfoModel
}

// PrintCodeSignGroup ...
func PrintCodeSignGroup(group CodeSignGroup) {
	log.Printf(group.Certificate.CommonName)
	for bundleID, profile := range group.BundleIDProfileMap {
		log.Printf("%s:", bundleID)
		log.Printf("- %s", profile.Name)
	}
}

func createCodeSignGroups(selectableGroups []SelectableCodeSignGroup) []CodeSignGroup {
	alreadyUsedProfileUUIDMap := map[string]bool{}

	singleWildcardGroups := []CodeSignGroup{}
	xcodeManagedGroups := []CodeSignGroup{}
	notXcodeManagedGroups := []CodeSignGroup{}
	remainingGroups := []CodeSignGroup{}

	for _, selectableGroup := range selectableGroups {
		certificate := selectableGroup.Certificate
		bundleIDProfilesMap := selectableGroup.BundleIDProfilesMap

		bundleIDs := []string{}
		profiles := []profileutil.ProvisioningProfileInfoModel{}
		for bundleID, matchingProfiles := range bundleIDProfilesMap {
			bundleIDs = append(bundleIDs, bundleID)
			profiles = append(profiles, matchingProfiles...)
		}

		log.Printf("Checking certificate - profiles group: %s", certificate.CommonName)

		//
		// create groups with single wildcard profiles
		{
			log.Printf("Checking for group with single wildcard profile")
			for _, profile := range profiles {
				if alreadyUsedProfileUUIDMap[profile.UUID] {
					continue
				}

				matchesForAllBundleID := true
				for _, bundleID := range bundleIDs {
					if !glob.Glob(profile.BundleID, bundleID) {
						matchesForAllBundleID = false
						break
					}
				}
				if matchesForAllBundleID {
					bundleIDProfileMap := map[string]profileutil.ProvisioningProfileInfoModel{}
					for _, bundleID := range bundleIDs {
						bundleIDProfileMap[bundleID] = profile
					}

					group := CodeSignGroup{
						Certificate:        certificate,
						BundleIDProfileMap: bundleIDProfileMap,
					}
					singleWildcardGroups = append(singleWildcardGroups, group)

					alreadyUsedProfileUUIDMap[profile.UUID] = true

					log.Printf("Group with single wildcard profile found:")
					PrintCodeSignGroup(group)
				}
			}
		}

		//
		// create groups with xcode managed profiles
		{
			log.Printf("Checking for group with xcode managed profiles")

			// collect xcode managed profiles
			xcodeManagedProfiles := []profileutil.ProvisioningProfileInfoModel{}
			for _, profile := range profiles {
				if !alreadyUsedProfileUUIDMap[profile.UUID] && profile.IsXcodeManaged() {
					xcodeManagedProfiles = append(xcodeManagedProfiles, profile)
				}
			}

			// map profiles to bundle ids
			bundleIDMannagedProfilesMap := map[string][]profileutil.ProvisioningProfileInfoModel{}
			for _, bundleID := range bundleIDs {
				for _, profile := range xcodeManagedProfiles {
					if !glob.Glob(profile.BundleID, bundleID) {
						continue
					}

					matchingProfiles := bundleIDMannagedProfilesMap[bundleID]
					if matchingProfiles == nil {
						matchingProfiles = []profileutil.ProvisioningProfileInfoModel{}
					}
					matchingProfiles = append(matchingProfiles, profile)
					bundleIDMannagedProfilesMap[bundleID] = matchingProfiles
				}
			}

			if len(bundleIDMannagedProfilesMap) == len(bundleIDs) {
				// if only one profile can sign a bundle id, remove it from other bundle id - profiles map
				alreadyUsedManagedProfileMap := map[string]bool{}
				for _, profiles := range bundleIDMannagedProfilesMap {
					if len(profiles) == 1 {
						profile := profiles[0]
						alreadyUsedManagedProfileMap[profile.UUID] = true
					}
				}

				bundleIDMannagedProfileMap := map[string]profileutil.ProvisioningProfileInfoModel{}
				for bundleID, profiles := range bundleIDMannagedProfilesMap {
					if len(profiles) == 1 {
						bundleIDMannagedProfileMap[bundleID] = profiles[0]
					} else {
						remainingProfiles := []profileutil.ProvisioningProfileInfoModel{}
						for _, profile := range profiles {
							if !alreadyUsedManagedProfileMap[profile.UUID] {
								remainingProfiles = append(remainingProfiles, profile)
							}
						}
						if len(remainingProfiles) == 1 {
							bundleIDMannagedProfileMap[bundleID] = remainingProfiles[0]
						}
					}
				}

				// create code sign group
				if len(bundleIDMannagedProfileMap) == len(bundleIDs) {
					group := CodeSignGroup{
						Certificate:        certificate,
						BundleIDProfileMap: bundleIDMannagedProfileMap,
					}
					xcodeManagedGroups = append(xcodeManagedGroups, group)

					log.Printf("Group with xcode managed profiles found:")
					PrintCodeSignGroup(group)
				}
			}
		}

		//
		// create groups with NOT xcode managed profiles
		{
			log.Printf("Checking for group with NOT xcode managed profiles")

			// collect xcode managed profiles
			notXcodeManagedProfiles := []profileutil.ProvisioningProfileInfoModel{}
			for _, profile := range profiles {
				if !alreadyUsedProfileUUIDMap[profile.UUID] && !profile.IsXcodeManaged() {
					notXcodeManagedProfiles = append(notXcodeManagedProfiles, profile)
				}
			}

			// map profiles to bundle ids
			bundleIDNotMannagedProfilesMap := map[string][]profileutil.ProvisioningProfileInfoModel{}
			for _, bundleID := range bundleIDs {
				for _, profile := range notXcodeManagedProfiles {
					if !glob.Glob(profile.BundleID, bundleID) {
						continue
					}

					matchingProfiles := bundleIDNotMannagedProfilesMap[bundleID]
					if matchingProfiles == nil {
						matchingProfiles = []profileutil.ProvisioningProfileInfoModel{}
					}
					matchingProfiles = append(matchingProfiles, profile)
					bundleIDNotMannagedProfilesMap[bundleID] = matchingProfiles
				}
			}

			if len(bundleIDNotMannagedProfilesMap) == len(bundleIDs) {
				// if only one profile can sign a bundle id, remove it from other bundle id - profiles map
				alreadyUsedManagedProfileMap := map[string]bool{}
				for _, profiles := range bundleIDNotMannagedProfilesMap {
					if len(profiles) == 1 {
						profile := profiles[0]
						alreadyUsedManagedProfileMap[profile.UUID] = true
					}
				}

				bundleIDMannagedProfileMap := map[string]profileutil.ProvisioningProfileInfoModel{}
				for bundleID, profiles := range bundleIDNotMannagedProfilesMap {
					if len(profiles) == 1 {
						bundleIDMannagedProfileMap[bundleID] = profiles[0]
					} else {
						remainingProfiles := []profileutil.ProvisioningProfileInfoModel{}
						for _, profile := range profiles {
							if !alreadyUsedManagedProfileMap[profile.UUID] {
								remainingProfiles = append(remainingProfiles, profile)
							}
						}
						if len(remainingProfiles) == 1 {
							bundleIDMannagedProfileMap[bundleID] = remainingProfiles[0]
						}
					}
				}

				// create code sign group
				if len(bundleIDMannagedProfileMap) == len(bundleIDs) {
					codeSignGroup := CodeSignGroup{
						Certificate:        certificate,
						BundleIDProfileMap: bundleIDMannagedProfileMap,
					}
					notXcodeManagedGroups = append(notXcodeManagedGroups, codeSignGroup)

					log.Printf("Group with NOT xcode managed profiles found:")
					PrintCodeSignGroup(codeSignGroup)
				}
			}
		}

		//
		// if there are remaining profiles we create a not exact group by using the first matching profile for every bundle id
		{
			if len(alreadyUsedProfileUUIDMap) != len(profiles) {
				log.Printf("There are remaining profile create group by using the first matching profile for every bundle id")

				bundleIDProfileMap := map[string]profileutil.ProvisioningProfileInfoModel{}
				for _, bundleID := range bundleIDs {
					for _, profile := range profiles {
						if alreadyUsedProfileUUIDMap[profile.UUID] {
							continue
						}

						if !glob.Glob(profile.BundleID, bundleID) {
							continue
						}

						bundleIDProfileMap[bundleID] = profile
						break
					}
				}

				if len(bundleIDProfileMap) == len(bundleIDs) {
					group := CodeSignGroup{
						Certificate:        certificate,
						BundleIDProfileMap: bundleIDProfileMap,
					}
					remainingGroups = append(remainingGroups, group)

					log.Printf("Group with first matching profiles:")
					PrintCodeSignGroup(group)
				}
			}
		}

		fmt.Println()
	}

	codeSignGroups := []CodeSignGroup{}
	codeSignGroups = append(codeSignGroups, notXcodeManagedGroups...)
	codeSignGroups = append(codeSignGroups, xcodeManagedGroups...)
	codeSignGroups = append(codeSignGroups, singleWildcardGroups...)
	codeSignGroups = append(codeSignGroups, remainingGroups...)

	return codeSignGroups
}

// ResolveCodeSignGroups ...
func ResolveCodeSignGroups(certificates []certificateutil.CertificateInfoModel, profiles []profileutil.ProvisioningProfileInfoModel, bundleIDCapabilities map[string]plistutil.PlistData, exportMethod exportoptions.Method) []CodeSignGroup {
	selectableCodeSignGroups := ResolveSelectableCodeSignGroups(certificates, profiles, bundleIDCapabilities, exportMethod)

	log.Printf("Creating code sign groups...")
	return createCodeSignGroups(selectableCodeSignGroups)
}
