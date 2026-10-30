package xcproj

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testdata/ios-sample-manual-signing.xcproj is what Xcode 27.2 wrote after switching the app
// target of testdata/ios-sample.xcproj to manual signing for Release only.
func TestEditor_xcodeManualSigning(t *testing.T) {
	automatic, err := os.ReadFile(filepath.Join("testdata", "ios-sample.xcproj"))
	require.NoError(t, err)
	manual, err := os.ReadFile(filepath.Join("testdata", "ios-sample-manual-signing.xcproj"))
	require.NoError(t, err)

	editor, err := NewEditor(automatic)
	require.NoError(t, err)

	const target = "sample-apps-ios-json-xcproj"
	for _, setting := range []struct{ key, value string }{
		{key: "CODE_SIGN_IDENTITY[sdk=iphoneos*]", value: "Apple Development"},
		{key: "CODE_SIGN_STYLE", value: "Manual"},
		{key: "DEVELOPMENT_TEAM", value: "DT2C2FZ7U2"},
		{key: "DEVELOPMENT_TEAM[sdk=iphoneos*]", value: "DT2C2FZ7U2"},
		{key: "PROVISIONING_PROFILE_SPECIFIER", value: "Sample Apps iOS JSON XCProj"},
		{key: "PROVISIONING_PROFILE_SPECIFIER[sdk=iphoneos*]", value: "Sample Apps iOS JSON XCProj"},
	} {
		require.NoError(t, editor.SetTargetBuildSetting(target, "Release", setting.key, setting.value))
	}

	assert.Equal(t, string(manual), string(editor.Content()))
}

func TestEditor_noEdit(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "ios-sample.xcproj"))
	require.NoError(t, err)

	editor, err := NewEditor(content)
	require.NoError(t, err)

	assert.Equal(t, string(content), string(editor.Content()))
}

func TestEditor_SetTargetBuildSetting(t *testing.T) {
	const project = `{
  "configurations": [
    "Debug",
    "Release",
  ],
  "targets": [
    {
      "name": "App",
      "build-settings": {
%s
      },
    },
  ],
}`
	tests := []struct {
		name          string
		settings      string
		configuration string
		key           string
		value         string
		want          string
	}{
		{
			name: "replaces the configuration's own key",
			settings: `        "CODE_SIGN_STYLE[config=Debug]": "Automatic",
        "CODE_SIGN_STYLE[config=Release]": "Automatic",`,
			configuration: "Release",
			key:           "CODE_SIGN_STYLE",
			value:         "Manual",
			want: `        "CODE_SIGN_STYLE[config=Debug]": "Automatic",
        "CODE_SIGN_STYLE[config=Release]": "Manual",`,
		},
		{
			name: "splits a shared key",
			settings: `        "CODE_SIGN_STYLE": "Automatic",
        "PRODUCT_NAME": "App",`,
			configuration: "Release",
			key:           "CODE_SIGN_STYLE",
			value:         "Manual",
			want: `        "CODE_SIGN_STYLE[config=Debug]": "Automatic",
        "CODE_SIGN_STYLE[config=Release]": "Manual",
        "PRODUCT_NAME": "App",`,
		},
		{
			name: "keeps an existing override when splitting",
			settings: `        "CODE_SIGN_STYLE": "Automatic",
        "CODE_SIGN_STYLE[config=Debug]": "Manual",`,
			configuration: "Release",
			key:           "CODE_SIGN_STYLE",
			value:         "Manual",
			want: `        "CODE_SIGN_STYLE[config=Debug]": "Manual",
        "CODE_SIGN_STYLE[config=Release]": "Manual",`,
		},
		{
			name:          "splits the only key, keeping the trailing comma",
			settings:      `        "CODE_SIGN_STYLE": "Automatic",`,
			configuration: "Debug",
			key:           "CODE_SIGN_STYLE",
			value:         "Manual",
			want: `        "CODE_SIGN_STYLE[config=Debug]": "Manual",
        "CODE_SIGN_STYLE[config=Release]": "Automatic",`,
		},
		{
			name:          "splits an array value",
			settings:      `        "OTHER_LDFLAGS": ["-ObjC", "-lz"],`,
			configuration: "Debug",
			key:           "OTHER_LDFLAGS",
			value:         "-ObjC",
			want: `        "OTHER_LDFLAGS[config=Debug]": "-ObjC",
        "OTHER_LDFLAGS[config=Release]": ["-ObjC", "-lz"],`,
		},
		{
			name:          "splits a key with an sdk condition",
			settings:      `        "CODE_SIGN_IDENTITY[sdk=iphoneos*]": "Apple Development",`,
			configuration: "Release",
			key:           "CODE_SIGN_IDENTITY[sdk=iphoneos*]",
			value:         "Apple Distribution",
			want: `        "CODE_SIGN_IDENTITY[config=Debug][sdk=iphoneos*]": "Apple Development",
        "CODE_SIGN_IDENTITY[config=Release][sdk=iphoneos*]": "Apple Distribution",`,
		},
		{
			name:          "adds a missing key at its sorted place",
			settings:      "        \"CODE_SIGN_STYLE\": \"Automatic\",\n        \"PRODUCT_NAME\": \"App\",",
			configuration: "Release",
			key:           "DEVELOPMENT_TEAM",
			value:         "ABCDE12345",
			want: `        "CODE_SIGN_STYLE": "Automatic",
        "DEVELOPMENT_TEAM[config=Release]": "ABCDE12345",
        "PRODUCT_NAME": "App",`,
		},
		{
			name:          "adds a missing key after the last one, keeping the trailing comma",
			settings:      `        "CODE_SIGN_STYLE": "Automatic",`,
			configuration: "Release",
			key:           "PROVISIONING_PROFILE",
			value:         "8a2f3c1d-0000-0000-0000-000000000000",
			want: `        "CODE_SIGN_STYLE": "Automatic",
        "PROVISIONING_PROFILE[config=Release]": "8a2f3c1d-0000-0000-0000-000000000000",`,
		},
		{
			name:          "clears a value",
			settings:      `        "PROVISIONING_PROFILE_SPECIFIER[config=Release]": "Sample",`,
			configuration: "Release",
			key:           "PROVISIONING_PROFILE_SPECIFIER",
			value:         "",
			want:          `        "PROVISIONING_PROFILE_SPECIFIER[config=Release]": "",`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			editor, err := NewEditor(fmt.Appendf(nil, project, tt.settings))
			require.NoError(t, err)

			require.NoError(t, editor.SetTargetBuildSetting("App", tt.configuration, tt.key, tt.value))

			assert.Equal(t, fmt.Sprintf(project, tt.want), string(editor.Content()))
		})
	}
}

