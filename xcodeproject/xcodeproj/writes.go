package xcodeproj

import (
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"

	"github.com/bitrise-io/go-plist"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
)

// pbxProjFileMode is the mode project.pbxproj is written with. It matches v1 and Xcode: other tools
// and users read this file, so it must not be written owner-only.
const pbxProjFileMode = 0644

// ForceCodeSignOptions are the manual code signing settings ForceCodeSign applies.
type ForceCodeSignOptions struct {
	// TargetName is the target to change.
	TargetName string
	// Configuration is the name of the build configuration to change, such as Release.
	Configuration string
	// DevelopmentTeam is the team ID.
	DevelopmentTeam string
	// CodesignIdentity is the signing identity, such as "Apple Distribution: Bitrise (ABCD1234)".
	CodesignIdentity string
	// ProvisioningProfileUUID is the UUID of the provisioning profile to sign with.
	ProvisioningProfileUUID string
}

// ForceCodeSign switches one target's configuration to manual code signing, in memory. Call Save to
// write the change to disk.
//
// In the configuration's build settings it sets CODE_SIGN_STYLE, DEVELOPMENT_TEAM,
// CODE_SIGN_IDENTITY and PROVISIONING_PROFILE, clears PROVISIONING_PROFILE_SPECIFIER, and applies
// the same value to any SDK-specific variant already present, such as
// CODE_SIGN_IDENTITY[sdk=iphoneos*]. In the project's TargetAttributes, if the target has an entry,
// it sets ProvisioningStyle and DevelopmentTeam and clears DevelopmentTeamName.
func (p *XcodeProj) ForceCodeSign(opts ForceCodeSignOptions) error {
	target, ok := p.TargetByName(opts.TargetName)
	if !ok {
		return fmt.Errorf("failed to find target with name: %s", opts.TargetName)
	}

	buildSettings, err := targetBuildSettings(target, opts.Configuration)
	if err != nil {
		return err
	}

	forced := map[string]string{
		"CODE_SIGN_STYLE":                "Manual",
		"DEVELOPMENT_TEAM":               opts.DevelopmentTeam,
		"CODE_SIGN_IDENTITY":             opts.CodesignIdentity,
		"PROVISIONING_PROFILE_SPECIFIER": "",
		"PROVISIONING_PROFILE":           opts.ProvisioningProfileUUID,
	}
	for key, value := range forced {
		writeBuildSettingForAllSDKs(buildSettings, key, value)
	}

	// Projects that do not use TargetAttributes, or have no entry for this target, are left alone.
	if attributes, ok := p.targetAttributes.Object(target.ID); ok {
		attributes["ProvisioningStyle"] = "Manual"
		attributes["DevelopmentTeam"] = opts.DevelopmentTeam
		attributes["DevelopmentTeamName"] = ""
	}

	return nil
}

// SetBuildSetting sets one build setting of one target configuration, in memory. Call Save to write
// the change to disk.
//
// As in v1, a configuration's build settings are the ones in the project tree, not a copy: Target
// and BuildConfiguration values fetched before the call see the new value too.
func (p *XcodeProj) SetBuildSetting(targetName, configurationName, key string, value any) error {
	target, ok := p.TargetByName(targetName)
	if !ok {
		return fmt.Errorf("failed to find target with name: %s", targetName)
	}

	buildSettings, err := targetBuildSettings(target, configurationName)
	if err != nil {
		return err
	}

	buildSettings[key] = value

	return nil
}

// Save writes the project to <Path>/project.pbxproj.
//
// Only the objects that changed are rewritten; every other byte of the file, including ordering and
// comments, stays as it was. That keeps the file compatible with tools such as Cordova and Xcode's
// agvtool. If the in-place rewrite is not possible, for instance because an object was added, the
// whole file is written out instead and a warning is logged.
func (p *XcodeProj) Save() error {
	pth := filepath.Join(p.Path, pbxProjFileName)

	content, err := p.perObjectModify()
	if err != nil {
		p.logger.Warnf("failed to modify project in-place: %v", err)

		content, err = plist.MarshalIndent(p.rawProj, p.format, "\t")
		if err != nil {
			return fmt.Errorf("failed to marshal .pbxproj: %w", err)
		}
	}

	if err := p.fileManager.Write(pth, string(content), pbxProjFileMode); err != nil {
		return fmt.Errorf("failed to write %s: %w", pth, err)
	}

	return nil
}

