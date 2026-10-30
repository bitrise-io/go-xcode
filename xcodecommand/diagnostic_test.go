package xcodecommand

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// The parser's findings, one per argument it could not read or that xcodebuild would
// ignore, each naming the option it is about; every option still passes through. The
// message texts are pinned in TestDiagnosticMessages.
func TestOptions_Diagnostics(t *testing.T) {
	opts := additional("-skipMacroValidation", "-destination", "-destination generic/platform=iOS", "-ENABLE_BITCODE=NO")

	diagnostics := opts.Diagnostics()

	require.Equal(t, []DiagnosticKind{MalformedOption, SuspiciousUserDefault, SuspiciousUserDefault}, kinds(diagnostics))
	require.Equal(t, []string{"-destination", "-destination generic/platform=iOS", "-ENABLE_BITCODE=NO"}, options(diagnostics))
	require.Equal(t, []string{"-skipMacroValidation", "-destination", "-destination generic/platform=iOS", "-ENABLE_BITCODE=NO"}, opts.Args())
}

func TestDiagnosticKind_Informational(t *testing.T) {
	for kind, want := range map[DiagnosticKind]bool{
		MalformedOption: false, RejectedOption: false, ActionInOptions: false, SuspiciousUserDefault: false,
		RepeatedOption: false, PreferStepInput: false, Override: true, RedundantOption: true,
	} {
		require.Equal(t, want, kind.Informational(), "kind %d", kind)
	}
}

// Every message is pinned here, one per finding.
func TestDiagnosticMessages(t *testing.T) {
	tests := []struct {
		name string
		user []string
		want []string
	}{
		{
			name: "parser",
			user: []string{"", "-destination 'platform=iOS Simulator,name=iPhone 15'", "-sdk macosx", "-=x", "-only-testing:", "Distribution", "-ENABLE_BITCODE=NO", "-destination", "-derivedDataPath", "-quiet"},
			want: []string{
				`"" is empty. xcodebuild treats it as an unknown build action. Remove it.`,
				`"-destination 'platform=iOS Simulator,name=iPhone 15'" is a flag quoted together with its value. xcodebuild reads it as a user default and ignores it, so it has no effect. Use -destination 'platform=iOS Simulator,name=iPhone 15' to apply it, or remove it.`,
				`"-sdk macosx" is a flag quoted together with its value. xcodebuild refuses it. Use -sdk macosx instead.`,
				`"-=x" is not a valid flag. xcodebuild refuses it. Remove it.`,
				`"-only-testing:" has no value after the colon. xcodebuild refuses it. Add the value, or remove it.`,
				`"Distribution" is not a flag, a NAME=value build setting or a build action. xcodebuild treats it as an unknown build action. Quote a value with spaces, for example CODE_SIGN_IDENTITY="Apple Distribution", or remove it.`,
				`"-ENABLE_BITCODE=NO" looks like the build setting ENABLE_BITCODE=NO with a leading dash. xcodebuild reads it as a user default and ignores it, so it has no effect. Use ENABLE_BITCODE=NO to apply it, or remove it.`,
				`"-destination" has no value. xcodebuild refuses it. Add one, or remove it.`,
				`"-derivedDataPath" has no value, so xcodebuild takes "-quiet" as its value and that flag does not apply. Add a value, or remove it.`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, messages(additional(tt.user...).Diagnostics()))
		})
	}
}

func kinds(diagnostics []Diagnostic) []DiagnosticKind {
	var out []DiagnosticKind
	for _, d := range diagnostics {
		out = append(out, d.Kind)
	}
	return out
}

// options renders the option each diagnostic is about.
func options(diagnostics []Diagnostic) []string {
	var out []string
	for _, d := range diagnostics {
		out = append(out, d.Option.String())
	}
	return out
}

func messages(diagnostics []Diagnostic) []string {
	var out []string
	for _, d := range diagnostics {
		out = append(out, d.Message)
	}
	return out
}

// The corrected forms in the messages are pasted back into xcodebuild_options, which
// ParseOptions reads with POSIX shell rules; each rendering must read back as the same
// value.
func TestShellQuoted(t *testing.T) {
	for value, want := range map[string]string{
		"":                                      `''`,
		"iphoneos":                              "iphoneos",
		"generic/platform=iOS":                  "generic/platform=iOS",
		"platform=iOS Simulator,name=iPhone 15": `'platform=iOS Simulator,name=iPhone 15'`,
		"Apple Development: Bot":                `'Apple Development: Bot'`,
		`say "hi"`:                              `'say "hi"'`,
		"it's here":                             `'it'\''s here'`,
		`a\b`:                                   `a\\b`,
		"$HOME/dd":                              `\$HOME/dd`,
		"don't $shout":                          `'don'\''t $shout'`,
	} {
		require.Equal(t, want, shellQuoted(value), value)
		opts, err := ParseOptions("-flag " + shellQuoted(value))
		require.NoError(t, err)
		require.Equal(t, []string{"-flag", value}, opts.Args(), "must read back as the same value")
	}
	require.Equal(t, `-destination 'platform=iOS Simulator,name=iPhone 15'`, unquoteFlag("-destination 'platform=iOS Simulator,name=iPhone 15'"))
	require.Equal(t, "-sdk macosx", unquoteFlag("-sdk  macosx"))
	require.Equal(t, "-sdk macosx", unquoteFlag("-sdk\tmacosx"))
	require.Equal(t, "-sdk macosx", unquoteFlag("-sdk macosx"))
}
