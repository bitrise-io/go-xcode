package xcodecommand

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseAdditionalOptions(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		want  Options
		diags []DiagnosticKind
	}{
		{name: "nothing"},
		{
			name: "switch",
			args: []string{"-quiet"},
			want: Options{{Kind: Switch, Name: "-quiet"}},
		},
		{
			name: "flag followed by a bare word takes it as its value",
			args: []string{"-packageAuthorizationProvider", "netrc"},
			want: Options{{Kind: ValueOption, Name: "-packageAuthorizationProvider", Value: "netrc"}},
		},
		{
			name: "flag followed by another flag is a switch",
			args: []string{"-skipPackageSignatureValidation", "-quiet"},
			want: Options{{Kind: Switch, Name: "-skipPackageSignatureValidation"}, {Kind: Switch, Name: "-quiet"}},
		},
		{
			name: "flag followed by a build setting is a switch",
			args: []string{"-verbose", "ARCHS=x86_64"},
			want: Options{{Kind: Switch, Name: "-verbose"}, {Kind: BuildSetting, Name: "ARCHS", Value: "x86_64"}},
		},
		{
			name: "flag at the end is a switch",
			args: []string{"-verbose"},
			want: Options{{Kind: Switch, Name: "-verbose"}},
		},
		{
			name: "-destination takes a value that looks like a build setting",
			args: []string{"-destination", "platform=iOS Simulator,name=iPhone 15"},
			want: Options{{Kind: ValueOption, Name: "-destination", Value: "platform=iOS Simulator,name=iPhone 15"}},
		},
		{
			name: "-scheme takes a value that is an action name",
			args: []string{"-scheme", "test"},
			want: Options{{Kind: ValueOption, Name: "-scheme", Value: "test"}},
		},
		{
			name: "colon option",
			args: []string{"-only-testing:AppTests/LoginTests"},
			want: Options{{Kind: ColonOption, Name: "-only-testing", Value: "AppTests/LoginTests"}},
		},
		{
			name: "colon option whose value contains an equals sign",
			args: []string{"-only-testing:AppTests/test=1"},
			want: Options{{Kind: ColonOption, Name: "-only-testing", Value: "AppTests/test=1"}},
		},
		{
			name: "user default",
			args: []string{"-UseModernBuildSystem=NO"},
			want: Options{{Kind: UserDefault, Name: "-UseModernBuildSystem", Value: "NO"}},
		},
		{
			name: "user default whose value contains a colon",
			args: []string{"-IDECustomDerivedDataLocation=/a:b"},
			want: Options{{Kind: UserDefault, Name: "-IDECustomDerivedDataLocation", Value: "/a:b"}},
		},
		{
			name:  "user default that looks like a build setting is reported and kept",
			args:  []string{"-ENABLE_BITCODE=NO"},
			want:  Options{{Kind: UserDefault, Name: "-ENABLE_BITCODE", Value: "NO"}},
			diags: []DiagnosticKind{SuspiciousUserDefault},
		},
		{
			name: "build settings, any case, split at the first equals sign",
			args: []string{"CODE_SIGNING_ALLOWED=NO", "my_setting=1", "OTHER_LDFLAGS=", "OTHER_SWIFT_FLAGS=-D A=B"},
			want: Options{
				{Kind: BuildSetting, Name: "CODE_SIGNING_ALLOWED", Value: "NO"},
				{Kind: BuildSetting, Name: "my_setting", Value: "1"},
				{Kind: BuildSetting, Name: "OTHER_LDFLAGS", Value: ""},
				{Kind: BuildSetting, Name: "OTHER_SWIFT_FLAGS", Value: "-D A=B"},
			},
		},
		{
			name: "build action",
			args: []string{"clean"},
			want: Options{{Kind: Action, Name: "clean"}},
		},
		{
			name: "steps-xcode-archive style mix",
			args: []string{"-destination", "generic/platform=iOS", "-skipPackagePluginValidation", "COMPILER_INDEX_STORE_ENABLE=NO"},
			want: Options{
				{Kind: ValueOption, Name: "-destination", Value: "generic/platform=iOS"},
				{Kind: Switch, Name: "-skipPackagePluginValidation"},
				{Kind: BuildSetting, Name: "COMPILER_INDEX_STORE_ENABLE", Value: "NO"},
			},
		},
		{
			name:  "bare word is kept verbatim and reported",
			args:  []string{"-sdk", "iphoneos", "PRODUCT_NAME=My", "App"},
			want:  Options{{Kind: ValueOption, Name: "-sdk", Value: "iphoneos"}, {Kind: BuildSetting, Name: "PRODUCT_NAME", Value: "My"}, {Kind: Unknown, Name: "App"}},
			diags: []DiagnosticKind{MalformedOption},
		},
		{
			name:  "empty argument is kept verbatim and reported",
			args:  []string{""},
			want:  Options{{Kind: Unknown, Name: ""}},
			diags: []DiagnosticKind{MalformedOption},
		},
		{
			name:  "free-form value flag without a value is reported",
			args:  []string{"-destination"},
			want:  Options{{Kind: Unknown, Name: "-destination"}},
			diags: []DiagnosticKind{MalformedOption},
		},
		{
			name:  "flag and value quoted together are reported",
			args:  []string{"-destination generic/platform=iOS"},
			want:  Options{{Kind: Unknown, Name: "-destination generic/platform=iOS"}},
			diags: []DiagnosticKind{MalformedOption},
		},
		{
			name:  "lone dash, empty user default name and empty colon value are reported",
			args:  []string{"-", "-=x", "-only-testing:"},
			want:  Options{{Kind: Unknown, Name: "-"}, {Kind: Unknown, Name: "-=x"}, {Kind: Unknown, Name: "-only-testing:"}},
			diags: []DiagnosticKind{MalformedOption, MalformedOption, MalformedOption},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, diags := ParseAdditionalOptions(tt.args)
			require.Equal(t, tt.want, got)
			require.Equal(t, tt.args, got.Args(), "parsing must round-trip")

			var kinds []DiagnosticKind
			for _, d := range diags {
				kinds = append(kinds, d.Kind)
			}
			require.Equal(t, tt.diags, kinds)
		})
	}
}

