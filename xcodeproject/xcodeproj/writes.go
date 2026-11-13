package xcodeproj

import (
	"fmt"
	"regexp"

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

// ForceCodeSign switches a target configuration to manual code signing. Call Save to persist
// it.
//
// It sets CODE_SIGN_STYLE, DEVELOPMENT_TEAM, CODE_SIGN_IDENTITY
// and PROVISIONING_PROFILE, including their existing [sdk=...] variants, clears
// PROVISIONING_PROFILE_SPECIFIER, and records manual signing where the format has a place for it
// outside the build settings.
func (p *Project) ForceCodeSign(opts ForceCodeSignOptions) error {
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

// SetBuildSetting sets a build setting of a target configuration, such as
// CURRENT_PROJECT_VERSION or PRODUCT_BUNDLE_IDENTIFIER, in the project file. Several settings
// can be set before one Save persists them. Targets fetched earlier see the change.
func (p *Project) SetBuildSetting(targetName, configurationName, key string, value any) error {
	target, configuration, err := p.targetConfiguration(targetName, configurationName)
	if err != nil {
		return err
	}

	return p.setBuildSetting(target, configuration, key, value)
}

// setBuildSetting writes to the project file first and to memory only if the file accepted the
// value, so a rejected edit leaves the two in agreement.
func (p *Project) setBuildSetting(target Target, configuration BuildConfiguration, key string, value any) error {
	if err := p.file.setBuildSetting(target, configuration, key, value); err != nil {
		return err
	}
	configuration.buildSettings[key] = value

	return nil
}

// Save writes the project file, changing only what the edits changed.
func (p *Project) Save() error {
	content, err := p.file.content()
	if err != nil {
		return err
	}

	if err := filewriter.WriteKeepingMode(p.fileManager, p.filePath, content, newProjectFileMode); err != nil {
		return fmt.Errorf("failed to write %s: %w", p.filePath, err)
	}

	return nil
}

func (p *Project) targetConfiguration(targetName, configurationName string) (Target, BuildConfiguration, error) {
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
