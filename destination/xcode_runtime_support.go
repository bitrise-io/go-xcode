package destination

func isIOSRuntimeSupportedByXcode(runtimeMajorVersion, xcodeMajorVersion int64) bool {
	// Very simplified version of https://developer.apple.com/support/xcode/
	// Only considering major versions for simplicity
	var latestSupportedIOSRuntime = map[int64]int64{
		15: 17,
		14: 16,
		13: 15,
	}

	latestSupportedMajorVersion, ok := latestSupportedIOSRuntime[runtimeMajorVersion]
	if !ok {
		return true
	}

	return latestSupportedMajorVersion >= runtimeMajorVersion
}
