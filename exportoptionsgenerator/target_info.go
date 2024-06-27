package exportoptionsgenerator

import (
	"github.com/bitrise-io/go-xcode/xcodeproject/serialized"
	"github.com/bitrise-io/go-xcode/xcodeproject/xcodeproj"
)

// TargetInfoProvider can determine a target's bundle id and codesign entitlements.
type TargetInfoProvider interface {
	TargetBundleID(target, configuration string) (string, error)
	TargetCodeSignEntitlements(target, configuration string) (serialized.Object, error)
}

// XcodebuildTargetInfoProvider implements TargetInfoProvider.
type XcodebuildTargetInfoProvider struct {
	xcodeProj *xcodeproj.XcodeProj
}

// TargetBundleID ...
func (b XcodebuildTargetInfoProvider) TargetBundleID(target, configuration string) (string, error) {
	return b.xcodeProj.TargetBundleID(target, configuration)
}

// TargetCodeSignEntitlements ...
func (b XcodebuildTargetInfoProvider) TargetCodeSignEntitlements(target, configuration string) (serialized.Object, error) {
	return b.xcodeProj.TargetCodeSignEntitlements(target, configuration)
}
