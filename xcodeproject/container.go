// Package xcodeproject is the common view of Xcode projects and workspaces: what a path to either
// gives a caller that does not care which of the two it holds.
package xcodeproject

import (
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcodeproj"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcscheme"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcworkspace"
)

// Container is an opened Xcode project or workspace, so a caller never needs to know which of the
// two it has.
type Container interface {
	// Path is the absolute path of the .xcodeproj or .xcworkspace bundle that was opened.
	Path() string
	// Name is the bundle's file name without its extension.
	Name() string
	// Scheme returns the named scheme and the path of the bundle holding it: for a workspace, either
	// the workspace itself or one of the projects it references.
	Scheme(name string) (xcscheme.Scheme, string, error)
}

// container is the Container over either a project or a workspace.
type container struct {
	project   *xcodeproj.XcodeProj   // the opened project, nil for a workspace
	workspace *xcworkspace.Workspace // the opened workspace, nil for a project
}

// Path implements Container.
func (c *container) Path() string {
	if c.workspace != nil {
		return c.workspace.Path
	}
	return c.project.Path
}

// Name implements Container.
func (c *container) Name() string {
	if c.workspace != nil {
		return c.workspace.Name
	}
	return c.project.Name
}

// Scheme implements Container.
func (c *container) Scheme(name string) (xcscheme.Scheme, string, error) {
	if c.workspace != nil {
		return c.workspace.Scheme(name)
	}
	return c.project.Scheme(name)
}
