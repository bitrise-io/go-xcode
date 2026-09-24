package xcodeproj

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcodeproj/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func projectWithProvider(t *testing.T, provider BuildSettingsProvider) *XcodeProj {
	t.Helper()

	return &XcodeProj{
		Name:          "App",
		Path:          filepath.Join(t.TempDir(), "App.xcodeproj"),
		logger:        log.NewLogger(),
		buildSettings: provider,
		fileManager:   fileutil.NewFileManager(),
	}
}

func projectWithBuildSettings(t *testing.T, settings serialized.Object) *XcodeProj {
	t.Helper()

	provider := mocks.NewBuildSettingsProvider(t)
	provider.On("TargetBuildSettings", mock.Anything, mock.Anything, mock.Anything).Return(settings, nil)

	return projectWithProvider(t, provider)
}

func writePlist(t *testing.T, pth string, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(pth), 0755))
	require.NoError(t, os.WriteFile(pth, []byte(content), 0644))
}

func plistWith(entries string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
` + entries + `
</dict>
</plist>`
}

func TestXcodeProj_TargetBuildSettings(t *testing.T) {
	provider := mocks.NewBuildSettingsProvider(t)
	project := projectWithProvider(t, provider)
	provider.On("TargetBuildSettings", project.Path, "App", "Release", []string{"-destination", "generic/platform=iOS"}).
		Return(serialized.Object{"SDKROOT": "iphoneos"}, nil).Once()

	settings, err := project.TargetBuildSettings("App", "Release", "-destination", "generic/platform=iOS")
	require.NoError(t, err)
	assert.Equal(t, serialized.Object{"SDKROOT": "iphoneos"}, settings)
}

func TestXcodeProj_TargetBuildSettings_noProviderInjected(t *testing.T) {
	project := &XcodeProj{Path: "/p/App.xcodeproj"}

	_, err := project.TargetBuildSettings("App", "Debug")
	require.ErrorContains(t, err, "BuildSettingsProvider")
}

func TestXcodeProj_TargetBuildSettings_providerFailure(t *testing.T) {
	failure := errors.New("xcodebuild exploded")
	provider := mocks.NewBuildSettingsProvider(t)
	provider.On("TargetBuildSettings", mock.Anything, mock.Anything, mock.Anything).Return(nil, failure)

	_, err := projectWithProvider(t, provider).TargetBuildSettings("App", "Debug")
	assert.ErrorIs(t, err, failure)
}

func TestXcodeProj_TargetBundleID(t *testing.T) {
	t.Run("from the build setting", func(t *testing.T) {
		project := projectWithBuildSettings(t, serialized.Object{
			"PRODUCT_BUNDLE_IDENTIFIER": "io.bitrise.App",
		})

		bundleID, err := project.TargetBundleID("App", "Release")
		require.NoError(t, err)
		assert.Equal(t, "io.bitrise.App", bundleID)
	})

	t.Run("references in the build setting are expanded", func(t *testing.T) {
		project := projectWithBuildSettings(t, serialized.Object{
			"PRODUCT_BUNDLE_IDENTIFIER": "io.bitrise.$(PRODUCT_NAME:rfc1034identifier)",
			"PRODUCT_NAME":              "App",
		})

		bundleID, err := project.TargetBundleID("App", "Release")
		require.NoError(t, err)
		assert.Equal(t, "io.bitrise.App", bundleID)
	})

	t.Run("falls back to the Info.plist CFBundleIdentifier", func(t *testing.T) {
		project := projectWithBuildSettings(t, serialized.Object{
			"INFOPLIST_FILE": "App/Info.plist",
		})
		writePlist(t, filepath.Join(filepath.Dir(project.Path), "App", "Info.plist"),
			plistWith("\t<key>CFBundleIdentifier</key>\n\t<string>io.bitrise.FromPlist</string>"))

		bundleID, err := project.TargetBundleID("App", "Release")
		require.NoError(t, err)
		assert.Equal(t, "io.bitrise.FromPlist", bundleID)
	})

	t.Run("references in the Info.plist value are expanded too", func(t *testing.T) {
		project := projectWithBuildSettings(t, serialized.Object{
			"INFOPLIST_FILE": "App/Info.plist",
			"PRODUCT_NAME":   "App",
		})
		writePlist(t, filepath.Join(filepath.Dir(project.Path), "App", "Info.plist"),
			plistWith("\t<key>CFBundleIdentifier</key>\n\t<string>io.bitrise.$(PRODUCT_NAME)</string>"))

		bundleID, err := project.TargetBundleID("App", "Release")
		require.NoError(t, err)
		assert.Equal(t, "io.bitrise.App", bundleID)
	})

	t.Run("neither source available", func(t *testing.T) {
		project := projectWithBuildSettings(t, serialized.Object{})

		_, err := project.TargetBundleID("App", "Release")
		require.ErrorContains(t, err, "PRODUCT_BUNDLE_IDENTIFIER")
	})

	t.Run("Info.plist named but missing", func(t *testing.T) {
		project := projectWithBuildSettings(t, serialized.Object{
			"INFOPLIST_FILE": "App/Info.plist",
		})

		_, err := project.TargetBundleID("App", "Release")
		require.Error(t, err)
	})
}

func TestXcodeProj_TargetCodeSignEntitlements(t *testing.T) {
	t.Run("reads the entitlements file", func(t *testing.T) {
		project := projectWithBuildSettings(t, serialized.Object{
			"CODE_SIGN_ENTITLEMENTS": "App/App.entitlements",
		})
		writePlist(t, filepath.Join(filepath.Dir(project.Path), "App", "App.entitlements"),
			plistWith("\t<key>com.apple.developer.applesignin</key>\n\t<array>\n\t\t<string>Default</string>\n\t</array>"))

		entitlements, err := project.TargetCodeSignEntitlements("App", "Release")
		require.NoError(t, err)
		assert.True(t, entitlements.Has("com.apple.developer.applesignin"))
	})

	t.Run("no entitlements setting yields ErrEntitlementsNotFound", func(t *testing.T) {
		project := projectWithBuildSettings(t, serialized.Object{})

		_, err := project.TargetCodeSignEntitlements("App", "Release")
		assert.ErrorIs(t, err, ErrEntitlementsNotFound)
	})

	t.Run("entitlements named but unreadable is NOT ErrEntitlementsNotFound", func(t *testing.T) {
		project := projectWithBuildSettings(t, serialized.Object{
			"CODE_SIGN_ENTITLEMENTS": "App/App.entitlements",
		})

		_, err := project.TargetCodeSignEntitlements("App", "Release")
		require.Error(t, err)
		assert.NotErrorIs(t, err, ErrEntitlementsNotFound,
			"a missing-but-declared entitlements file is a real failure and must not be swallowed")
	})

	t.Run("malformed entitlements is a real error", func(t *testing.T) {
		project := projectWithBuildSettings(t, serialized.Object{
			"CODE_SIGN_ENTITLEMENTS": "App/App.entitlements",
		})
		writePlist(t, filepath.Join(filepath.Dir(project.Path), "App", "App.entitlements"), "not a plist")

		_, err := project.TargetCodeSignEntitlements("App", "Release")
		require.Error(t, err)
		assert.NotErrorIs(t, err, ErrEntitlementsNotFound)
	})
}

func TestXcodeProj_TargetInfoplistPath(t *testing.T) {
	t.Run("relative paths resolve against the project directory", func(t *testing.T) {
		project := projectWithBuildSettings(t, serialized.Object{
			"INFOPLIST_FILE": "App/Info.plist",
		})

		pth, err := project.TargetInfoplistPath("App", "Release")
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(filepath.Dir(project.Path), "App", "Info.plist"), pth)
	})

	t.Run("absolute paths are returned unchanged", func(t *testing.T) {
		project := projectWithBuildSettings(t, serialized.Object{
			"INFOPLIST_FILE": "/elsewhere/Info.plist",
		})

		pth, err := project.TargetInfoplistPath("App", "Release")
		require.NoError(t, err)
		assert.Equal(t, "/elsewhere/Info.plist", pth)
	})

	t.Run("absent setting", func(t *testing.T) {
		project := projectWithBuildSettings(t, serialized.Object{})

		_, err := project.TargetInfoplistPath("App", "Release")
		require.ErrorContains(t, err, "INFOPLIST_FILE")
	})
}

func TestResolveBundleID(t *testing.T) {
	tests := []struct {
		name          string
		bundleID      string
		buildSettings serialized.Object
		want          string
		wantErr       bool
	}{
		{
			name:          "no references",
			bundleID:      "io.bitrise.App",
			buildSettings: serialized.Object{},
			want:          "io.bitrise.App",
		},
		{
			name:          "braced and bare references together",
			bundleID:      `prefix.${PRODUCT_NAME}.$VERSION`,
			buildSettings: serialized.Object{"PRODUCT_NAME": "ios-simple-objc", "VERSION": "beta"},
			want:          "prefix.ios-simple-objc.beta",
		},
		{
			name:          "braced reference inside literal braces",
			bundleID:      `prefix.{text.${PRODUCT_NAME}.text}`,
			buildSettings: serialized.Object{"PRODUCT_NAME": "ios-simple-objc"},
			want:          "prefix.{text.ios-simple-objc.text}",
		},
		{
			name:          "parenthesised reference with a modifier",
			bundleID:      `auto_provision.$(PRODUCT_NAME:rfc1034identifier)`,
			buildSettings: serialized.Object{"PRODUCT_NAME": "ios-simple-objc"},
			want:          "auto_provision.ios-simple-objc",
		},
		{
			name:          "bare reference",
			bundleID:      `auto_provision.$PRODUCT_NAME`,
			buildSettings: serialized.Object{"PRODUCT_NAME": "ios-simple-objc"},
			want:          "auto_provision.ios-simple-objc",
		},
		{
			name:          "bare reference running into a suffix",
			bundleID:      `auto_provision.$PRODUCT_NAMEsuffix`,
			buildSettings: serialized.Object{"PRODUCT_NAME": "ios-simple-objc"},
			want:          "auto_provision.ios-simple-objcsuffix",
		},
		{
			name:          "bare reference repeated with a suffix",
			bundleID:      `auto_provision.$PRODUCT_NAMEsuffix$PRODUCT_NAME`,
			buildSettings: serialized.Object{"PRODUCT_NAME": "ios-simple-objc"},
			want:          "auto_provision.ios-simple-objcsuffixios-simple-objc",
		},
		{
			name:          "two adjacent bare references",
			bundleID:      `auto_provision.$PRODUCT_NAME$VERSION`,
			buildSettings: serialized.Object{"PRODUCT_NAME": "ios-simple-objc", "VERSION": "beta"},
			want:          "auto_provision.ios-simple-objcbeta",
		},
		{
			name:          "nested reference is resolved repeatedly",
			bundleID:      `io.bitrise.$(OUTER)`,
			buildSettings: serialized.Object{"OUTER": "$(INNER)", "INNER": "App"},
			want:          "io.bitrise.App",
		},
		{
			name:          "unknown braced reference",
			bundleID:      `io.bitrise.$(NOPE)`,
			buildSettings: serialized.Object{},
			wantErr:       true,
		},
		{
			name:          "reference cycle",
			bundleID:      `io.bitrise.$(A)`,
			buildSettings: serialized.Object{"A": "$(B)", "B": "$(A)"},
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveBundleID(tt.bundleID, tt.buildSettings)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsRelativePath(t *testing.T) {
	tests := map[string]bool{
		"App/Info.plist":            true,
		"./App/Info.plist":          true,
		"/abs/App/Info.plist":       false,
		"$(SRCROOT)/App/Info.plist": false,
		"$SRCROOT/App/Info.plist":   false,
		"":                          true,
	}

	for pth, want := range tests {
		assert.Equal(t, want, isRelativePath(pth), "isRelativePath(%q)", pth)
	}
}

func TestXcodeProj_TargetInfoplistPath_unexpandedReferenceIsLeftAlone(t *testing.T) {
	project := projectWithBuildSettings(t, serialized.Object{
		"INFOPLIST_FILE": "$(SRCROOT)/App/Info.plist",
	})

	pth, err := project.TargetInfoplistPath("App", "Release")
	require.NoError(t, err)
	assert.Equal(t, "$(SRCROOT)/App/Info.plist", pth)
}