// TestOptions_Filter reproduces xcode-archive's filterSPMAdditionalOptions on typed
// options: keep SPM flags and build settings for -showBuildSettings, drop the rest.
func TestOptions_Filter(t *testing.T) {
	spmFlags := []string{"-skipPackagePluginValidation", "-skipMacroValidation", "-skipPackageUpdates",
		"-disableAutomaticPackageResolution", "-onlyUsePackageVersionsFromResolvedFile", "-clonedSourcePackagesDirPath"}
	opts, diags := ParseAdditionalOptions([]string{
		"-destination", "generic/platform=iOS",
		"-skipPackagePluginValidation",
		"-clonedSourcePackagesDirPath", "/tmp/spm",
		"BUNDLE_IDENTIFIER=io.bitrise.sample",
		"-quiet",
	})
	require.Empty(t, diags)

	kept := opts.Filter(func(o Option) bool { return o.Kind == BuildSetting || slices.Contains(spmFlags, o.Name) })

	require.Equal(t, []string{
		"-skipPackagePluginValidation",
		"-clonedSourcePackagesDirPath", "/tmp/spm",
		"BUNDLE_IDENTIFIER=io.bitrise.sample",
	}, kept.Args(), "the destination value generic/platform=iOS is not mistaken for a build setting")
}

func TestParseAdditionalOptions_pathValueNamedLikeAnAction(t *testing.T) {
	got, diags := ParseAdditionalOptions([]string{"-derivedDataPath", "build", "-clonedSourcePackagesDirPath", "test"})
	require.Empty(t, diags)
	require.Equal(t, Options{
		{Kind: ValueOption, Name: "-derivedDataPath", Value: "build"},
		{Kind: ValueOption, Name: "-clonedSourcePackagesDirPath", Value: "test"},
	}, got)
}

func TestOption_Key(t *testing.T) {
	require.Equal(t, "-quiet", Option{Kind: Switch, Name: "-quiet"}.Key())
	require.Equal(t, "-sdk", Option{Kind: ValueOption, Name: "-sdk", Value: "iphoneos"}.Key())
	require.Equal(t, "-sdk=", Option{Kind: UserDefault, Name: "-sdk", Value: "iphoneos"}.Key(), "a user default never collides with the flag of the same name")
	require.Equal(t, "ARCHS", Option{Kind: BuildSetting, Name: "ARCHS", Value: "arm64"}.Key())
}
