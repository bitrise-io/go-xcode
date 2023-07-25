package destination

import (
	"github.com/bitrise-io/go-xcode/v2/xcodeversion"
	"github.com/hashicorp/go-version"
)

func isIOSRuntimeSupportedByXcode(runtimeVersion *version.Version, xcodeVersion xcodeversion.Version) bool {
	// Very simplified version of https://developer.apple.com/support/xcode/
	// Only considering major versions for simplicity
	var latestSupportedIOSRuntime = map[int64]int64{
		15: 17,
		14: 16,
		13: 15,
	}

	if len(runtimeVersion.Segments64()) == 0 || xcodeVersion.MajorVersion == 0 {
		return true
	}
	runtimeMajorVersion := runtimeVersion.Segments64()[0]

	latestSupportedMajorVersion, ok := latestSupportedIOSRuntime[xcodeVersion.MajorVersion]
	if !ok {
		return true
	}

	return latestSupportedMajorVersion >= runtimeMajorVersion
}
