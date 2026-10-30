package xcproj

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBuildSettingKey(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want buildSettingKey
		ok   bool
	}{
		{name: "plain", key: "PRODUCT_NAME", want: buildSettingKey{name: "PRODUCT_NAME"}, ok: true},
		{
			name: "config",
			key:  "CODE_SIGN_STYLE[config=Release]",
			want: buildSettingKey{name: "CODE_SIGN_STYLE", conditions: []string{"config=Release"}},
			ok:   true,
		},
		{
			name: "config and sdk",
			key:  "CODE_SIGN_IDENTITY[config=Release][sdk=iphoneos*]",
			want: buildSettingKey{name: "CODE_SIGN_IDENTITY", conditions: []string{"config=Release", "sdk=iphoneos*"}},
			ok:   true,
		},
		{
			name: "config, sdk and arch",
			key:  "ARCHS[config=Debug][sdk=iphonesimulator*][arch=arm64]",
			want: buildSettingKey{name: "ARCHS", conditions: []string{"config=Debug", "sdk=iphonesimulator*", "arch=arm64"}},
			ok:   true,
		},
		{
			name: "empty pattern",
			key:  "X[config=]",
			want: buildSettingKey{name: "X", conditions: []string{"config="}},
			ok:   true,
		},
		{name: "empty key", key: "", ok: false},
		{name: "no name", key: "[config=Release]", ok: false},
		{name: "unterminated", key: "X[config=Release", ok: false},
		{name: "text between conditions", key: "X[config=Release]y[sdk=iphoneos*]", ok: false},
		{name: "text after conditions", key: "X[config=Release]y", ok: false},
		{name: "condition without equals", key: "X[sdk]", ok: false},
		{name: "empty condition", key: "X[]", ok: false},
		{name: "condition without parameter", key: "X[=x]", ok: false},
		{name: "nested bracket", key: "X[a[b]", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseBuildSettingKey(tt.key)

			require.Equal(t, tt.ok, ok)
			if !ok {
				return
			}
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.key, got.String(), "joining the parsed key must give the original back")
		})
	}
}

