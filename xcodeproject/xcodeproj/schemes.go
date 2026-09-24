package xcodeproj

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/text/unicode/norm"

	"github.com/bitrise-io/go-plist"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcscheme"
)

const (
	xcschemeExtension             = ".xcscheme"
	schemeManagementFileName      = "xcschememanagement.plist"
	autocreateSchemesSettingKey   = "IDEWorkspaceSharedSettings_AutocreateContextsIfNeeded"
	errNoSchemesAutocreateOffText = "no schemes found and the Xcode project's 'Autocreate schemes' option is disabled"
)

// Schemes returns the schemes Xcode shows when it opens the project on its own: the shared
// schemes, then the current user's schemes.
//
// If there are none, Xcode generates default schemes — one per native, non-test target — when
// the user's scheme-management file exists or when the project's "Autocreate schemes" option is
// on, which is the default. Generated schemes are built in memory and marked shared. If neither
// applies, Schemes returns an error.
func (p *XcodeProj) Schemes() ([]xcscheme.Scheme, error) {
	p.logger.TDebugf("Searching schemes in project: %s", p.Path)

	schemes, found, err := p.existingOrManagedSchemes()
	if err != nil || found {
		return schemes, err
	}

	// Read only when needed, as in v1: a project that has schemes never depends on this file.
	autocreate, err := p.isAutocreateSchemesEnabled()
	if err != nil {
		return nil, fmt.Errorf("failed to read the project autocreate scheme option: %w", err)
	}

	if autocreate {
		p.logger.TDebugf("Autocreating the default schemes")
		return p.defaultSchemes(), nil
	}

	return nil, errors.New(errNoSchemesAutocreateOffText)
}

// SchemesWithAutocreateOverride is Schemes for a project opened as part of a workspace: the
// "Autocreate schemes" decision comes from the caller, which applies the workspace's setting,
// instead of from this project's own settings.
//
// It differs from Schemes in one case: when there are no schemes and autocreate is off, it returns
// no schemes and no error. A project with no accessible schemes is normal inside a workspace — the
// Pods.xcodeproj that CocoaPods generates is one.
func (p *XcodeProj) SchemesWithAutocreateOverride(autocreateEnabled bool) ([]xcscheme.Scheme, error) {
	p.logger.TDebugf("Searching schemes in project: %s", p.Path)

	schemes, found, err := p.existingOrManagedSchemes()
	if err != nil || found {
		return schemes, err
	}

	if autocreateEnabled {
		p.logger.TDebugf("Autocreating the default schemes")
		return p.defaultSchemes(), nil
	}

	p.logger.TDebugf("No schemes found")

	return nil, nil
}

// Scheme returns the scheme with the given name and the path of the project that contains it.
// Names are compared after Unicode normalisation, so a name typed with composed or decomposed
// accents matches either way. An unknown name yields an xcscheme.NotFoundError.
func (p *XcodeProj) Scheme(name string) (xcscheme.Scheme, string, error) {
	schemes, err := p.Schemes()
	if err != nil {
		return xcscheme.Scheme{}, "", err
	}

	normalisedName := norm.NFC.String(name)
	for _, scheme := range schemes {
		if norm.NFC.String(scheme.Name) == normalisedName {
			return scheme, p.Path, nil
		}
	}

	return xcscheme.Scheme{}, "", xcscheme.NotFoundError{Scheme: name, Container: p.Name}
}

// existingOrManagedSchemes covers the steps Schemes and SchemesWithAutocreateOverride share. found
// is true when the answer is settled: the project has schemes on disk, or Xcode has managed schemes
// for this user before, in which case the default schemes are the answer. Otherwise the caller
// decides based on the autocreate option.
func (p *XcodeProj) existingOrManagedSchemes() (schemes []xcscheme.Scheme, found bool, err error) {
	schemes, err = p.visibleSchemes()
	if err != nil {
		return nil, false, err
	}

	if len(schemes) > 0 {
		p.logger.TDebugf("%d scheme(s) found", len(schemes))
		return schemes, true, nil
	}

	schemeManagementExists, err := p.isUserSchemeManagementFileExist()
	if err != nil {
		return nil, false, err
	}

	if schemeManagementExists {
		p.logger.TDebugf("Default scheme found")
		return p.defaultSchemes(), true, nil
	}

	return nil, false, nil
}

