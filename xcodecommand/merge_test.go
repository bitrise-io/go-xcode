package xcodecommand

import (
	"slices"
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
		name    string
		user    Options
		want    []string
		dropped []string // the Step's options left off the command line
	}{
		{
			name: "no user options leaves the derived options untouched",
			want: base,
		},
		{
			name: "options with a new key are appended in the order given",
			user: Options{{Kind: Switch, Name: "-quiet"}, {Kind: ValueOption, Name: "-sdk", Value: "iphoneos"}},
			want: append(slices.Clone(base), "-quiet", "-sdk", "iphoneos"),
		},
		{
			name: "a repeated value option is passed through for xcodebuild to refuse",
			user: Options{{Kind: ValueOption, Name: "-configuration", Value: "Debug"}},
			want: append(slices.Clone(base), "-configuration", "Debug"),
		},
		{
			name: "a repeated owned path is passed through the same way",
			user: Options{{Kind: ValueOption, Name: "-xcconfig", Value: "mine.xcconfig"}},
			want: append(slices.Clone(base), "-xcconfig", "mine.xcconfig"),
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
			dropped: []string{"-destination generic/platform=iOS"},
		},
		{
			name:    "a repeated build setting yields to the user's, as xcodebuild takes the last value",
			user:    Options{{Kind: BuildSetting, Name: "CODE_SIGNING_ALLOWED", Value: "YES"}},
			want:    append(without("CODE_SIGNING_ALLOWED=NO", 1), "CODE_SIGNING_ALLOWED=YES"),
			dropped: []string{"CODE_SIGNING_ALLOWED=NO"},
		},
		{
			name:    "a repeated switch yields, so it is sent once",
			user:    Options{{Kind: Switch, Name: "-allowProvisioningUpdates"}},
			want:    append(without("-allowProvisioningUpdates", 1), "-allowProvisioningUpdates"),
			dropped: []string{"-allowProvisioningUpdates"},
		},
		{
			name:    "a switch the user repeats on top of the step's yields the same way",
			user:    Options{{Kind: Switch, Name: "-allowProvisioningUpdates"}, {Kind: Switch, Name: "-allowProvisioningUpdates"}},
			want:    append(without("-allowProvisioningUpdates", 1), "-allowProvisioningUpdates", "-allowProvisioningUpdates"),
			dropped: []string{"-allowProvisioningUpdates"},
		},
		{
			name:    "a default repeated with the same value yields too",
			user:    Options{{Kind: ValueOption, Name: "-destination", Value: "generic/platform=iOS"}},
			want:    append(without("-destination", 2), "-destination", "generic/platform=iOS"),
			dropped: []string{"-destination generic/platform=iOS"},
		},
		{
			name: "a value option repeated with the same value is still left for xcodebuild to refuse",
			user: Options{{Kind: ValueOption, Name: "-scheme", Value: "App"}},
			want: append(slices.Clone(base), "-scheme", "App"),
		},
		{
			name: "actions and unknown tokens are appended, never merged",
			user: Options{{Kind: Action, Name: "clean"}, {Kind: Unknown, Name: "-scheme"}},
			want: append(slices.Clone(base), "clean", "-scheme"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := merge(derived, tt.user, policy)
			require.Equal(t, tt.want, got.Args())
			require.Equal(t, tt.dropped, dropped(got))
		})
	}
}

// dropped renders the options the merge left off the command line.
func dropped(opts Options) []string {
	var out []string
	for _, o := range opts {
		if o.dropped {
			out = append(out, o.String())
		}
	}
	return out
}

func TestMerge_appendableKeepsDerivedEntries(t *testing.T) {
	derived := Options{
		{Kind: Action, Name: "test"},
		{Kind: ColonOption, Name: "-skip-testing", Value: "AppTests/Flaky"},
		{Kind: ColonOption, Name: "-skip-testing", Value: "AppTests/Slow"},
	}
	user := Options{{Kind: ColonOption, Name: "-skip-testing", Value: "AppTests/Manual"}}

	got := merge(derived, user, actionPolicy{name: "test", appendable: []string{"-skip-testing"}})
	require.Equal(t, []string{"test", "-skip-testing:AppTests/Flaky", "-skip-testing:AppTests/Slow", "-skip-testing:AppTests/Manual"}, got.Args())
	require.Empty(t, dropped(got))
}

func TestMerge_userDefaultDoesNotReplaceAFlag(t *testing.T) {
	derived := Options{{Kind: ValueOption, Name: "-collect-test-diagnostics", Value: "never"}}
	user := Options{{Kind: UserDefault, Name: "-collect-test-diagnostics", Value: "on-failure"}}

	got := merge(derived, user, actionPolicy{name: "test", defaults: []string{"-collect-test-diagnostics"}})

	require.Equal(t, []string{"-collect-test-diagnostics", "never", "-collect-test-diagnostics=on-failure"}, got.Args(), "the step's flag stays; xcodebuild ignores the = form")
	require.Empty(t, dropped(got), "a user default keys as -flag=, so it never replaces the flag; lint reports it")
}