func TestEditor_SetTargetBuildSetting_errors(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		target        string
		configuration string
		key           string
		wantErr       string
	}{
		{
			name:          "unknown configuration",
			content:       `{"configurations": ["Debug"], "targets": [{"name": "App", "build-settings": {"A": "1"}}]}`,
			target:        "App",
			configuration: "Release",
			key:           "KEY",
			wantErr:       "project has no configuration Release",
		},
		{
			name:          "unknown target",
			content:       `{"configurations": ["Debug"], "targets": [{"name": "App", "build-settings": {"A": "1"}}]}`,
			target:        "Other",
			configuration: "Debug",
			key:           "KEY",
			wantErr:       "project has no target Other",
		},
		{
			name:          "target without build settings",
			content:       `{"configurations": ["Debug"], "targets": [{"name": "App"}]}`,
			target:        "App",
			configuration: "Debug",
			key:           "KEY",
			wantErr:       "target App has no build settings",
		},
		{
			name:          "project without targets",
			content:       `{"configurations": ["Debug"]}`,
			target:        "App",
			configuration: "Debug",
			key:           "KEY",
			wantErr:       "project has no targets",
		},
		{
			name:          "key with a configuration condition",
			content:       `{"configurations": ["Debug"], "targets": [{"name": "App", "build-settings": {"A": "1"}}]}`,
			target:        "App",
			configuration: "Debug",
			key:           "KEY[config=Debug]",
			wantErr:       "invalid build setting key: KEY[config=Debug]",
		},
		{
			name:          "malformed key",
			content:       `{"configurations": ["Debug"], "targets": [{"name": "App", "build-settings": {"A": "1"}}]}`,
			target:        "App",
			configuration: "Debug",
			key:           "KEY[sdk=iphoneos*",
			wantErr:       "invalid build setting key: KEY[sdk=iphoneos*",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			editor, err := NewEditor([]byte(tt.content))
			require.NoError(t, err)

			err = editor.SetTargetBuildSetting(tt.target, tt.configuration, tt.key, "value")
			require.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestNewEditor_JSON5OnlySyntax(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "hand-edited-json5.xcproj"))
	require.NoError(t, err)

	_, err = NewEditor(content)
	require.ErrorContains(t, err, "line 5, column 3")
}
