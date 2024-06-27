package xcodeversion

import (
	"fmt"
	"strconv"
	"strings"
)

func getXcodeVersionFromXcodebuildOutput(outStr string) (Version, error) {
	split := strings.Split(outStr, "\n")
	if len(split) == 0 {
		return Version{}, fmt.Errorf("failed to parse xcodebuild version output (%s)", outStr)
	}

	filteredOutput, err := filterXcodeWarnings(split)
	if err != nil {
		return Version{}, err
	}

	xcodebuildVersion := filteredOutput[0]
	buildVersion := filteredOutput[1]

	split = strings.Split(xcodebuildVersion, " ")
	if len(split) != 2 {
		return Version{}, fmt.Errorf("failed to parse xcodebuild version output (%s)", outStr)
	}

	version := split[1]

	split = strings.Split(version, ".")
	majorVersionStr := split[0]
	var minorVersionStr string
	if len(split) > 1 {
		minorVersionStr = split[1]
	}
	var patchVersionStr string
	if len(split) > 2 {
		patchVersionStr = split[2]
	}

	majorVersion, err := strconv.ParseInt(majorVersionStr, 10, 32)
	if err != nil {
		return Version{}, fmt.Errorf("failed to parse xcodebuild major version (%s) as integer: %s", majorVersionStr, err)
	}

	minorVersion := int64(0)
	if minorVersionStr != "" {
		minorVersion, err = strconv.ParseInt(minorVersionStr, 10, 32)
		if err != nil {
			return Version{}, fmt.Errorf("failed to parse xcodebuild minor version (%s) as integer: %s", minorVersionStr, err)
		}
	}

	patchVersion := int64(0)
	if patchVersionStr != "" {
		patchVersion, err = strconv.ParseInt(patchVersionStr, 10, 32)
		if err != nil {
			return Version{}, fmt.Errorf("failed to parse xcodebuild patch version  (%s) as integer: %s", patchVersionStr, err)
		}

	}

	return Version{
		Version:       xcodebuildVersion,
		BuildVersion:  buildVersion,
		MajorVersion:  int(majorVersion),
		MinorVersion:  int(minorVersion),
		PatchVersions: int(patchVersion),
	}, nil
}

func filterXcodeWarnings(cmdOutputLines []string) ([]string, error) {
	firstLineIndex := -1
	for i, line := range cmdOutputLines {
		if strings.HasPrefix(line, "Xcode ") {
			firstLineIndex = i
			break
		}
	}

	if firstLineIndex < 0 {
		return []string{}, fmt.Errorf("couldn't find Xcode version in output: %s", cmdOutputLines)
	}

	return cmdOutputLines[firstLineIndex:], nil
}
