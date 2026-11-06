package xcodeproj

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"

	"github.com/bitrise-io/go-plist"
	"github.com/bitrise-io/go-xcode/v2/internal/filewriter"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
)

// newProjectFileMode is the mode of a newly created project file; an existing one keeps its mode.
const newProjectFileMode = 0644

// ForceCodeSignOptions are the manual code signing settings ForceCodeSign applies.
type ForceCodeSignOptions struct {
	TargetName              string
	Configuration           string
	DevelopmentTeam         string
	CodesignIdentity        string
	ProvisioningProfileUUID string
}

// ForceCodeSign implements XcodeProj. It sets CODE_SIGN_STYLE, DEVELOPMENT_TEAM, CODE_SIGN_IDENTITY
// and PROVISIONING_PROFILE, including their existing [sdk=...] variants, clears
// PROVISIONING_PROFILE_SPECIFIER, and records manual signing where the format has a place for it
// outside the build settings.
func (p *baseProj) ForceCodeSign(opts ForceCodeSignOptions) error {
	target, configuration, err := p.targetConfiguration(opts.TargetName, opts.Configuration)
	if err != nil {
		return err
	}

	for key, value := range forcedCodeSignBuildSettings(opts) {
		for _, sdkKey := range keysForAllSDKs(configuration.buildSettings, key) {
			if err := p.setBuildSetting(target, configuration, sdkKey, value); err != nil {
				return err
			}
		}
	}

	p.file.setManualSigning(target, opts.DevelopmentTeam)

	return nil
}

// SetBuildSetting implements XcodeProj.
func (p *baseProj) SetBuildSetting(targetName, configurationName, key string, value any) error {
	target, configuration, err := p.targetConfiguration(targetName, configurationName)
	if err != nil {
		return err
	}

	return p.setBuildSetting(target, configuration, key, value)
}

// setBuildSetting writes to the project file first and to memory only if the file accepted the
// value, so a rejected edit leaves the two in agreement.
func (p *baseProj) setBuildSetting(target Target, configuration BuildConfiguration, key string, value any) error {
	if err := p.file.setBuildSetting(target, configuration, key, value); err != nil {
		return err
	}
	configuration.buildSettings[key] = value

	return nil
}

// Save implements XcodeProj.
func (p *baseProj) Save() error {
	content, err := p.file.content()
	if err != nil {
		return err
	}

	if err := filewriter.WriteKeepingMode(p.fileManager, p.filePath, content, newProjectFileMode); err != nil {
		return fmt.Errorf("failed to write %s: %w", p.filePath, err)
	}

	return nil
}

func (p *baseProj) targetConfiguration(targetName, configurationName string) (Target, BuildConfiguration, error) {
	target, ok := p.TargetByName(targetName)
	if !ok {
		return Target{}, BuildConfiguration{}, fmt.Errorf("failed to find target with name: %s", targetName)
	}

	for _, configuration := range target.BuildConfigurations {
		if configuration.Name == configurationName {
			return target, configuration, nil
		}
	}

	return Target{}, BuildConfiguration{}, fmt.Errorf("failed to find build configuration %s of target %s", configurationName, target.Name)
}

func forcedCodeSignBuildSettings(opts ForceCodeSignOptions) map[string]string {
	return map[string]string{
		"CODE_SIGN_STYLE":                "Manual",
		"DEVELOPMENT_TEAM":               opts.DevelopmentTeam,
		"CODE_SIGN_IDENTITY":             opts.CodesignIdentity,
		"PROVISIONING_PROFILE_SPECIFIER": "",
		"PROVISIONING_PROFILE":           opts.ProvisioningProfileUUID,
	}
}

// key is returned even when buildSettings does not have it yet.
func keysForAllSDKs(buildSettings serialized.Object, key string) []string {
	keys := []string{key}

	sdkVariant := regexp.MustCompile(fmt.Sprintf(`^%s\[sdk=.*\]$`, regexp.QuoteMeta(key)))
	for existingKey := range buildSettings {
		if sdkVariant.MatchString(existingKey) {
			keys = append(keys, existingKey)
		}
	}

	return keys
}

type objectChange struct {
	start, end int
	content    []byte
}

// perObjectModify splices re-serialised changed objects into the original bytes.
// The original and annotated trees are decoded from originalContents here rather than kept from
// parse time, so an opened project holds no extra copies of the tree; Save is called rarely.
func (p *pbxprojFile) perObjectModify() ([]byte, error) {
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
