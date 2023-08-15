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

// Schemes returns the schemes considered by Xcode.
// The considered schemes are the shared schemes and the current user's schemes.
// The default (shared scheme) is present in the user's xcschememanagement.plist,
// any schemes related change trigger generating all the schemes as xcscheme files.
// If no schemes are found, Xcode recreates the default schemes unless ‘Autocreate schemes' option is disabled
// (in this case actions are disabled in Xcode, and 'No schemes’ message appears).
func (p XcodeProj) Schemes() ([]xcscheme.Scheme, error) {
	log.TDebugf("Locating scheme for project path: %s", p.Path)

	sharedSchemes, err := p.sharedSchemes()
	if err != nil {
		return nil, err
	}

	userSchemes, err := p.userSchemes()
	if err != nil {
		return nil, err
	}

	schemes := append(sharedSchemes, userSchemes...)

	if len(schemes) == 0 {
		isUserSchememanagementFileExist, err := p.isUserSchememanagementFileExist()
		if err != nil {
			return nil, err
		}

		if isUserSchememanagementFileExist {
			defaultSchemes := p.ReCreateSchemes()
			return defaultSchemes, nil
		}

		isAutocreateSchemesEnabled, err := p.isAutocreateSchemesEnabled()
		if err != nil {
			return nil, err
		}

		if isAutocreateSchemesEnabled {
			defaultSchemes := p.ReCreateSchemes()
			return defaultSchemes, nil
		}

		return nil, fmt.Errorf("no schemes found and the Xcode project's 'Autocreate schemes' option is disabled")
	}

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

func (p XcodeProj) isAutocreateSchemesEnabled() (bool, error) {
	// <project_name>.xcodeproj/project.xcworkspace/xcshareddata/WorkspaceSettings.xcsettings
	embeddedWorkspaceDir := filepath.Join(p.Path, "project.xcworkspace")
	shareddataDir := filepath.Join(embeddedWorkspaceDir, "xcshareddata")
	workspaceSettingsPth := filepath.Join(shareddataDir, "WorkspaceSettings.xcsettings")

	workspaceSettingsContent, err := os.ReadFile(workspaceSettingsPth)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// By default Autocreate Schemes is enabled
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
		return false, err
	}

	return autoCreate, nil
}

func (p XcodeProj) sharedSchemeFilePaths() ([]string, error) {
	// <project_name>.xcodeproj/xcshareddata/xcschemes/<scheme_name>.xcscheme
	sharedSchemesDir := filepath.Join(p.Path, "xcshareddata", "xcschemes")
	return listSchemeFilePaths(sharedSchemesDir)
}

func (p XcodeProj) userSchemesDir() (string, error) {
	// <project_name>.xcodeproj/xcuserdata/<current_user>.xcuserdatad/xcschemes/<scheme_name>.xcscheme
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	username := currentUser.Username

	return filepath.Join(p.Path, "xcuserdata", username+".xcuserdatad", "xcschemes"), nil
}

func (p XcodeProj) userSchemeFilePaths() ([]string, error) {
	userSchemesDir, err := p.userSchemesDir()
	if err != nil {
		return nil, err
	}
	return listSchemeFilePaths(userSchemesDir)
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
