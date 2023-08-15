package xcworkspace

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-xcode/xcodebuild"
	"github.com/bitrise-io/go-xcode/xcodeproject/serialized"
	"github.com/bitrise-io/go-xcode/xcodeproject/xcodeproj"
	"github.com/bitrise-io/go-xcode/xcodeproject/xcscheme"
	"golang.org/x/text/unicode/norm"
)

const (
	// XCWorkspaceExtension ...
	XCWorkspaceExtension = ".xcworkspace"
)

// Workspace represents an Xcode workspace
type Workspace struct {
	FileRefs []FileRef `xml:"FileRef"`
	Groups   []Group   `xml:"Group"`

	Name string
	Path string
}

// Open ...
func Open(pth string) (Workspace, error) {
	contentsPth := filepath.Join(pth, "contents.xcworkspacedata")
	b, err := fileutil.ReadBytesFromFile(contentsPth)
	if err != nil {
		return Workspace{}, err
	}

	var workspace Workspace
	if err := xml.Unmarshal(b, &workspace); err != nil {
		return Workspace{}, fmt.Errorf("failed to unmarshal workspace file: %s, error: %s", pth, err)
	}

	workspace.Name = strings.TrimSuffix(filepath.Base(pth), filepath.Ext(pth))
	workspace.Path = pth

	return workspace, nil
}

// Schemes returns the schemes considered by Xcode, when opening the given workspace.
// The considered schemes are the workspace shared schemes, the workspace user schemes (for the current user)
// and the embedded project's schemes (XcodeProj.Schemes).
func (w Workspace) Schemes() (map[string][]xcscheme.Scheme, error) {
	log.TDebugf("Searching schemes in workspace: %s", w.Path)

	schemesByContainer := map[string][]xcscheme.Scheme{}

	sharedSchemes, err := w.sharedSchemes()
	if err != nil {
		return nil, err
	}

	userSchemes, err := w.userSchemes()
	if err != nil {
		return nil, err
	}

	workspaceSchemes := append(sharedSchemes, userSchemes...)

	log.TDebugf("%d scheme(s) found", len(workspaceSchemes))
	schemesByContainer[w.Path] = workspaceSchemes

	// project schemes
	projectLocations, err := w.ProjectFileLocations()
	if err != nil {
		return nil, err
	}

	for _, projectLocation := range projectLocations {
		if exist, err := pathutil.IsPathExists(projectLocation); err != nil {
			return nil, fmt.Errorf("failed to check if project exist at: %s, error: %s", projectLocation, err)
		} else if !exist {
			// at this point we are interested the schemes visible for the workspace
			continue
		}

		project, err := xcodeproj.Open(projectLocation)
		if err != nil {
			return nil, err
		}

		projectSchemes, err := project.Schemes()
		if err != nil {
			return nil, err
		}

		schemesByContainer[project.Path] = projectSchemes
	}

	return schemesByContainer, nil
}

// Scheme returns the scheme by name, and it's container's absolute path.
func (w Workspace) Scheme(name string) (*xcscheme.Scheme, string, error) {
	schemesByContainer, err := w.Schemes()
	if err != nil {
		return nil, "", err
	}

	normName := norm.NFC.String(name)
	for container, schemes := range schemesByContainer {
		for _, scheme := range schemes {
			if norm.NFC.String(scheme.Name) == normName {
				return &scheme, container, nil
			}
		}
	}

	return nil, "", xcscheme.NotFoundError{Scheme: name, Container: w.Name}
}

// SchemeBuildSettings ...
func (w Workspace) SchemeBuildSettings(scheme, configuration string, customOptions ...string) (serialized.Object, error) {
	log.TDebugf("Fetching %s scheme build settings", scheme)

	commandModel := xcodebuild.NewShowBuildSettingsCommand(w.Path)
	commandModel.SetScheme(scheme)
	commandModel.SetConfiguration(configuration)
	commandModel.SetCustomOptions(customOptions)

	object, err := commandModel.RunAndReturnSettings()

	log.TDebugf("Fetched %s scheme build settings", scheme)

	return object, err
}

// FileLocations ...
func (w Workspace) FileLocations() ([]string, error) {
	var fileLocations []string

	for _, fileRef := range w.FileRefs {
		pth, err := fileRef.AbsPath(filepath.Dir(w.Path))
		if err != nil {
			return nil, err
		}

		fileLocations = append(fileLocations, pth)
	}

	for _, group := range w.Groups {
		groupFileLocations, err := group.FileLocations(filepath.Dir(w.Path))
		if err != nil {
			return nil, err
		}

		fileLocations = append(fileLocations, groupFileLocations...)
	}

	return fileLocations, nil
}

// ProjectFileLocations ...
func (w Workspace) ProjectFileLocations() ([]string, error) {
	var projectLocations []string
	fileLocations, err := w.FileLocations()
	if err != nil {
		return nil, err
	}
	for _, fileLocation := range fileLocations {
		if xcodeproj.IsXcodeProj(fileLocation) {
			projectLocations = append(projectLocations, fileLocation)
		}
	}
	return projectLocations, nil
}

func (w Workspace) sharedSchemes() ([]xcscheme.Scheme, error) {
	sharedSchemeFilePaths, err := w.sharedSchemeFilePaths()
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

func (w Workspace) sharedSchemeFilePaths() ([]string, error) {
	// <workspace_name>.xcworkspace/xcshareddata/xcschemes/<scheme_name>.xcscheme
	sharedSchemesDir := filepath.Join(w.Path, "xcshareddata", "xcschemes")
	return listSchemeFilePaths(sharedSchemesDir)
}

func (w Workspace) userSchemes() ([]xcscheme.Scheme, error) {
	userSchemeFilePaths, err := w.userSchemeFilePaths()
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

func (w Workspace) userSchemeFilePaths() ([]string, error) {
	// <workspace_name>.xcworkspace/xcuserdata/<current_user>.xcuserdatad/xcschemes/<scheme_name>.xcscheme
	userSchemesDir, err := w.userSchemesDir()
	if err != nil {
		return nil, err
	}
	return listSchemeFilePaths(userSchemesDir)
}

func (w Workspace) userSchemesDir() (string, error) {
	// <workspace_name>.xcworkspace/xcuserdata/<current_user>.xcuserdatad/xcschemes/
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	username := currentUser.Username

	return filepath.Join(w.Path, "xcuserdata", username+".xcuserdatad", "xcschemes"), nil
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
