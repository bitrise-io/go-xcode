// Package xcodeproject is the common view of Xcode projects and workspaces: what a path to either
// gives a caller that does not care which of the two it holds.
package xcodeproject

import (
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcscheme"
)

// Container is an opened Xcode project or workspace, so a caller never needs to know which of the
// two it has.
type Container interface {
	// Scheme returns the named scheme and the path of the bundle holding it: for a workspace, either
	// the workspace itself or one of the projects it references.
	Scheme(name string) (xcscheme.Scheme, string, error)
	// SchemeBuildSettings returns the effective build settings of the scheme's main target, resolved
	// by xcodebuild within the bundle: -workspace for a workspace, -project for a project. That is
	// the context the scheme builds in, so packages and the build cache are the build's own.
	SchemeBuildSettings(scheme, configuration string, additionalOptions ...string) (serialized.Object, error)
}

// bundle is what an opened project and an opened workspace both answer. *xcodeproj.XcodeProj and
// *xcworkspace.Workspace implement it.
type bundle interface {
	Scheme(name string) (xcscheme.Scheme, string, error)
	SchemeBuildSettings(scheme, configuration string, extraArgs ...string) (serialized.Object, error)
}

// container is the Container over either a project or a workspace: it keeps what both offer, so no
// method needs to know which of the two it holds.
type container struct {
	bundle bundle
}

// Scheme implements Container.
func (c *container) Scheme(name string) (xcscheme.Scheme, string, error) {
	return c.bundle.Scheme(name)
}

// SchemeBuildSettings implements Container.
func (c *container) SchemeBuildSettings(scheme, configuration string, additionalOptions ...string) (serialized.Object, error) {
	return c.bundle.SchemeBuildSettings(scheme, configuration, additionalOptions...)
}
