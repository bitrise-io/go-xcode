package xcodeproj

import (
	"fmt"

	"github.com/bitrise-io/go-utils/pathutil"
)

// WorkspaceModel ...
type WorkspaceModel struct {
	Pth            string
	Projects       []ProjectModel
	IsPodWorkspace bool
}

// NewWorkspace ...
func NewWorkspace(xcworkspacePth string) (WorkspaceModel, error) {
	workspace := WorkspaceModel{
		Pth: xcworkspacePth,
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
