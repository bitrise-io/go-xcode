package xcodeproj

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/pathutil"
)

// WorkspaceModel ...
type WorkspaceModel struct {
	Pth            string
	Name           string
	Projects       []ProjectModel
	IsPodWorkspace bool
}

// NewWorkspace ...
func NewWorkspace(xcworkspacePth string) (WorkspaceModel, error) {
	workspace := WorkspaceModel{
		Pth:  xcworkspacePth,
		Name: strings.TrimSuffix(filepath.Base(xcworkspacePth), filepath.Ext(xcworkspacePth)),
	}

	projects, err := WorkspaceProjectReferences(xcworkspacePth)
	if err != nil {
		return WorkspaceModel{}, err
	}

	for _, xcodeprojPth := range projects {
		if exist, err := pathutil.IsPathExists(xcodeprojPth); err != nil {
			return WorkspaceModel{}, err
		} else if !exist {
			return WorkspaceModel{}, fmt.Errorf("referred project (%s) not found", xcodeprojPth)
		}

		project, err := NewProject(xcodeprojPth)
		if err != nil {
			return WorkspaceModel{}, err
		}

		workspace.Projects = append(workspace.Projects, project)
	}

	return workspace, nil
}

// GetSharedSchemes ...
func (w WorkspaceModel) GetSharedSchemes() []SchemeModel {
	schemes := []SchemeModel{}
	for _, project := range w.Projects {
		schemes = append(schemes, project.SharedSchemes...)
	}
	return schemes
}

// GetTargets ...
func (w WorkspaceModel) GetTargets() []TargetModel {
	targets := []TargetModel{}
	for _, project := range w.Projects {
		targets = append(targets, project.Targets...)
	}
	return targets
}