func TestSettingsForConfiguration(t *testing.T) {
	tests := []struct {
		name          string
		settings      map[string]any
		configuration string
		want          map[string]any
	}{
		{
			name: "unconditioned keys apply to every configuration",
			settings: map[string]any{
				"PRODUCT_NAME": "App",
			},
			configuration: "Release",
			want:          map[string]any{"PRODUCT_NAME": "App"},
		},
		{
			name: "keeps the matching configuration and drops the others",
			settings: map[string]any{
				"PRODUCT_NAME":                             "App",
				"DEBUG_INFORMATION_FORMAT[config=Debug]":   "dwarf",
				"DEBUG_INFORMATION_FORMAT[config=Release]": "dwarf-with-dsym",
			},
			configuration: "Release",
			want: map[string]any{
				"PRODUCT_NAME":             "App",
				"DEBUG_INFORMATION_FORMAT": "dwarf-with-dsym",
			},
		},
		{
			name: "there is no unconditioned base once configurations differ",
			settings: map[string]any{
				"CODE_SIGN_STYLE[config=Debug]":   "Automatic",
				"CODE_SIGN_STYLE[config=Release]": "Manual",
			},
			configuration: "Debug",
			want:          map[string]any{"CODE_SIGN_STYLE": "Automatic"},
		},
		{
			name: "sdk and arch conditions stay in the key",
			settings: map[string]any{
				"CODE_SIGN_IDENTITY[config=Release][sdk=iphoneos*]":       "Apple Distribution",
				"ARCHS[config=Release][sdk=iphonesimulator*][arch=arm64]": "arm64",
			},
			configuration: "Release",
			want: map[string]any{
				"CODE_SIGN_IDENTITY[sdk=iphoneos*]":       "Apple Distribution",
				"ARCHS[sdk=iphonesimulator*][arch=arm64]": "arm64",
			},
		},
		{
			name: "a configuration key overrides the unconditioned one",
			settings: map[string]any{
				"SWIFT_VERSION":                 "5.0",
				"SWIFT_VERSION[config=Release]": "6.0",
			},
			configuration: "Release",
			want:          map[string]any{"SWIFT_VERSION": "6.0"},
		},
		{
			name: "a configuration and sdk key overrides the sdk-only one",
			settings: map[string]any{
				"CODE_SIGN_IDENTITY[sdk=iphoneos*]":                 "Apple Development",
				"CODE_SIGN_IDENTITY[config=Release][sdk=iphoneos*]": "Apple Distribution",
			},
			configuration: "Release",
			want:          map[string]any{"CODE_SIGN_IDENTITY[sdk=iphoneos*]": "Apple Distribution"},
		},
		{
			name: "an sdk-only key applies to every configuration",
			settings: map[string]any{
				"CODE_SIGN_IDENTITY[sdk=iphoneos*]": "Apple Development",
			},
			configuration: "Debug",
			want:          map[string]any{"CODE_SIGN_IDENTITY[sdk=iphoneos*]": "Apple Development"},
		},
		{
			name: "array values are kept as they are",
			settings: map[string]any{
				"GCC_PREPROCESSOR_DEFINITIONS[config=Debug]": []any{"DEBUG=1", "$(inherited)"},
			},
			configuration: "Debug",
			want:          map[string]any{"GCC_PREPROCESSOR_DEFINITIONS": []any{"DEBUG=1", "$(inherited)"}},
		},
		{
			name: "a config condition that is not the first is not a configuration key",
			settings: map[string]any{
				"X[sdk=iphoneos*][config=Release]": "a",
			},
			configuration: "Release",
			want:          map[string]any{"X[sdk=iphoneos*][config=Release]": "a"},
		},
		{
			name: "a key with two config conditions matches no configuration",
			settings: map[string]any{
				"X[config=Debug][config=Release]": "a",
			},
			configuration: "Debug",
			want:          map[string]any{},
		},
		{
			name: "config is matched literally",
			settings: map[string]any{
				"X[config=Rel*]": "a",
				"Y[config=*]":    "b",
			},
			configuration: "Release",
			want:          map[string]any{},
		},
		{
			name: "malformed keys are kept as they are",
			settings: map[string]any{
				"X[config=Release": "a",
			},
			configuration: "Release",
			want:          map[string]any{"X[config=Release": "a"},
		},
		{
			name: "keys with a malformed condition are kept as they are for every configuration",
			settings: map[string]any{
				"X[config=Release][sdk]": "a",
				"Y[config=Release][]":    "b",
				"Z[config=Release][=x]":  "c",
			},
			configuration: "Debug",
			want: map[string]any{
				"X[config=Release][sdk]": "a",
				"Y[config=Release][]":    "b",
				"Z[config=Release][=x]":  "c",
			},
		},
		{
			name: "a malformed condition does not turn the key into a configuration key",
			settings: map[string]any{
				"X[config=Release][sdk]": "a",
			},
			configuration: "Release",
			want:          map[string]any{"X[config=Release][sdk]": "a"},
		},
		{
			name: "an unknown configuration only gets the unconditioned keys",
			settings: map[string]any{
				"PRODUCT_NAME":                  "App",
				"CODE_SIGN_STYLE[config=Debug]": "Automatic",
			},
			configuration: "Staging",
			want:          map[string]any{"PRODUCT_NAME": "App"},
		},
		{
			name:          "no settings gives an empty, non-nil map",
			settings:      nil,
			configuration: "Release",
			want:          map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SettingsForConfiguration(tt.settings, tt.configuration)

			require.NotNil(t, got)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Mirrors the Debug automatic, Release manual signing fixture written by Xcode 27.2.
func TestSettingsForConfiguration_ManualSigningRelease(t *testing.T) {
	settings := map[string]any{
		"PRODUCT_NAME": "$(TARGET_NAME)",
		"CODE_SIGN_IDENTITY[config=Release][sdk=iphoneos*]": "Apple Development",
		"CODE_SIGN_STYLE[config=Debug]":                     "Automatic",
		"CODE_SIGN_STYLE[config=Release]":                   "Manual",
		"DEVELOPMENT_TEAM[config=Debug]":                    "DT2C2FZ7U2",
		"DEVELOPMENT_TEAM[config=Release]":                  "DT2C2FZ7U2",
		"DEVELOPMENT_TEAM[config=Release][sdk=iphoneos*]":   "DT2C2FZ7U2",
		"PROVISIONING_PROFILE_SPECIFIER[config=Release]":    "Sample Apps iOS JSON XCProj",
	}

	assert.Equal(t, map[string]any{
		"PRODUCT_NAME":                      "$(TARGET_NAME)",
		"CODE_SIGN_IDENTITY[sdk=iphoneos*]": "Apple Development",
		"CODE_SIGN_STYLE":                   "Manual",
		"DEVELOPMENT_TEAM":                  "DT2C2FZ7U2",
		"DEVELOPMENT_TEAM[sdk=iphoneos*]":   "DT2C2FZ7U2",
		"PROVISIONING_PROFILE_SPECIFIER":    "Sample Apps iOS JSON XCProj",
	}, SettingsForConfiguration(settings, "Release"))

	assert.Equal(t, map[string]any{
		"PRODUCT_NAME":     "$(TARGET_NAME)",
		"CODE_SIGN_STYLE":  "Automatic",
		"DEVELOPMENT_TEAM": "DT2C2FZ7U2",
	}, SettingsForConfiguration(settings, "Debug"))
}