func (p *XcodeProj) defaultSchemes() []xcscheme.Scheme {
	schemes := p.generateSchemes()
	for i := range schemes {
		schemes[i].IsShared = true
	}
	return schemes
}

func (p *XcodeProj) visibleSchemes() ([]xcscheme.Scheme, error) {
	shared, err := p.readSchemes(p.sharedSchemesDir())
	if err != nil {
		return nil, err
	}
	for i := range shared {
		shared[i].IsShared = true
	}

	userDir, err := p.userSchemesDir()
	if err != nil {
		return nil, err
	}

	user, err := p.readSchemes(userDir)
	if err != nil {
		return nil, err
	}

	return append(shared, user...), nil
}

// sharedSchemesDir is <project>.xcodeproj/xcshareddata/xcschemes.
func (p *XcodeProj) sharedSchemesDir() string {
	return filepath.Join(p.Path, "xcshareddata", "xcschemes")
}

// userSchemesDir is <project>.xcodeproj/xcuserdata/<current user>.xcuserdatad/xcschemes.
func (p *XcodeProj) userSchemesDir() (string, error) {
	username, err := p.userProvider.CurrentUserName()
	if err != nil {
		return "", fmt.Errorf("failed to get the current user: %w", err)
	}

	return filepath.Join(p.Path, "xcuserdata", username+".xcuserdatad", "xcschemes"), nil
}

// readSchemes reads every .xcscheme file in dir. A missing directory means no schemes.
func (p *XcodeProj) readSchemes(dir string) ([]xcscheme.Scheme, error) {
	names, err := p.fileManager.ReadDirEntryNames(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var schemes []xcscheme.Scheme
	for _, name := range names {
		if filepath.Ext(name) != xcschemeExtension {
			continue
		}

		scheme, err := p.readScheme(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}

		schemes = append(schemes, scheme)
	}

	return schemes, nil
}

// readScheme reads a scheme through the injected FileManager, matching what xcscheme.Open does
// with the filesystem directly.
func (p *XcodeProj) readScheme(pth string) (xcscheme.Scheme, error) {
	file, err := p.fileManager.Open(pth)
	if err != nil {
		return xcscheme.Scheme{}, err
	}
	defer func() {
		_ = file.Close()
	}()

	scheme, err := xcscheme.Parse(file)
	if err != nil {
		return xcscheme.Scheme{}, fmt.Errorf("failed to unmarshal scheme file: %s: %w", pth, err)
	}

	scheme.Name = trimExtension(filepath.Base(pth))
	scheme.Path = pth

	return scheme, nil
}

// isUserSchemeManagementFileExist reports whether Xcode has already managed schemes for the current
// user, which it records in xcschememanagement.plist.
func (p *XcodeProj) isUserSchemeManagementFileExist() (bool, error) {
	userDir, err := p.userSchemesDir()
	if err != nil {
		return false, err
	}

	file, err := p.fileManager.Open(filepath.Join(userDir, schemeManagementFileName))
	if err != nil {
		return false, nil
	}
	_ = file.Close()

	return true, nil
}

// isAutocreateSchemesEnabled reads the "Autocreate schemes" option from the project's embedded
// workspace settings. It is on by default: a missing settings file or a missing key means on.
func (p *XcodeProj) isAutocreateSchemesEnabled() (bool, error) {
	pth := filepath.Join(p.Path, "project.xcworkspace", "xcshareddata", "WorkspaceSettings.xcsettings")

	file, err := p.fileManager.Open(pth)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return true, nil
		}
		return false, err
	}
	defer func() {
		_ = file.Close()
	}()

	content, err := io.ReadAll(file)
	if err != nil {
		return false, err
	}

	var settings serialized.Object
	if _, err := plist.Unmarshal(content, &settings); err != nil {
		return false, err
	}

	if !settings.Has(autocreateSchemesSettingKey) {
		return true, nil
	}

	autocreate, ok := settings.Bool(autocreateSchemesSettingKey)
	if !ok {
		return false, fmt.Errorf("%s is not a boolean", autocreateSchemesSettingKey)
	}

	return autocreate, nil
}

func trimExtension(name string) string {
	return name[:len(name)-len(filepath.Ext(name))]
}