func targetBuildSettings(target Target, configurationName string) (serialized.Object, error) {
	for _, configuration := range target.BuildConfigurations {
		if configuration.Name != configurationName {
			continue
		}
		if configuration.buildSettings == nil {
			return nil, fmt.Errorf("build configuration %s of target %s has no buildSettings", configurationName, target.Name)
		}
		return configuration.buildSettings, nil
	}

	return nil, fmt.Errorf("failed to find build configuration %s of target %s", configurationName, target.Name)
}

// writeBuildSettingForAllSDKs sets key, and also every SDK-specific variant of it already present.
// Example: setting CODE_SIGN_IDENTITY also sets CODE_SIGN_IDENTITY[sdk=iphoneos*].
// See https://stackoverflow.com/a/5382708/5842489
func writeBuildSettingForAllSDKs(buildSettings serialized.Object, key, value string) {
	buildSettings[key] = value

	sdkVariant := regexp.MustCompile(fmt.Sprintf(`^%s\[sdk=.*\]$`, regexp.QuoteMeta(key)))
	for existingKey := range buildSettings {
		if sdkVariant.MatchString(existingKey) {
			buildSettings[existingKey] = value
		}
	}
}

type objectChange struct {
	start, end int
	content    []byte
}

// perObjectModify rebuilds the file by splicing re-serialised versions of changed objects into the
// original bytes. It compares the current raw tree with a fresh decode of the original contents to
// find what changed, and uses the byte offsets the annotating decoder records to find where.
func (p *XcodeProj) perObjectModify() ([]byte, error) {
	var annotated serialized.Object
	if _, err := plist.UnmarshalWithCustomAnnotation(p.originalContents, &annotated); err != nil {
		return nil, fmt.Errorf("failed to decode original project: %w", err)
	}

	original, _, err := decodePBXProj(p.originalContents)
	if err != nil {
		return nil, err
	}

	objectsModified, ok := p.rawProj.Object("objects")
	if !ok {
		return nil, fmt.Errorf("failed to parse project: no objects")
	}
	objectsOriginal, ok := original.Object("objects")
	if !ok {
		return nil, fmt.Errorf("failed to parse original project: no objects")
	}
	objectsAnnotated, ok := annotated.Object("objects")
	if !ok {
		return nil, fmt.Errorf("failed to parse annotated project: no objects")
	}

	var changes []objectChange
	for id := range objectsModified {
		objectModified, ok := objectsModified.Object(id)
		if !ok {
			return nil, fmt.Errorf("object %s is not a dictionary", id)
		}

		objectOriginal, ok := objectsOriginal.Object(id)
		if !ok {
			return nil, fmt.Errorf("new object added, not in original project: %s", id)
		}

		if reflect.DeepEqual(objectOriginal, objectModified) {
			continue
		}

		objectAnnotated, ok := objectsAnnotated.Object(id)
		if !ok {
			return nil, fmt.Errorf("new object added, not in original annotated project: %s", id)
		}

		position, ok := objectAnnotated.Object(customAnnotationKey)
		if !ok {
			return nil, fmt.Errorf("no raw object position available for %s", id)
		}
		start, ok := position.Int64(startKey)
		if !ok {
			return nil, fmt.Errorf("no raw object start position available for %s", id)
		}
		end, ok := position.Int64(endKey)
		if !ok {
			return nil, fmt.Errorf("no raw object end position available for %s", id)
		}

		content, err := plist.MarshalIndent(objectModified, p.format, "\t")
		if err != nil {
			return nil, fmt.Errorf("could not marshal object %s: %w", id, err)
		}

		changes = append(changes, objectChange{start: int(start), end: int(end), content: content})
	}

	if len(changes) == 0 {
		return p.originalContents, nil
	}

	sort.Slice(changes, func(i, j int) bool {
		if changes[i].start == changes[j].start {
			return changes[i].end < changes[j].end
		}
		return changes[i].start < changes[j].start
	})

	var result []byte
	previousEnd := 0
	for i, change := range changes {
		if i < len(changes)-1 && change.end >= changes[i+1].start {
			return nil, fmt.Errorf("overlapping changes: %d, %d", change.end, changes[i+1].start)
		}

		result = append(result, p.originalContents[previousEnd:change.start]...)
		result = append(result, change.content...)
		previousEnd = change.end
	}

	if previousEnd <= len(p.originalContents)-1 {
		result = append(result, p.originalContents[previousEnd:]...)
	}

	return result, nil
}
