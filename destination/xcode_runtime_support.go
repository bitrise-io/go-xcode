package destination

import (
	"github.com/bitrise-io/go-xcode/v2/xcodeversion"
	"github.com/hashicorp/go-version"
)

func isRuntimeSupportedByXcode(runtimePlatform string, runtimeVersion *version.Version, xcodeVersion xcodeversion.Version) bool {
	// Very simplified version of https://developer.apple.com/support/xcode/
	// Only considering major versions for simplicity
	var xcodeVersionToSupportedRuntimes = map[int64]map[string]int64{
		15: {
			string(IOS):     17,
			string(TvOS):    17,
			string(WatchOS): 10,
		},
		14: {
			string(IOS):     16,
			string(TvOS):    16,
			string(WatchOS): 9,
		},
		13: {
			string(IOS):     15,
			string(TvOS):    15,
			string(WatchOS): 8,
		},
	}

	if len(runtimeVersion.Segments64()) == 0 || xcodeVersion.MajorVersion == 0 {
		return true
	}
	runtimeMajorVersion := runtimeVersion.Segments64()[0]

	platformToLatestSupportedVersion, ok := xcodeVersionToSupportedRuntimes[xcodeVersion.MajorVersion]
	if !ok {
		return true
	}

	latestSupportedMajorVersion, ok := platformToLatestSupportedVersion[runtimePlatform]
	if !ok {
		return true
	}

	return latestSupportedMajorVersion >= runtimeMajorVersion
}
