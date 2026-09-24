package xcodeproj

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bitrise-io/go-plist"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
)

// Build setting keys this package resolves by name.
const (
	bundleIDBuildSettingKey     = "PRODUCT_BUNDLE_IDENTIFIER"
	entitlementsBuildSettingKey = "CODE_SIGN_ENTITLEMENTS"
	infoPlistBuildSettingKey    = "INFOPLIST_FILE"
)

var (
	// bracedReferenceRegexp matches $(KEY) / ${KEY}, optionally with a :modifier suffix.
	bracedReferenceRegexp = regexp.MustCompile(`[$][{(][^$]*?[)}]`)
	// bareReferenceRegexp matches an undelimited $KEY reference.
	bareReferenceRegexp = regexp.MustCompile(`[$][^$]*`)

	referencePunctuationReplacer = strings.NewReplacer("$", "", "(", "", ")", "", "{", "", "}", "")
)

// ErrEntitlementsNotFound reports that a target declares no CODE_SIGN_ENTITLEMENTS file. Most
// targets do not, so this is a normal outcome rather than a failure — distinguish it from a real
// error (an entitlements file that is named but unreadable) with errors.Is.
var ErrEntitlementsNotFound = errors.New("target has no code sign entitlements file")

// TargetBuildSettings returns the target's effective build settings for a configuration: the
// values Xcode would build with, after variable expansion and after any .xcconfig files are
// applied.
//
// Settings are looked up by target (`xcodebuild -project … -target …`), not by scheme, so this
// works for targets that have no scheme of their own, such as app extensions. It runs
// `xcodebuild -showBuildSettings`, so it costs seconds.
//
// extraArgs are appended to the command. xcodebuild ignores destination-style arguments such as
// -destination in target mode, so they cannot be used to pick a platform here. Settings that depend
// on the platform (SDKROOT, PLATFORM_NAME) are therefore unreliable for multiplatform targets.
func (p *XcodeProj) TargetBuildSettings(target, configuration string, extraArgs ...string) (serialized.Object, error) {
	if p.buildSettings == nil {
		return nil, errors.New("no BuildSettingsProvider was injected")
	}
	return p.buildSettings.TargetBuildSettings(p.Path, target, configuration, extraArgs...)
}

// TargetBundleID returns the target's bundle identifier for a configuration, fully resolved.
//
// It reads PRODUCT_BUNDLE_IDENTIFIER from the effective build settings and falls back to the
// Info.plist's CFBundleIdentifier when that setting is absent. Either value may contain build
// setting references such as $(PRODUCT_NAME:rfc1034identifier), which are expanded against the
// same build settings.
func (p *XcodeProj) TargetBundleID(target, configuration string) (string, error) {
	buildSettings, err := p.TargetBuildSettings(target, configuration)
	if err != nil {
		return "", err
	}

	if bundleID, ok := buildSettings.String(bundleIDBuildSettingKey); ok && bundleID != "" {
		return resolveBundleID(bundleID, buildSettings)
	}

	infoPlist, _, err := p.readTargetInfoPlist(target, configuration, buildSettings)
	if err != nil {
		return "", fmt.Errorf("no %s build setting, and reading the Info.plist instead failed: %w", bundleIDBuildSettingKey, err)
	}

	bundleID, ok := infoPlist.String("CFBundleIdentifier")
	if !ok || bundleID == "" {
		return "", fmt.Errorf("no %s build setting nor CFBundleIdentifier Info.plist entry found for target %s", bundleIDBuildSettingKey, target)
	}

	return resolveBundleID(bundleID, buildSettings)
}

// TargetCodeSignEntitlements returns the contents of the target's entitlements file.
//
// If the target declares no CODE_SIGN_ENTITLEMENTS setting, the error satisfies
// errors.Is(err, ErrEntitlementsNotFound). Any other error means the file was named but could not
// be read or parsed, which is a real failure.
func (p *XcodeProj) TargetCodeSignEntitlements(target, configuration string) (serialized.Object, error) {
	buildSettings, err := p.TargetBuildSettings(target, configuration)
	if err != nil {
		return nil, err
	}

	pth, ok := p.buildSettingPath(buildSettings, entitlementsBuildSettingKey)
	if !ok {
		return nil, fmt.Errorf("target %s: %w", target, ErrEntitlementsNotFound)
	}

	entitlements, _, err := p.readPlist(pth)
	if err != nil {
		return nil, fmt.Errorf("failed to read entitlements of target %s: %w", target, err)
	}

	return entitlements, nil
}

// TargetInfoplistPath returns the absolute path of the target's Info.plist.
//
// The path comes from the effective build settings, so it is already resolved: variable references
// are expanded and an .xcconfig file that defines INFOPLIST_FILE outside the project file is
// accounted for. Reading INFOPLIST_FILE from the project file instead would miss both.
func (p *XcodeProj) TargetInfoplistPath(target, configuration string) (string, error) {
	buildSettings, err := p.TargetBuildSettings(target, configuration)
	if err != nil {
		return "", err
	}

	pth, ok := p.buildSettingPath(buildSettings, infoPlistBuildSettingKey)
	if !ok {
		return "", fmt.Errorf("no %s build setting found for target %s", infoPlistBuildSettingKey, target)
	}

	return pth, nil
}

