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

// A command reports what its policy refuses and what became of each option next to its
// own, and never repeats the parser's findings; every option passes through. The rows are
// production xcodebuild_options seen on xcode-archive; the message texts are pinned in
// TestDiagnosticMessages.
func TestCommand_Diagnostics(t *testing.T) {
	tests := []struct {
		name      string
		params    ArchiveParams
		wantArgs  []string
		wantKinds []DiagnosticKind
	}{
		{
			name:      "actions in the options",
			params:    ArchiveParams{AdditionalOptions: additional("clean", "archive")},
			wantArgs:  []string{"archive", "-project", "App.xcodeproj", "clean", "archive"},
			wantKinds: []DiagnosticKind{ActionInOptions, ActionInOptions},
		},
		{
			name:      "test-only flags",
			params:    ArchiveParams{AdditionalOptions: additional("-test-iterations", "2", "-retry-tests-on-failure")},
			wantArgs:  []string{"archive", "-project", "App.xcodeproj", "-test-iterations", "2", "-retry-tests-on-failure"},
			wantKinds: []DiagnosticKind{RejectedOption, RejectedOption},
		},
		{
			name:     "the parser's findings are not repeated",
			params:   ArchiveParams{AdditionalOptions: additional("-skipMacroValidation", "-destination", "-destination generic/platform=iOS", "-ENABLE_BITCODE=NO")},
			wantArgs: []string{"archive", "-project", "App.xcodeproj", "-skipMacroValidation", "-destination", "-destination generic/platform=iOS", "-ENABLE_BITCODE=NO"},
		},
		{
			name:      "repeated value option",
			params:    ArchiveParams{XCConfigPath: "/tmp/temp.xcconfig", AdditionalOptions: additional("-xcconfig", "mine.xcconfig")},
			wantArgs:  []string{"archive", "-project", "App.xcodeproj", "-xcconfig", "/tmp/temp.xcconfig", "-xcconfig", "mine.xcconfig"},
			wantKinds: []DiagnosticKind{RepeatedOption},
		},
		{
			// -allowProvisioningUpdates only works with the API key flags the step adds.
			name:      "lone -allowProvisioningUpdates points at the step input",
			params:    ArchiveParams{AdditionalOptions: additional("-allowProvisioningUpdates")},
			wantArgs:  []string{"archive", "-project", "App.xcodeproj", "-allowProvisioningUpdates"},
			wantKinds: []DiagnosticKind{PreferStepInput},
		},
		{
			name:     "a redundant switch and a replaced default are informational",
			params:   ArchiveParams{Destination: "generic/platform=iOS", Authentication: &testAuth, AdditionalOptions: additional("-allowProvisioningUpdates", "-destination", "id=SIM")},
			wantArgs: []string{"archive", "-project", "App.xcodeproj", "-authenticationKeyPath", "/key/path", "-authenticationKeyID", "keyID", "-authenticationKeyIssuerID", "issuerID", "-allowProvisioningUpdates", "-destination", "id=SIM"},
			// reported in derived-flag order
			wantKinds: []DiagnosticKind{Override, RedundantOption},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := tt.params
			params.ProjectPath = "App.xcodeproj"

			cmd, err := Archive(params)
			require.NoError(t, err)
			require.Equal(t, tt.wantArgs, cmd.Args())
			require.Equal(t, tt.wantKinds, kinds(cmd.Diagnostics()))
		})
	}
}

