package xcodebuild

import (
	"os"
	"testing"

	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseShowBuildSettingsOutput(t *testing.T) {
	tests := []struct {
		name string
		out  string
		want serialized.Object
	}{
		{
			name: "key value pairs",
			out: `Build settings for action build and target App:
    ACTION = build
    ALTERNATE_GROUP = staff
    SDKROOT = iphoneos18.0`,
			want: serialized.Object{
				"ACTION":          "build",
				"ALTERNATE_GROUP": "staff",
				"SDKROOT":         "iphoneos18.0",
			},
		},
		{
			name: "only the first = separates key from value",
			out:  `    OTHER_LDFLAGS = -Wl,-U,_foo=bar`,
			want: serialized.Object{"OTHER_LDFLAGS": "-Wl,-U,_foo=bar"},
		},
		{
			name: "surrounding quotes are stripped",
			out:  `    PRODUCT_NAME = "My App"`,
			want: serialized.Object{"PRODUCT_NAME": "My App"},
		},
		{
			name: "empty values are kept",
			out:  `    PROVISIONING_PROFILE_SPECIFIER = `,
			want: serialized.Object{"PROVISIONING_PROFILE_SPECIFIER": ""},
		},
		{
			name: "lines without = are ignored",
			out: `Build settings for action build and target App:
    ACTION = build`,
			want: serialized.Object{"ACTION": "build"},
		},
		{
			name: "empty output",
			out:  "",
			want: serialized.Object{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseShowBuildSettingsOutput(tt.out)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Regression test for v1 fix 0c84f25: in multi-target scheme output the first block is the main target.
func TestParseShowBuildSettingsOutput_firstOccurrenceWins(t *testing.T) {
	out := `Build settings for action build and target App:
    PRODUCT_BUNDLE_IDENTIFIER = io.bitrise.App
    SDKROOT = iphoneos18.0

Build settings for action build and target AppTests:
    PRODUCT_BUNDLE_IDENTIFIER = io.bitrise.AppTests
    SDKROOT = iphoneos18.0`

	got, err := parseShowBuildSettingsOutput(out)
	require.NoError(t, err)

	assert.Equal(t, "io.bitrise.App", got["PRODUCT_BUNDLE_IDENTIFIER"],
		"the main target's value must win over a later target's")
}

func TestParseShowBuildSettingsOutput_lineLongerThanScannerLimit(t *testing.T) {
	out, err := os.ReadFile("./testdata/buildSettingsWithLongLine.txt")
	require.NoError(t, err)

	expectedLongValue, err := os.ReadFile("./testdata/expectedSingleLongLine.txt")
	require.NoError(t, err)

	got, err := parseShowBuildSettingsOutput(string(out))
	require.NoError(t, err)

	assert.Equal(t, serialized.Object{
		"ACTION":                      "build",
		"AD_HOC_CODE_SIGNING_ALLOWED": "NO",
		"REALLY_LONG_LINE":            string(expectedLongValue),
		"ALTERNATE_GROUP":             "staff",
		"BUILD_STYLE":                 "fast",
	}, got)
}

func TestShowBuildSettingsArgs(t *testing.T) {
	tests := []struct {
		name        string
		projectPath string
		nameFlag    string
		targetName  string
		config      string
		extraArgs   []string
		want        []string
	}{
		{
			name:        "project and target",
			projectPath: "/p/App.xcodeproj",
			nameFlag:    targetFlag,
			targetName:  "App",
			config:      "Release",
			want:        []string{"-project", "/p/App.xcodeproj", "-target", "App", "-configuration", "Release", "-showBuildSettings", "-json"},
		},
		{
			name:        "a workspace path is passed as -workspace",
			projectPath: "/p/App.xcworkspace",
			nameFlag:    schemeFlag,
			targetName:  "App",
			config:      "Debug",
			want:        []string{"-workspace", "/p/App.xcworkspace", "-scheme", "App", "-configuration", "Debug", "-showBuildSettings", "-json"},
		},
		{
			name:        "extra args come last",
			projectPath: "/p/App.xcodeproj",
			nameFlag:    targetFlag,
			targetName:  "App",
			config:      "Debug",
			extraArgs:   []string{"-destination", "generic/platform=iOS"},
			want: []string{"-project", "/p/App.xcodeproj", "-target", "App", "-configuration", "Debug", "-showBuildSettings", "-json",
				"-destination", "generic/platform=iOS"},
		},
		{
			name:       "empty container and configuration are omitted",
			nameFlag:   targetFlag,
			targetName: "App",
			want:       []string{"-target", "App", "-showBuildSettings", "-json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, showBuildSettingsArgs(tt.projectPath, tt.nameFlag, tt.targetName, tt.config, tt.extraArgs))
		})
	}
}

func TestParseShowBuildSettingsJSON(t *testing.T) {
	schemeWithTests, err := os.ReadFile("./testdata/showBuildSettingsSchemeWithTestTarget.json")
	require.NoError(t, err)
	quotedValue, err := os.ReadFile("./testdata/showBuildSettingsQuotedValue.json")
	require.NoError(t, err)

	tests := []struct {
		name   string
		out    []byte
		target string
		want   map[string]string // a subset of the result
		absent []string
	}{
		{
			name: "scheme: the first entry, the main target, without the test target's keys",
			out:  schemeWithTests,
			want: map[string]string{
				"PRODUCT_BUNDLE_IDENTIFIER":    "Bitrise.ios-simple-objc",
				"LD_RUNPATH_SEARCH_PATHS":      "@executable_path/Frameworks", // xcodebuild pads it with a space
				"GCC_PREPROCESSOR_DEFINITIONS": "DEBUG=1",
			},
			absent: []string{"BUNDLE_LOADER"},
		},
		{
			name:   "target: the entry named after it",
			out:    schemeWithTests,
			target: "ios-simple-objcTests",
			want:   map[string]string{"PRODUCT_BUNDLE_IDENTIFIER": "Bitrise.ios-simple-objcTests"},
		},
		{
			name:   "target not listed: the first entry",
			out:    schemeWithTests,
			target: "Nope",
			want:   map[string]string{"PRODUCT_BUNDLE_IDENTIFIER": "Bitrise.ios-simple-objc"},
		},
		{
			name: "quotes inside a value stay, which the text parser loses",
			out:  quotedValue,
			want: map[string]string{"OTHER_LDFLAGS": `-framework "SnapKit"`},
		},
		{
			name: "no entries, as for a Swift package scheme",
			out:  []byte("[\n\n]"),
			want: map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseShowBuildSettingsJSON(tt.out, tt.target)
			require.NoError(t, err)
			for key, value := range tt.want {
				assert.Equal(t, value, got[key], key)
			}
			for _, key := range tt.absent {
				assert.NotContains(t, got, key)
			}
		})
	}

	_, err = parseShowBuildSettingsJSON([]byte("Build settings for action build and target App:"), "")
	assert.Error(t, err, "text output is not JSON; the provider falls back to the text parser")
}