func (p *XcodeProj) readTargetInfoPlist(target, configuration string, buildSettings serialized.Object) (serialized.Object, int, error) {
	pth, ok := p.buildSettingPath(buildSettings, infoPlistBuildSettingKey)
	if !ok {
		return nil, plist.InvalidFormat, fmt.Errorf("no %s build setting found for target %s", infoPlistBuildSettingKey, target)
	}
	return p.readPlist(pth)
}

// buildSettingPath reads a path-valued build setting and makes it absolute. Relative values are
// resolved against the directory holding the .xcodeproj, which is what Xcode does.
func (p *XcodeProj) buildSettingPath(buildSettings serialized.Object, key string) (string, bool) {
	pth, ok := buildSettings.String(key)
	if !ok || pth == "" {
		return "", false
	}

	if isRelativePath(pth) {
		pth = filepath.Join(filepath.Dir(p.Path), pth)
	}

	return pth, true
}

// isRelativePath reports whether a build setting's path value should be resolved against the
// project directory. It mirrors go-utils v1 pathutil.IsRelativePath, which has no go-utils v2
// equivalent, so that this package treats paths exactly as the v1 implementation did.
//
// Note the "$" case: a value that still contains an unexpanded build setting reference is left
// alone rather than being joined onto the project directory, which would only produce a longer
// broken path. Effective build settings come from xcodebuild and are normally already expanded, so
// this is a guard rather than a common route.
func isRelativePath(pth string) bool {
	switch {
	case strings.HasPrefix(pth, "./"):
		return true
	case strings.HasPrefix(pth, "/"):
		return false
	case strings.HasPrefix(pth, "$"):
		return false
	default:
		return true
	}
}

func (p *XcodeProj) readPlist(pth string) (serialized.Object, int, error) {
	file, err := p.fileManager.Open(pth)
	if err != nil {
		return nil, plist.InvalidFormat, err
	}
	defer func() {
		_ = file.Close()
	}()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, plist.InvalidFormat, err
	}

	var object serialized.Object
	format, err := plist.Unmarshal(content, &object)
	if err != nil {
		return nil, plist.InvalidFormat, fmt.Errorf("failed to unmarshal %s: %w", pth, err)
	}

	return object, format, nil
}

// resolveBundleID expands build setting references in a bundle identifier until none remain.
//
// A bundle ID in the project file can reference build settings, for example
// `Bitrise.Test.$(PRODUCT_NAME:rfc1034identifier).Suffix`, and the referenced value may itself
// contain a reference. Expansion repeats until the result is literal, and stops with an error if
// the references form a cycle.
func resolveBundleID(bundleID string, buildSettings serialized.Object) (string, error) {
	seen := map[string]bool{}
	resolved := bundleID

	for strings.Contains(resolved, "$") {
		expanded, err := expandBuildSetting(resolved, buildSettings)
		if err != nil {
			return "", err
		}

		if seen[expanded] {
			return "", fmt.Errorf("bundle id reference cycle found while resolving %s", bundleID)
		}
		seen[expanded] = true

		resolved = expanded
	}

	return resolved, nil
}

func expandBuildSetting(value string, buildSettings serialized.Object) (string, error) {
	if bracedReferenceRegexp.MatchString(value) {
		// $(PRODUCT_NAME) / ${PRODUCT_NAME} / $(PRODUCT_NAME:rfc1034identifier)
		return expandBracedReference(value, buildSettings)
	}
	// $PRODUCT_NAME
	return expandBareReference(value, buildSettings)
}

// expandBracedReference replaces the first $(...) or ${...} reference.
// Example: `prefix.$(ENV_KEY:rfc1034identifier).suffix` => `prefix.value.suffix`
func expandBracedReference(value string, buildSettings serialized.Object) (string, error) {
	reference := bracedReferenceRegexp.FindString(value)

	// Strip the punctuation, then drop any :modifier suffix such as :rfc1034identifier.
	key := referencePunctuationReplacer.Replace(reference)
	key = strings.Split(key, ":")[0]

	settingValue, ok := buildSettings.String(key)
	if !ok {
		return "", fmt.Errorf("failed to find env in build settings: %s", key)
	}

	return strings.ReplaceAll(value, reference, settingValue), nil
}

// expandBareReference replaces a `$KEY` reference that has no braces.
//
// Where the reference runs into surrounding text, as in `$PRODUCT_NAME.suffix`, the key is not
// delimited, so the longest match is tried first and characters are dropped from the end until a
// build setting matches.
func expandBareReference(value string, buildSettings serialized.Object) (string, error) {
	if !bareReferenceRegexp.MatchString(value) {
		return "", fmt.Errorf("failed to match a build setting reference in %s", value)
	}

	reference := bareReferenceRegexp.FindString(value)

	var settingValue string
	for len(reference) > 1 {
		var ok bool
		settingValue, ok = buildSettings.String(strings.Replace(reference, "$", "", 1))
		if ok {
			break
		}
		reference = reference[:len(reference)-1]
	}

	return strings.ReplaceAll(value, reference, settingValue), nil
}
