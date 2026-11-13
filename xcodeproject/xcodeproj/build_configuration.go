package xcodeproj

import (
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
)

// BuildConfiguration is one named build configuration, such as Debug or Release.
type BuildConfiguration struct {
	ID   string
	Name string

	// buildSettings is the project's in-memory copy, shared by every copy of the Target, so a
	// Target fetched before an edit sees it. For project.pbxproj, configurations with the same ID
	// (one XCBuildConfiguration in several configuration lists) share it too. Only
	// Project.setBuildSetting (behind SetBuildSetting and ForceCodeSign) writes it, after the
	// project file accepted the value. It is not the project file's own representation: for
	// project.pbxproj it is a shallow clone, so nested values such as arrays are still shared with
	// the raw tree and must not be modified in place.
	buildSettings serialized.Object
}

// BuildSetting returns the value of key as declared in the project file: not variable-expanded
// and without .xcconfig overrides. For the effective value, use Project.TargetBuildSettings.
func (c BuildConfiguration) BuildSetting(key string) (string, bool) {
	return c.buildSettings.String(key)
}
