package destination

import (
	"fmt"

	"github.com/hashicorp/go-version"
)

func isRuntimeSupportedByXcode(runtimePlatform string, runtimeVersion *version.Version, xcodeVersion *version.Version) bool {
	// Disregard runtime patch version (Xcode 26.1.1 should work with Simulator 26.1.2).
	// Simulator runtime are usually only specified with major.minor (e.g., 18.2).
	runtimeVersionWithMinor := version.Must(version.NewVersion(
		fmt.Sprintf("%d.%d", runtimeVersion.Segments64()[0], runtimeVersion.Segments64()[1]),
	))
	xcodeMajor := xcodeVersion.Segments64()[0]
	xcodeVersionWithMinor := version.Must(version.NewVersion(
		fmt.Sprintf("%d.%d", xcodeMajor, xcodeVersion.Segments64()[1]),
	))

	if xcodeMajor >= 26 {
		// Xcode 26 unified Simulator and Xcode versioning
		return runtimeVersionWithMinor.LessThanOrEqual(xcodeVersionWithMinor)
	}

	// Very simplified version of https://developer.apple.com/support/xcode/
	// Only considering major versions for simplicity
	var xcodeVersionToSupportedSimulatorVersion = map[int64]map[string]int64{
		16: {
			string(IOS):     18,
			string(TvOS):    18,
			string(WatchOS): 11,
		},
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

	if len(runtimeVersion.Segments64()) == 0 || xcodeMajor == 0 {
		return true
	}
	runtimeMajorVersion := runtimeVersion.Segments64()[0]

	platformToLatestSupportedVersion, ok := xcodeVersionToSupportedSimulatorVersion[xcodeMajor]
	if !ok {
		return true
	}

	latestSupportedMajorVersion, ok := platformToLatestSupportedVersion[runtimePlatform]
	if !ok {
		return true
	}

	return latestSupportedMajorVersion >= runtimeMajorVersion
}
