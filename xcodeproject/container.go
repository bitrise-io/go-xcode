// Package xcodeproject is the common view of Xcode projects and workspaces: what a path to either
// gives a caller that does not care which of the two it holds.
package xcodeproject

import (
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcodeproj"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcscheme"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcworkspace"
)

// Container is an opened Xcode project or workspace. It answers for the bundle that was opened and
// for the projects that bundle references, so a caller never needs to know which of the two it has.
type Container interface {
	// Path is the absolute path of the .xcodeproj or .xcworkspace bundle that was opened.
	Path() string
	// Name is the bundle's file name without its extension.
	Name() string
	// Schemes returns the schemes Xcode considers when opening the bundle, in search order: a
	// workspace's own schemes first, then each referenced project's in file order. Each scheme
	// names the bundle holding it in ContainerPath.
	Schemes() ([]xcscheme.Scheme, error)
	// Scheme returns the named scheme, with the bundle holding it in ContainerPath.
	Scheme(name string) (xcscheme.Scheme, error)
}

// container is the Container over either a project or a workspace.
type container struct {
	project   *xcodeproj.XcodeProj   // the opened project, nil for a workspace
	workspace *xcworkspace.Workspace // the opened workspace, nil for a project
}

func (c *container) Path() string {
	if c.workspace != nil {
		return c.workspace.Path
	}
	return c.project.Path
}

func (c *container) Name() string {
	if c.workspace != nil {
		return c.workspace.Name
	}
	return c.project.Name
}

func (c *container) Schemes() ([]xcscheme.Scheme, error) {
	if c.workspace != nil {
		return c.workspace.Schemes()
	}
	return c.project.Schemes()
}

func (c *container) Scheme(name string) (xcscheme.Scheme, error) {
	if c.workspace != nil {
		return c.workspace.Scheme(name)
	}
	return c.project.Scheme(name)
}
