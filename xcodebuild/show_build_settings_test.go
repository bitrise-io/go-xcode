package xcodebuild

import (
	"os"
	"testing"

	"github.com/bitrise-io/go-xcode/xcodeproject/serialized"
	"github.com/stretchr/testify/require"
)

func Test_parseShowBuildSettingsOutput(t *testing.T) {
	tests := []struct {
		name string
		out  string
		want serialized.Object
	}{
		{
			name: "empty output",
			out:  "",
			want: serialized.Object{},
		},
		{
			name: "simple output",
			out: `    ACTION = build
    AD_HOC_CODE_SIGNING_ALLOWED = NO
    ALTERNATE_GROUP = staff`,
			want: serialized.Object{"ACTION": "build", "AD_HOC_CODE_SIGNING_ALLOWED": "NO", "ALTERNATE_GROUP": "staff"},
		},
		{
			name: "output header",
			out: `Build settings for action build and target ios-simple-objc:
    ACTION = build
    AD_HOC_CODE_SIGNING_ALLOWED = NO`,
			want: serialized.Object{"ACTION": "build", "AD_HOC_CODE_SIGNING_ALLOWED": "NO"},
		},
		{
			name: "Build setting without value",
			out:  `    ACTION = `,
			want: serialized.Object{"ACTION": ""},
		},
		{
			name: "Build setting without =",
			out:  `    ACTION `,
			want: serialized.Object{},
		},
		{
			name: "Build setting without key",
			out:  `    = `,
			want: serialized.Object{"": ""},
		},
		{
			name: "Split the first = ",
			out:  `    ACTION = build+=test`,
			want: serialized.Object{"ACTION": "build+=test"},
		},
		{
			name: "Complete build settings output",
			out:  testBuildSettingsOut,
			want: serialized.Object{
				"VERSION_INFO_STRING":                           "@(#)PROGRAM:sample-apps-osx-10-12  PROJECT:sample-apps-osx-10-12-",
				"AVAILABLE_PLATFORMS":                           "appletvos appletvsimulator iphoneos iphonesimulator macosx watchos watchsimulator",
				"EXCLUDED_RECURSIVE_SEARCH_PATH_SUBDIRECTORIES": "*.nib *.lproj *.framework *.gch *.xcode* *.xcassets (*) .DS_Store CVS .svn .git .hg *.pbproj *.pbxproj",
				"BUILD_STYLE":                                   "",
				"ACTION":                                        "build",
				"SDK_VERSION_MINOR":                             "1200",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseBuildSettings(tt.out)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestReadLongBuildSettingsLine(t *testing.T) {
	// The default line limit for a bufio.Scanner is 65000 characters, and this fie contains a line longer than that
	buildSettingsWithLongLine, err := os.ReadFile("./testdata/buildSettingsWithLongLine.txt")
	require.NoError(t, err)

	got, err := parseBuildSettings(string(buildSettingsWithLongLine))
	require.NoError(t, err)

	// Reading the same single long line value from a file, so it does not hurt test readability
	expectedSingleLongLine, err := os.ReadFile("./testdata/expectedSingleLongLine.txt")
	require.NoError(t, err)

	want := serialized.Object{
		"ACTION":                      "build",
		"AD_HOC_CODE_SIGNING_ALLOWED": "NO",
		"REALLY_LONG_LINE":            string(expectedSingleLongLine),
		"ALTERNATE_GROUP":             "staff",
		"BUILD_STYLE":                 "fast",
	}
	require.Equal(t, want, got)
}

const testBuildSettingsOut = `Build settings for action build and target sample-apps-osx-10-12:
    VERSION_INFO_STRING = "@(#)PROGRAM:sample-apps-osx-10-12  PROJECT:sample-apps-osx-10-12-"
    AVAILABLE_PLATFORMS = appletvos appletvsimulator iphoneos iphonesimulator macosx watchos watchsimulator
    EXCLUDED_RECURSIVE_SEARCH_PATH_SUBDIRECTORIES = *.nib *.lproj *.framework *.gch *.xcode* *.xcassets (*) .DS_Store CVS .svn .git .hg *.pbproj *.pbxproj
    BUILD_STYLE =
    ACTION = build
    SDK_VERSION_MINOR = 1200`
