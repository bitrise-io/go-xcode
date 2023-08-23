package xcodeproj

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"golang.org/x/text/unicode/norm"

	"github.com/bitrise-io/go-plist"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-xcode/xcodeproject/serialized"
	"github.com/bitrise-io/go-xcode/xcodeproject/xcscheme"
)

// Schemes returns the schemes considered by Xcode, when opening the given project.
// The considered schemes are the project shared schemes and the project user schemes (for the current user).
// The default (shared scheme) is present in the user's xcschememanagement.plist,
// any schemes related change trigger generating all the schemes as xcscheme files.
// If no schemes are found, Xcode recreates the default schemes unless 'Autocreate schemes' option is disabled
// (in this case actions are disabled in Xcode, and 'No schemes' message appears).
func (p XcodeProj) Schemes() ([]xcscheme.Scheme, error) {
	log.TDebugf("Searching schemes in project: %s", p.Path)

	schemes, err := p.visibleSchemes()
	if err != nil {
		return nil, err
	}

	if len(schemes) == 0 {
		isUserSchememanagementFileExist, err := p.isUserSchememanagementFileExist()
		if err != nil {
			return nil, err
		}

		if isUserSchememanagementFileExist {
			log.TDebugf("Default scheme found")
			defaultSchemes := p.defaultSchemes()
			return defaultSchemes, nil
		}

		isAutocreateSchemesEnabled, err := p.isAutocreateSchemesEnabled()
		if err != nil {
			return nil, err
		}

		if isAutocreateSchemesEnabled {
			log.TDebugf("Autocreating the default scheme")
			defaultSchemes := p.defaultSchemes()
			return defaultSchemes, nil
		}

		return nil, fmt.Errorf("no schemes found and the Xcode project's 'Autocreate schemes' option is disabled")
	}

	log.TDebugf("%d scheme(s) found", len(schemes))
	return schemes, nil
}

// SchemesWithAutocreateEnabled returns the schemes considered by Xcode, when opening the given project as part of a workspace.
// SchemesWithAutocreateEnabled behaves similarly to XcodeProj.Schemes,
// the only difference is that the 'Autocreate schemes' option is coming from the workspace settings.
func (p XcodeProj) SchemesWithAutocreateEnabled(isAutocreateSchemesEnabled bool) ([]xcscheme.Scheme, error) {
	log.TDebugf("Searching schemes in project: %s", p.Path)

	schemes, err := p.visibleSchemes()
	if err != nil {
		return nil, err
	}

	if len(schemes) == 0 {
		isUserSchememanagementFileExist, err := p.isUserSchememanagementFileExist()
		if err != nil {
			return nil, err
		}

		if isUserSchememanagementFileExist {
			log.TDebugf("Default scheme found")
			defaultSchemes := p.defaultSchemes()
			return defaultSchemes, nil
		}

		if isAutocreateSchemesEnabled {
			log.TDebugf("Autocreating the default scheme")
			defaultSchemes := p.defaultSchemes()
			return defaultSchemes, nil
		}

		return nil, fmt.Errorf("no schemes found and the Xcode project's 'Autocreate schemes' option is disabled")
	}

	log.TDebugf("%d scheme(s) found", len(schemes))
	return schemes, nil
}

// Scheme returns the project's scheme by name and the project's absolute path.
func (p XcodeProj) Scheme(name string) (*xcscheme.Scheme, string, error) {
	schemes, err := p.Schemes()
	if err != nil {
		return nil, "", err
	}

	normName := norm.NFC.String(name)
	for _, scheme := range schemes {
		if norm.NFC.String(scheme.Name) == normName {
			return &scheme, p.Path, nil
		}
	}

	return nil, "", xcscheme.NotFoundError{Scheme: name, Container: p.Name}
}

func (p XcodeProj) visibleSchemes() ([]xcscheme.Scheme, error) {
	sharedSchemes, err := p.sharedSchemes()
	if err != nil {
		return nil, err
	}

	userSchemes, err := p.userSchemes()
	if err != nil {
		return nil, err
	}

	schemes := append(sharedSchemes, userSchemes...)
	return schemes, nil
}

func (p XcodeProj) sharedSchemes() ([]xcscheme.Scheme, error) {
	sharedSchemeFilePaths, err := p.sharedSchemeFilePaths()
	if err != nil {
		return nil, err
	}

	var sharedSchemes []xcscheme.Scheme
	for _, pth := range sharedSchemeFilePaths {
		scheme, err := xcscheme.Open(pth)
		if err != nil {
			return nil, err
		}

		scheme.IsShared = true
		sharedSchemes = append(sharedSchemes, scheme)
	}

	return sharedSchemes, nil
}

