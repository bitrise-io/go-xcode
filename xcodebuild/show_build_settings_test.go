package xcodebuild

import (
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

const testBuildSettingsOut = `Build settings for action build and target sample-apps-osx-10-12:
    VERSION_INFO_STRING = "@(#)PROGRAM:sample-apps-osx-10-12  PROJECT:sample-apps-osx-10-12-"
    AVAILABLE_PLATFORMS = appletvos appletvsimulator iphoneos iphonesimulator macosx watchos watchsimulator
    EXCLUDED_RECURSIVE_SEARCH_PATH_SUBDIRECTORIES = *.nib *.lproj *.framework *.gch *.xcode* *.xcassets (*) .DS_Store CVS .svn .git .hg *.pbproj *.pbxproj
    BUILD_STYLE =
    ACTION = build
    SDK_VERSION_MINOR = 1200`
