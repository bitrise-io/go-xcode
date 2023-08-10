package xcodeproj

// In a newly generated Xcode project a scheme gets created for the main app target.
// The default scheme is shared, but there isn't a scheme file in the xcshareddata dir,
// instead the scheme is listed in the user data dir in the xcschememanagement.plist file
// and the scheme file name contains a '_^#shared#^' suffix.
// If this scheme is unshared, the suffix gets removed and a scheme file gets created in the user data dir.
// If the scheme gets shared again, the scheme file gets moved to the shared data dir.
//
// Xcode considers the shared schemes and user schemes of the given user, others user's schemes are not considered.
// If a project doesn't have any shared schemes and gets opened on a new machine and 'Autocreate schemes' option is enabled,
// a default scheme is created by Xcode in a same way as for a new project.
// If 'Autocreate schemes' option is disabled Xcode can't work with the project until a scheme is created.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitrise-io/go-plist"
	"github.com/bitrise-io/go-xcode/xcodeproject/serialized"
)

// TODO: is it ok to save the newly created schemes to the file system? e.g.: this will result in git changes.
func (p XcodeProj) autoCreateSchemesIfNeeded() error {
	isAutocreateSchemesNeeded, err := p.isAutocreateSchemesNeeded()
	if err != nil {
		return fmt.Errorf("failed to determine whether schemes should be auto created: %w", err)
	}

	if !isAutocreateSchemesNeeded {
		return nil
	}

	isAutocreateSchemesEnabled, err := p.isAutocreateSchemesEnabled()
	if err != nil {
		return fmt.Errorf("schemes should be auto created, but failed to check whether autocreate schemes is enabled: %s", err)
	}

	if !isAutocreateSchemesEnabled {
		return fmt.Errorf("no schemes found and autocreate schemes is disabled")
	}

	if err := p.autocreateSchemes(); err != nil {
		return fmt.Errorf("autocreating schemes failed: %s", err)
	}

	return nil
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

func (p XcodeProj) isAutocreateSchemesNeeded() (bool, error) {
	// if a shared scheme given -> no
	// if a current user's user scheme given -> no
	// if no schemes (no shared, no user scheme) given -> yes
	// if another user's user scheme given -> yes
	sharedSchemeFilePaths, err := p.sharedSchemeFilePaths()
	if err != nil {
		return false, err
	}
	if len(sharedSchemeFilePaths) > 0 {
		return false, nil
	}

	// TODO: should we check for user schemes of the current user?
	// 	It is unlikely that the current user has user schemes in a CI environment.

	// TODO: should we handle config errors?
	// 	e.g.: when the user has custom shared schemes and shared data is gitignored, but user data isn't
	return true, nil
}

func (p XcodeProj) autocreateSchemes() error {
	schemes := p.ReCreateSchemes()
	for _, scheme := range schemes {
		if err := p.SaveSharedScheme(scheme); err != nil {
			return err
		}
	}
	return nil
}

func (p XcodeProj) sharedSchemeFilePaths() ([]string, error) {
	// <project_name>.xcodeproj/xcshareddata/xcschemes/<scheme_name>.xcscheme
	sharedSchemesDir := filepath.Join(p.Path, "xcshareddata", "xcschemes")
	return listSchemeFilePaths(sharedSchemesDir)
}

//func (p XcodeProj) userSchemesDir() (string, error) {
//	// <project_name>.xcodeproj/xcuserdata/<current_user>.xcuserdatad/xcschemes/<scheme_name>.xcscheme
//	currentUser, err := user.Current()
//	if err != nil {
//		return "", err
//	}
//
//	username := currentUser.Username
//
//	return filepath.Join(p.Path, "xcuserdata", username+".xcuserdatad", "xcschemes"), nil
//}
//
//func (p XcodeProj) userSchemeFilePaths() ([]string, error) {
//	userSchemesDir, err := p.userSchemesDir()
//	if err != nil {
//		return nil, err
//	}
//	return listSchemeFilePaths(userSchemesDir)
//}
//
//func (p XcodeProj) schemeNamesFromSchememanagementFile() ([]string, error) {
//	// <project_name>.xcodeproj/xcuserdata/<current_user>.xcuserdatad/xcschemes/xcschememanagement.plist
//	userSchemesDir, err := p.userSchemesDir()
//	if err != nil {
//		return nil, err
//	}
//	schemeManagementPth := filepath.Join(userSchemesDir, "xcschememanagement.plist")
//	schemeManagementContent, err := os.ReadFile(schemeManagementPth)
//	if err != nil {
//		if errors.Is(err, os.ErrNotExist) {
//			return nil, nil
//		}
//
//		return nil, err
//	}
//
//	var schemeManagement serialized.Object
//	if _, err := plist.Unmarshal(schemeManagementContent, &schemeManagement); err != nil {
//		return nil, err
//	}
//
//	schemeUserState, err := schemeManagement.Object("SchemeUserState")
//	if err != nil {
//		return nil, err
//	}
//
//	schemeNames := schemeUserState.Keys()
//	return schemeNames, nil
//}

func listSchemeFilePaths(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
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
