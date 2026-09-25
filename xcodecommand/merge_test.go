package xcodecommand

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMerge(t *testing.T) {
	derived := Options{
		{Kind: Action, Name: "clean"},
		{Kind: Action, Name: "archive"},
		{Kind: ValueOption, Name: "-project", Value: "App.xcodeproj"},
		{Kind: ValueOption, Name: "-scheme", Value: "App"},
		{Kind: ValueOption, Name: "-configuration", Value: "Release"},
		{Kind: ValueOption, Name: "-destination", Value: "generic/platform=iOS"},
		{Kind: ValueOption, Name: "-xcconfig", Value: "/tmp/temp.xcconfig"},
		{Kind: Switch, Name: "-allowProvisioningUpdates"},
		{Kind: BuildSetting, Name: "CODE_SIGNING_ALLOWED", Value: "NO"},
	}
	policy := actionPolicy{name: "archive", defaults: []string{"-destination"}, appendable: []string{"-skip-testing"}}
	base := []string{
		"clean", "archive", "-project", "App.xcodeproj", "-scheme", "App",
		"-configuration", "Release", "-destination", "generic/platform=iOS",
		"-xcconfig", "/tmp/temp.xcconfig", "-allowProvisioningUpdates", "CODE_SIGNING_ALLOWED=NO",
	}
	without := func(flag string, width int) []string {
		var out []string
		for i := 0; i < len(base); i++ {
			if base[i] == flag {
				i += width - 1
				continue
			}
			out = append(out, base[i])
		}
		return out
	}

	tests := []struct {
		name        string
		user        Options
		want        []string
		diagnostics []Diagnostic
	}{
		{
			name: "no user options leaves the derived options untouched",
			want: base,
		},
		{
			name: "options with a new key are appended in the order given",
			user: Options{{Kind: Switch, Name: "-quiet"}, {Kind: ValueOption, Name: "-sdk", Value: "iphoneos"}},
			want: append(without("", 0), "-quiet", "-sdk", "iphoneos"),
		},
		{
			name:        "a repeated value option is passed through for xcodebuild to refuse, and reported",
			user:        Options{{Kind: ValueOption, Name: "-configuration", Value: "Debug"}},
			want:        append(without("", 0), "-configuration", "Debug"),
			diagnostics: []Diagnostic{{Kind: RepeatedOption, Message: `"-configuration Release" is set by archive and again as additional option [-configuration Debug]; xcodebuild refuses a repeated option`}},
		},
		{
			name:        "a repeated owned path is passed through the same way",
			user:        Options{{Kind: ValueOption, Name: "-xcconfig", Value: "mine.xcconfig"}},
			want:        append(without("", 0), "-xcconfig", "mine.xcconfig"),
			diagnostics: []Diagnostic{{Kind: RepeatedOption, Message: `"-xcconfig /tmp/temp.xcconfig" is set by archive and again as additional option [-xcconfig mine.xcconfig]; xcodebuild refuses a repeated option`}},
		},
		{
			name: "several user destinations replace the default as a group and all survive",
			user: Options{
				{Kind: ValueOption, Name: "-destination", Value: "platform=iOS Simulator,name=iPhone 15"},
				{Kind: ValueOption, Name: "-destination", Value: "platform=iOS Simulator,name=iPad Air"},
			},
			want: append(without("-destination", 2),
				"-destination", "platform=iOS Simulator,name=iPhone 15",
				"-destination", "platform=iOS Simulator,name=iPad Air"),
			diagnostics: []Diagnostic{{Kind: Override, Message: `"-destination generic/platform=iOS" replaced by additional option [-destination platform=iOS Simulator,name=iPhone 15 -destination platform=iOS Simulator,name=iPad Air]`}},
		},
		{
			name:        "a repeated build setting yields to the user's, as xcodebuild takes the last value",
			user:        Options{{Kind: BuildSetting, Name: "CODE_SIGNING_ALLOWED", Value: "YES"}},
			want:        append(without("CODE_SIGNING_ALLOWED=NO", 1), "CODE_SIGNING_ALLOWED=YES"),
			diagnostics: []Diagnostic{{Kind: Override, Message: `"CODE_SIGNING_ALLOWED=NO" replaced by additional option [CODE_SIGNING_ALLOWED=YES]`}},
		},
		{
			name:        "a repeated switch is dropped and reported as redundant",
			user:        Options{{Kind: Switch, Name: "-allowProvisioningUpdates"}},
			want:        append(without("-allowProvisioningUpdates", 1), "-allowProvisioningUpdates"),
			diagnostics: []Diagnostic{{Kind: RedundantOption, Message: `"-allowProvisioningUpdates" is already set by archive; it can be removed from the additional options`}},
		},
		{
			name:        "a default repeated with the same value is redundant, not an override",
			user:        Options{{Kind: ValueOption, Name: "-destination", Value: "generic/platform=iOS"}},
			want:        append(without("-destination", 2), "-destination", "generic/platform=iOS"),
			diagnostics: []Diagnostic{{Kind: RedundantOption, Message: `"-destination generic/platform=iOS" is already set by archive; it can be removed from the additional options`}},
		},
		{
			name:        "a value option repeated with the same value is still left for xcodebuild to refuse",
			user:        Options{{Kind: ValueOption, Name: "-scheme", Value: "App"}},
			want:        append(without("", 0), "-scheme", "App"),
			diagnostics: []Diagnostic{{Kind: RepeatedOption, Message: `"-scheme App" is set by archive and again as additional option [-scheme App]; xcodebuild refuses a repeated option`}},
		},
		{
			name: "actions and unknown tokens are appended, never merged",
			user: Options{{Kind: Action, Name: "clean"}, {Kind: Unknown, Name: "-scheme"}},
			want: append(without("", 0), "clean", "-scheme"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, diags := merge(derived, tt.user, policy)
			require.Equal(t, tt.want, got.Args())
			require.Equal(t, tt.diagnostics, diags)
		})
	}
}

func TestMerge_appendableKeepsDerivedEntries(t *testing.T) {
	derived := Options{
		{Kind: Action, Name: "test"},
		{Kind: ColonOption, Name: "-skip-testing", Value: "AppTests/Flaky"},
		{Kind: ColonOption, Name: "-skip-testing", Value: "AppTests/Slow"},
	}
	user := Options{{Kind: ColonOption, Name: "-skip-testing", Value: "AppTests/Manual"}}

	got, diags := merge(derived, user, actionPolicy{name: "test", appendable: []string{"-skip-testing"}})
	require.Equal(t, []string{"test", "-skip-testing:AppTests/Flaky", "-skip-testing:AppTests/Slow", "-skip-testing:AppTests/Manual"}, got.Args())
	require.Empty(t, diags)
}

func TestMerge_userDefaultDoesNotCollideWithAFlag(t *testing.T) {
	derived := Options{{Kind: ValueOption, Name: "-collect-test-diagnostics", Value: "never"}}
	user := Options{{Kind: UserDefault, Name: "-collect-test-diagnostics", Value: "on-failure"}}

	got, diags := merge(derived, user, actionPolicy{name: "test", defaults: []string{"-collect-test-diagnostics"}})

	require.Equal(t, []string{"-collect-test-diagnostics", "never", "-collect-test-diagnostics=on-failure"}, got.Args(), "the step's flag stays; xcodebuild ignores the = form")
	require.Equal(t, []Diagnostic{{Kind: SuspiciousUserDefault, Message: `"-collect-test-diagnostics never" is written as "-collect-test-diagnostics=on-failure" in the additional options: xcodebuild reads the "=" form as a user default and ignores it; use "-collect-test-diagnostics value"`}}, diags)
}
