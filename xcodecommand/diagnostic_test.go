package xcodecommand

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// The parser's findings, one per argument xcodebuild accepts and does something else with,
// each naming the option it is about; every option still passes through. The message texts
// are pinned in TestDiagnosticMessages.
func TestOptions_Diagnostics(t *testing.T) {
	opts := additional("-skipMacroValidation", "-destination", "-destination generic/platform=iOS", "-ENABLE_BITCODE=NO")

	diagnostics := opts.Diagnostics()

	require.Equal(t, []DiagnosticKind{SuspiciousUserDefault, SuspiciousUserDefault}, kinds(diagnostics))
	require.Equal(t, []string{"-destination generic/platform=iOS", "-ENABLE_BITCODE=NO"}, options(diagnostics), "the lone -destination, which xcodebuild refuses, gets no line")
	require.Equal(t, []string{"-skipMacroValidation", "-destination", "-destination generic/platform=iOS", "-ENABLE_BITCODE=NO"}, opts.Args())
}

// A finding is informational when the command line does what the user meant; the parser's
// never are. A step logs informational findings at info level, the rest as warnings.
func TestDiagnostic_Informational(t *testing.T) {
	for _, d := range additional("-destination generic/platform=iOS", "-ENABLE_BITCODE=NO", "-derivedDataPath", "-quiet").Diagnostics() {
		require.False(t, d.Informational, d.Message)
	}
	require.True(t, report(Override, Option{}, "").Informational)
	require.True(t, report(RedundantOption, Option{}, "").Informational)
	require.False(t, report(PreferStepInput, Option{}, "").Informational)
}

// Every message is pinned here, one per finding.
func TestDiagnosticMessages(t *testing.T) {
	tests := []struct {
		name string
		user []string
		want []string
	}{
		{
			// "", "-sdk macosx", "-=x", "-only-testing:", "Distribution" and the lone -destination
			// get no line: xcodebuild refuses them with its own message.
			name: "parser",
			user: []string{"", "-destination 'platform=iOS Simulator,name=iPhone 15'", "-sdk macosx", "-=x", "-only-testing:", "Distribution", "-ENABLE_BITCODE=NO", "-destination", "-derivedDataPath", "-quiet"},
			want: []string{
				`"-destination 'platform=iOS Simulator,name=iPhone 15'" is a flag quoted together with its value. xcodebuild reads it as a user default and ignores it, so it has no effect. Use -destination 'platform=iOS Simulator,name=iPhone 15' to apply it, or remove it.`,
				`"-ENABLE_BITCODE=NO" looks like the build setting ENABLE_BITCODE=NO with a leading dash. xcodebuild reads it as a user default and ignores it, so it has no effect. Use ENABLE_BITCODE=NO to apply it, or remove it.`,
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
