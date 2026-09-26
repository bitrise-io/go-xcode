package xcodecommand

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Every message string is pinned here, one per finding; the rules that produce the
// findings are covered in options_test.go, policy_test.go and merge_test.go.

func TestOptions_Diagnostics_messages(t *testing.T) {
	opts := ParseAdditionalOptions([]string{"", "-destination 'platform=iOS Simulator,name=iPhone 15'", "-sdk macosx", "-=x", "-only-testing:", "Distribution", "-ENABLE_BITCODE=NO", "-destination"})
	require.Equal(t, []string{
		`"" is empty. xcodebuild treats it as an unknown build action. Remove it.`,
		`"-destination 'platform=iOS Simulator,name=iPhone 15'" is a flag quoted together with its value. xcodebuild reads it as a user default and ignores it. Use -destination "platform=iOS Simulator,name=iPhone 15" instead.`,
		`"-sdk macosx" is a flag quoted together with its value. xcodebuild refuses it. Use -sdk macosx instead.`,
		`"-=x" is not a valid flag. xcodebuild refuses it. Remove it.`,
		`"-only-testing:" has no value after the colon. Add the value or remove the flag.`,
		`"Distribution" is not a flag, a NAME=value build setting or a build action. xcodebuild treats it as an unknown build action. Quote a value with spaces, for example CODE_SIGN_IDENTITY="Apple Distribution", or remove it.`,
		`"-ENABLE_BITCODE=NO" looks like the build setting ENABLE_BITCODE=NO with a leading dash. xcodebuild reads it as a user default and the setting never applies. Use ENABLE_BITCODE=NO instead.`,
		`"-destination" has no value. Add one or remove the flag.`,
	}, messages(opts.Diagnostics()))
}

func TestLintPolicy_messages(t *testing.T) {
	opts := ParseAdditionalOptions([]string{"-exportArchive", "-test-iterations", "2", "clean"})
	require.Equal(t, []string{
		`"-exportArchive" switches xcodebuild into another mode and is not valid for archive. Remove it.`,
		`"-test-iterations 2" applies to test actions only and is not valid for archive. Remove it.`,
		`"clean" is a build action. The archive command sets its own actions. Remove it.`,
	}, messages(lintPolicy(opts, archivePolicy)))
}

func TestLint_mergeMessages(t *testing.T) {
	derived := Options{
		{Kind: ValueOption, Name: "-destination", Value: "generic/platform=iOS"},
		{Kind: Switch, Name: "-allowProvisioningUpdates"},
		{Kind: ValueOption, Name: "-xcconfig", Value: "/tmp/temp.xcconfig"},
		{Kind: ValueOption, Name: "-collect-test-diagnostics", Value: "never"},
	}
	user := Options{
		{Kind: ValueOption, Name: "-destination", Value: "generic/platform=tvOS"},
		{Kind: Switch, Name: "-allowProvisioningUpdates"},
		{Kind: ValueOption, Name: "-xcconfig", Value: "mine.xcconfig"},
		{Kind: UserDefault, Name: "-collect-test-diagnostics", Value: "on-failure"},
	}
	policy := actionPolicy{name: "archive", defaults: []string{"-destination"}}
	_, collisions := merge(derived, user, policy)

	require.Equal(t, []string{
		`"-collect-test-diagnostics=on-failure" is written with "=". xcodebuild reads it as a user default and ignores it, so the Step's "-collect-test-diagnostics never" stays. Use -collect-test-diagnostics on-failure instead.`,
		`"-destination generic/platform=tvOS" replaces the Step's default "-destination generic/platform=iOS".`,
		`"-allowProvisioningUpdates" is already set by the Step. Remove it.`,
		`"-xcconfig mine.xcconfig" repeats "-xcconfig /tmp/temp.xcconfig", which the Step sets. xcodebuild refuses a repeated option. Remove it, or change the Step input instead.`,
	}, messages(lint(user, derived, collisions, policy)))
}

func TestLintStepInputs_message(t *testing.T) {
	user := Options{{Kind: Switch, Name: "-allowProvisioningUpdates"}}
	require.Equal(t, []string{
		`"-allowProvisioningUpdates" cannot update provisioning profiles on its own. xcodebuild needs the App Store Connect API key flags, which the Step adds when its automatic code signing is enabled. Remove it and enable automatic code signing instead.`,
	}, messages(lintStepInputs(nil, user)))
	require.Empty(t, lintStepInputs(user, user), "derived by the step: the merge reports it as redundant instead")
}

func messages(diagnostics []Diagnostic) []string {
	var out []string
	for _, d := range diagnostics {
		out = append(out, d.Message)
	}
	return out
}
