package xcworkspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/text/unicode/norm"

	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcscheme"
)

const (
	xcschemeExtension           = ".xcscheme"
	autocreateSchemesSettingKey = "IDEWorkspaceSharedSettings_AutocreateContextsIfNeeded"
)

// containerSchemes are the schemes one container, the workspace or a project, holds.
type containerSchemes struct {
	path    string
	schemes []xcscheme.Scheme
}

// Schemes returns the schemes Xcode considers when opening the workspace, keyed by the path of the
// container that holds them: the workspace's own shared and user schemes under the workspace path,
// and each referenced project's schemes under that project's path. A project without schemes gets
// Xcode's default ones when the workspace's "Autocreate schemes" option is on.
func (w *Workspace) Schemes() (map[string][]xcscheme.Scheme, error) {
	containers, err := w.schemesByContainer()
	if err != nil {
		return nil, err
	}

	schemesByContainer := make(map[string][]xcscheme.Scheme, len(containers))
	for _, container := range containers {
		schemesByContainer[container.path] = container.schemes
	}

	return schemesByContainer, nil
}

// Scheme returns the scheme with the given name and the path of the container that holds it. The
// workspace's own schemes are searched first, then each project's in ProjectFileLocations order.
// Names are compared after Unicode normalisation.
func (w *Workspace) Scheme(name string) (xcscheme.Scheme, string, error) {
	containers, err := w.schemesByContainer()
	if err != nil {
		return xcscheme.Scheme{}, "", err
	}

	normalisedName := norm.NFC.String(name)
	for _, container := range containers {
		for _, scheme := range container.schemes {
			if norm.NFC.String(scheme.Name) == normalisedName {
				return scheme, container.path, nil
			}
		}
	}

	return xcscheme.Scheme{}, "", xcscheme.NotFoundError{Scheme: name, Container: w.Name}
}

// schemesByContainer lists the containers holding at least one scheme, the workspace first.
func (w *Workspace) schemesByContainer() ([]containerSchemes, error) {
	w.logger.TDebugf("Searching schemes in workspace: %s", w.Path)

	var containers []containerSchemes

	sharedSchemes, err := w.sharedSchemes()
	if err != nil {
		return nil, fmt.Errorf("failed to read shared schemes: %w", err)
	}

	userSchemes, err := w.userSchemes()
	if err != nil {
		return nil, fmt.Errorf("failed to read user schemes: %w", err)
	}

	workspaceSchemes := append(sharedSchemes, userSchemes...)

	w.logger.TDebugf("%d scheme(s) found", len(workspaceSchemes))
	if len(workspaceSchemes) > 0 {
		containers = append(containers, containerSchemes{path: w.Path, schemes: workspaceSchemes})
	}

	// project schemes
	projectLocations, err := w.ProjectFileLocations()
	if err != nil {
		return nil, fmt.Errorf("failed to get project locations from workspace: %w", err)
	}

	isAutocreateSchemesEnabled, err := w.isAutocreateSchemesEnabled()
	if err != nil {
		return nil, fmt.Errorf("failed to read the workspace autocreate scheme option: %w", err)
	}

	for _, projectLocation := range projectLocations {
		if exist, err := w.pathChecker.IsPathExists(projectLocation); err != nil {
			return nil, fmt.Errorf("failed to check if project (%s) exists: %w", projectLocation, err)
		} else if !exist {
			// at this point we are interested the schemes visible for the workspace
			continue
		}

		project, err := w.projectFactory.Open(projectLocation)
		if err != nil {
			return nil, fmt.Errorf("failed to open project (%s): %w", projectLocation, err)
		}

		projectSchemes, err := project.SchemesWithAutocreateOverride(isAutocreateSchemesEnabled)
		if err != nil {
			return nil, fmt.Errorf("failed to read project (%s) schemes: %w", projectLocation, err)
		}

		if len(projectSchemes) > 0 {
			containers = append(containers, containerSchemes{path: project.Path, schemes: projectSchemes})
		}
	}

	return containers, nil
}

func (w *Workspace) sharedSchemes() ([]xcscheme.Scheme, error) {
	// <workspace_name>.xcworkspace/xcshareddata/xcschemes/<scheme_name>.xcscheme
	schemes, err := w.readSchemes(filepath.Join(w.Path, "xcshareddata", "xcschemes"))
	if err != nil {
		return nil, err
	}

	for i := range schemes {
		schemes[i].IsShared = true
	}

	return schemes, nil
}

func (w *Workspace) userSchemes() ([]xcscheme.Scheme, error) {
	// <workspace_name>.xcworkspace/xcuserdata/<current_user>.xcuserdatad/xcschemes/<scheme_name>.xcscheme
	username, err := w.userProvider.CurrentUserName()
	if err != nil {
		return nil, fmt.Errorf("failed to get the current user: %w", err)
	}

	return w.readSchemes(filepath.Join(w.Path, "xcuserdata", username+".xcuserdatad", "xcschemes"))
}

// readSchemes returns the schemes in dir, in directory order. A missing dir holds none.
func (w *Workspace) readSchemes(dir string) ([]xcscheme.Scheme, error) {
	names, err := w.fileManager.ReadDirEntryNames(dir)
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

		scheme, err := w.schemeFactory.Open(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}

		schemes = append(schemes, scheme)
	}

	return schemes, nil
}

// isAutocreateSchemesEnabled defaults to true when the settings file or the key is missing.
func (w *Workspace) isAutocreateSchemesEnabled() (bool, error) {
	// <workspace_name>.xcworkspace/xcshareddata/WorkspaceSettings.xcsettings
	pth := filepath.Join(w.Path, "xcshareddata", "WorkspaceSettings.xcsettings")

	data, _, err := w.plistFileHandler.Read(pth)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return true, nil
		}
		return false, err
	}
	settings := serialized.Object(data)

	if !settings.Has(autocreateSchemesSettingKey) {
		return true, nil
	}

	autocreate, ok := settings.Bool(autocreateSchemesSettingKey)
	if !ok {
		return false, fmt.Errorf("%s is not a boolean", autocreateSchemesSettingKey)
	}

	return autocreate, nil
}
