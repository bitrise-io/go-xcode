package xcodecommand

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Warn passes every option through and reports; Fail refuses the first non-informational
// finding before xcodebuild runs. The rows are production xcodebuild_options seen on
// xcode-archive; the message texts are pinned in TestDiagnosticMessages.
func TestValidation(t *testing.T) {
	tests := []struct {
		name      string
		params    ArchiveParams
		wantArgs  []string
		wantKinds []DiagnosticKind
		failPass  bool // only informational findings: Fail still builds the command
	}{
		{
			name:      "free-form value flag left without its value",
			params:    ArchiveParams{AdditionalOptions: []string{"-skipMacroValidation", "-destination"}},
			wantArgs:  []string{"archive", "-project", "App.xcodeproj", "-skipMacroValidation", "-destination"},
			wantKinds: []DiagnosticKind{MalformedOption},
		},
		{
			name:      "flag quoted together with its value, build setting with a dash",
			params:    ArchiveParams{AdditionalOptions: []string{"-destination generic/platform=iOS", "-ENABLE_BITCODE=NO"}},
			wantArgs:  []string{"archive", "-project", "App.xcodeproj", "-destination generic/platform=iOS", "-ENABLE_BITCODE=NO"},
			wantKinds: []DiagnosticKind{SuspiciousUserDefault, SuspiciousUserDefault},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := tt.params
			params.ProjectPath = "App.xcodeproj"

			cmd, err := Archive(params)
			require.NoError(t, err)
			require.Equal(t, tt.wantArgs, cmd.Args(), "Warn passes everything through")
			require.Equal(t, tt.wantKinds, kinds(cmd.Diagnostics()))

			params.Validation = Fail
			failed, err := Archive(params)
			if tt.failPass {
				require.NoError(t, err)
				require.Equal(t, cmd, failed)
				return
			}
			require.EqualError(t, err, "invalid additional option: "+cmd.Diagnostics()[0].Message)
		})
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
			user: []string{"", "-destination 'platform=iOS Simulator,name=iPhone 15'", "-sdk macosx", "-=x", "-only-testing:", "Distribution", "-ENABLE_BITCODE=NO", "-destination"},
			want: []string{
				`"" is empty. xcodebuild treats it as an unknown build action. Remove it.`,
				`"-destination 'platform=iOS Simulator,name=iPhone 15'" is a flag quoted together with its value. xcodebuild reads it as a user default and ignores it, so today's build runs without it. Remove it to keep that, or use -destination 'platform=iOS Simulator,name=iPhone 15' to apply it.`,
				`"-sdk macosx" is a flag quoted together with its value. xcodebuild refuses it. Use -sdk macosx instead.`,
				`"-=x" is not a valid flag. xcodebuild refuses it. Remove it.`,
				`"-only-testing:" has no value after the colon. Add the value or remove the flag.`,
				`"Distribution" is not a flag, a NAME=value build setting or a build action. xcodebuild treats it as an unknown build action. Quote a value with spaces, for example CODE_SIGN_IDENTITY="Apple Distribution", or remove it.`,
				`"-ENABLE_BITCODE=NO" looks like the build setting ENABLE_BITCODE=NO with a leading dash. xcodebuild reads it as a user default and ignores it, so today's build runs without it. Remove it to keep that, or use ENABLE_BITCODE=NO to apply it.`,
				`"-destination" has no value. Add one or remove the flag.`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, messages(lint(ParseAdditionalOptions(tt.user))))
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

func messages(diagnostics []Diagnostic) []string {
	var out []string
	for _, d := range diagnostics {
		out = append(out, d.Message)
	}
	return out
}

// The corrected forms in the messages are pasted back into xcodebuild_options, which
// SplitAdditionalOptions reads with POSIX shell rules; each rendering must read back as
// the same value.
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
		args, err := SplitAdditionalOptions("-flag " + shellQuoted(value))
		require.NoError(t, err)
		require.Equal(t, []string{"-flag", value}, args, "must read back as the same value")
	}
	require.Equal(t, `-destination 'platform=iOS Simulator,name=iPhone 15'`, unquoteFlag("-destination 'platform=iOS Simulator,name=iPhone 15'"))
	require.Equal(t, "-sdk macosx", unquoteFlag("-sdk  macosx"))
}
