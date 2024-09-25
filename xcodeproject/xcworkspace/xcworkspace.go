package xcworkspace

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-xcode/v2/xcodebuild"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcodeproj"
)

const (
	// XCWorkspaceExtension ...
	XCWorkspaceExtension = ".xcworkspace"
)

// Workspace represents an Xcode workspace
type Workspace struct {
	Name     string
	Path     string
	FileRefs []FileRef `xml:"FileRef"`
	Groups   []Group   `xml:"Group"`

	xcodebuildFactory xcodebuild.Factory
}

// Open ...
func NewFromFile(pth string, xcodebuildFactory xcodebuild.Factory) (Workspace, error) {
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

// SchemeBuildSettings ...
func (w Workspace) SchemeBuildSettings(scheme, configuration string, additionalArgs ...string) (serialized.Object, error) {
	log.TDebugf("Fetching %s scheme build settings", scheme)

	cmd := w.xcodebuildFactory.Create(&xcodebuild.CommandOptions{
		Workspace:         w.Path,
		Scheme:            scheme,
		Configuration:     configuration,
		ShowBuildSettings: true,
	}, nil, nil, additionalArgs, nil)

	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			// TODO: check if output is in a sensible size
			fmt.Println(out)
		}
		return nil, err
	}

	log.TDebugf("Fetched %s scheme build settings", scheme)

	return xcodebuild.ParseShowBuildSettingsCommandOutput(out)
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