// Every message is pinned here, one per finding. Each row is the log a step prints for the
// user options: the parser's findings, then the command's, with the user options merged
// over the derived ones the way assemble does.
func TestDiagnosticMessages(t *testing.T) {
	tests := []struct {
		name    string
		user    []string
		derived Options
		policy  actionPolicy
		want    []string
	}{
		{
			name:   "parser",
			user:   []string{"", "-destination 'platform=iOS Simulator,name=iPhone 15'", "-sdk macosx", "-=x", "-only-testing:", "Distribution", "-ENABLE_BITCODE=NO", "-destination"},
			policy: archivePolicy,
			want: []string{
				`"" is empty. xcodebuild treats it as an unknown build action. Remove it.`,
				`"-destination 'platform=iOS Simulator,name=iPhone 15'" is a flag quoted together with its value. xcodebuild reads it as a user default and ignores it, so it has no effect. Use -destination 'platform=iOS Simulator,name=iPhone 15' to apply it, or remove it.`,
				`"-sdk macosx" is a flag quoted together with its value. xcodebuild refuses it. Use -sdk macosx instead.`,
				`"-=x" is not a valid flag. xcodebuild refuses it. Remove it.`,
				`"-only-testing:" has no value after the colon. xcodebuild refuses it. Add the value, or remove it.`,
				`"Distribution" is not a flag, a NAME=value build setting or a build action. xcodebuild treats it as an unknown build action. Quote a value with spaces, for example CODE_SIGN_IDENTITY="Apple Distribution", or remove it.`,
				`"-ENABLE_BITCODE=NO" looks like the build setting ENABLE_BITCODE=NO with a leading dash. xcodebuild reads it as a user default and ignores it, so it has no effect. Use ENABLE_BITCODE=NO to apply it, or remove it.`,
				`"-destination" has no value. xcodebuild refuses it. Add one, or remove it.`,
			},
		},
		{
			name:   "policy and step input",
			user:   []string{"-exportArchive", "-test-iterations", "2", "-testPlan", "Full", "clean", "-allowProvisioningUpdates"},
			policy: archivePolicy,
			want: []string{
				`"-exportArchive" does not work with archive. It switches xcodebuild to another task instead. Remove it.`,
				`"-test-iterations 2" does not work with archive. It applies to test actions only, and xcodebuild ignores it here. Remove it.`,
				`"-testPlan Full" does not work with archive. It applies to test actions only, and xcodebuild refuses it here. Remove it.`,
				`"clean" is a build action. The Step sets the actions itself. Remove it.`,
				`"-allowProvisioningUpdates" cannot update provisioning profiles on its own. xcodebuild needs the App Store Connect API key flags, which the Step adds when its automatic code signing is enabled. Remove it and enable automatic code signing instead. See https://docs.bitrise.io/en/code-signing/ios-code-signing/managing-ios-code-signing-files---automatic-provisioning.html`,
			},
		},
		{
			name:   "-allowProvisioningUpdates with the user's own API key flags",
			user:   []string{"-allowProvisioningUpdates", "-authenticationKeyPath", "/key.p8", "-authenticationKeyID", "ID", "-authenticationKeyIssuerID", "ISSUER"},
			policy: archivePolicy,
		},
		{
			name: "merge",
			user: []string{"-destination", "generic/platform=tvOS", "-allowProvisioningUpdates", "-xcconfig", "mine.xcconfig", "-collect-test-diagnostics=on-failure"},
			derived: Options{
				{Kind: ValueOption, Name: "-destination", Value: "generic/platform=iOS"},
				{Kind: Switch, Name: "-allowProvisioningUpdates"},
				{Kind: ValueOption, Name: "-xcconfig", Value: "/tmp/temp.xcconfig"},
				{Kind: ValueOption, Name: "-collect-test-diagnostics", Value: "never"},
			},
			policy: actionPolicy{name: "archive", defaults: []string{"-destination"}},
			want: []string{
				`"-destination generic/platform=tvOS" replaces the Step's default "-destination generic/platform=iOS".`,
				`"-allowProvisioningUpdates" is already set by the Step. Remove it.`,
				`-xcconfig is managed by the Step. "-xcconfig mine.xcconfig" conflicts with its "-xcconfig /tmp/temp.xcconfig": xcodebuild accepts the flag only once. Remove it.`,
				`"-collect-test-diagnostics=on-failure" is written with "=". xcodebuild reads it as a user default and ignores it, and uses the Step's "-collect-test-diagnostics never" instead. Use -collect-test-diagnostics on-failure to apply it, or remove it.`,
			},
		},
		{
			name:    "a switch the user repeats on top of the Step's is redundant, not an override",
			user:    []string{"-allowProvisioningUpdates", "-allowProvisioningUpdates"},
			derived: Options{{Kind: Switch, Name: "-allowProvisioningUpdates"}},
			policy:  archivePolicy,
			want: []string{
				`"-allowProvisioningUpdates" is already set by the Step. Remove it.`,
			},
		},
		{
			name:    "a value given to the Step's switch, in each form",
			user:    []string{"-allowProvisioningUpdates", "YES", "-allowProvisioningUpdates:YES", "-allowProvisioningUpdates=YES"},
			derived: Options{{Kind: Switch, Name: "-allowProvisioningUpdates"}},
			policy:  archivePolicy,
			want: []string{
				`"-allowProvisioningUpdates YES" gives a value to -allowProvisioningUpdates, which takes none, so xcodebuild reads "YES" as a build action and fails. The Step already sets -allowProvisioningUpdates. Remove it.`,
				`"-allowProvisioningUpdates:YES" gives a value to -allowProvisioningUpdates, which takes none, so xcodebuild refuses it as an invalid option. The Step already sets -allowProvisioningUpdates. Remove it.`,
				`"-allowProvisioningUpdates=YES" gives a value to -allowProvisioningUpdates, which takes none, so xcodebuild reads it as a user default and ignores it. The Step already sets -allowProvisioningUpdates. Remove it.`,
			},
		},
		{
			name:   "a value given to the user's own switch",
			user:   []string{"-quiet", "-quiet:YES"},
			policy: archivePolicy,
			want: []string{
				`"-quiet:YES" gives a value to -quiet, which takes none, so xcodebuild refuses it as an invalid option. -quiet is already set. Remove it.`,
			},
		},
		{
			name: "a value option written with = is checked against what the build uses",
			user: []string{"-destination", "generic/platform=tvOS", "-destination=generic/platform=watchOS", "-xcconfig=mine.xcconfig", "-scheme=App", "-quiet", "-quiet=YES"},
			derived: Options{
				{Kind: ValueOption, Name: "-destination", Value: "generic/platform=iOS"},
				{Kind: ValueOption, Name: "-xcconfig", Value: "/tmp/temp.xcconfig"},
			},
			policy: actionPolicy{name: "archive", defaults: []string{"-destination"}},
			want: []string{
				`"-destination generic/platform=tvOS" replaces the Step's default "-destination generic/platform=iOS".`,
				`"-destination=generic/platform=watchOS" is written with "=". xcodebuild reads it as a user default and ignores it, and uses your "-destination generic/platform=tvOS" instead. Change that one to -destination generic/platform=watchOS to apply it, or remove it.`,
				`"-xcconfig=mine.xcconfig" is written with "=". xcodebuild reads it as a user default and ignores it, and uses the Step's "-xcconfig /tmp/temp.xcconfig" instead. Use -xcconfig mine.xcconfig to apply it, or remove it.`,
				`"-scheme=App" is written with "=". xcodebuild reads it as a user default and ignores it, so it has no effect. Use -scheme App to apply it, or remove it.`,
				`"-quiet=YES" gives a value to -quiet, which takes none, so xcodebuild reads it as a user default and ignores it. -quiet is already set. Remove it.`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := additional(tt.user...)
			got := append(user.Diagnostics(), lint(merge(tt.derived, user, tt.policy), tt.policy)...)
			require.Equal(t, tt.want, messages(got))
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
