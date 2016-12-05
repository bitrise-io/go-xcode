package xcodeproj

import (
	"github.com/bitrise-io/go-utils/pathutil"
)

// WorkspaceModel ...
type WorkspaceModel struct {
	Projects []ProjectModel
}

// NewWorkspace ...
func NewWorkspace(xcworkspacePth string) (WorkspaceModel, error) {
	workspace := WorkspaceModel{}

	projects, err := WorkspaceProjectReferences(xcworkspacePth)
	if err != nil {
		return WorkspaceModel{}, err
	}

	for _, xcodeprojPth := range projects {
		if exist, err := pathutil.IsPathExists(xcodeprojPth); err != nil {
			return WorkspaceModel{}, err
		} else if !exist {
			continue
		}

		project, err := NewProject(xcodeprojPth)
		if err != nil {
			return WorkspaceModel{}, err
		}

		workspace.Projects = append(workspace.Projects, project)
	}

	return workspace, nil
}