func (p XcodeProj) userSchemes() ([]xcscheme.Scheme, error) {
	userSchemeFilePaths, err := p.userSchemeFilePaths()
	if err != nil {
		return nil, err
	}

	var userSchemes []xcscheme.Scheme
	for _, pth := range userSchemeFilePaths {
		scheme, err := xcscheme.Open(pth)
		if err != nil {
			return nil, err
		}

		userSchemes = append(userSchemes, scheme)
	}

	return userSchemes, nil
}

func (p XcodeProj) sharedSchemeFilePaths() ([]string, error) {
	// <project_name>.xcodeproj/xcshareddata/xcschemes/<scheme_name>.xcscheme
	sharedSchemesDir := filepath.Join(p.Path, "xcshareddata", "xcschemes")
	return listSchemeFilePaths(sharedSchemesDir)
}

func (p XcodeProj) userSchemeFilePaths() ([]string, error) {
	// <project_name>.xcodeproj/xcuserdata/<current_user>.xcuserdatad/xcschemes/<scheme_name>.xcscheme
	userSchemesDir, err := p.userSchemesDir()
	if err != nil {
		return nil, err
	}
	return listSchemeFilePaths(userSchemesDir)
}

func (p XcodeProj) userSchemesDir() (string, error) {
	// <project_name>.xcodeproj/xcuserdata/<current_user>.xcuserdatad/xcschemes/
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	username := currentUser.Username

	return filepath.Join(p.Path, "xcuserdata", username+".xcuserdatad", "xcschemes"), nil
}

func (p XcodeProj) isUserSchememanagementFileExist() (bool, error) {
	// <project_name>.xcodeproj/xcuserdata/<current_user>.xcuserdatad/xcschemes/xcschememanagement.plist
	userSchemesDir, err := p.userSchemesDir()
	if err != nil {
		return false, err
	}
	schemeManagementPth := filepath.Join(userSchemesDir, "xcschememanagement.plist")
	_, err = os.Stat(schemeManagementPth)
	return err == nil, nil
}

func (p XcodeProj) isAutocreateSchemesEnabled() (bool, error) {
	// <project_name>.xcodeproj/project.xcworkspace/xcshareddata/WorkspaceSettings.xcsettings
	embeddedWorkspaceDir := filepath.Join(p.Path, "project.xcworkspace")
	shareddataDir := filepath.Join(embeddedWorkspaceDir, "xcshareddata")
	workspaceSettingsPth := filepath.Join(shareddataDir, "WorkspaceSettings.xcsettings")

	workspaceSettingsContent, err := os.ReadFile(workspaceSettingsPth)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// By default 'Autocreate Schemes' is enabled
			return true, nil
		}

		return false, err
	}

	var settings serialized.Object
	if _, err := plist.Unmarshal(workspaceSettingsContent, &settings); err != nil {
		return false, err
	}

	autoCreate, err := settings.Bool("IDEWorkspaceSharedSettings_AutocreateContextsIfNeeded")
	if err != nil {
		if serialized.IsKeyNotFoundError(err) {
			// By default 'Autocreate Schemes' is enabled
			return true, nil
		}
		return false, err
	}

	return autoCreate, nil
}

func (p XcodeProj) defaultSchemes() []xcscheme.Scheme {
	var schemes []xcscheme.Scheme
	for _, buildTarget := range p.Proj.Targets {
		if buildTarget.Type != NativeTargetType || buildTarget.IsTest() {
			continue
		}

		var testTargets []Target
		for _, testTarget := range p.Proj.Targets {
			if testTarget.IsTest() && testTarget.DependsOn(buildTarget.ID) {
				testTargets = append(testTargets, testTarget)
			}
		}

		scheme := newScheme(buildTarget, testTargets, filepath.Base(p.Path))
		scheme.IsShared = true
		schemes = append(schemes, scheme)
	}
	return schemes
}

func listSchemeFilePaths(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var schemeFilePaths []string
	for _, entry := range entries {
		baseName := entry.Name()
		if filepath.Ext(baseName) == ".xcscheme" {
			schemeFilePaths = append(schemeFilePaths, filepath.Join(dir, baseName))
		}
	}

	return schemeFilePaths, nil
}
