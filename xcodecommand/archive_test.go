package xcodecommand

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var testAuth = Authentication{KeyPath: "/key/path", KeyID: "keyID", IssuerID: "issuerID"}

func TestArchive(t *testing.T) {
	tests := []struct {
		name   string
		params ArchiveParams
		want   []string
	}{
		{
			name:   "archive path",
			params: ArchiveParams{ArchivePath: "archive/path"},
			want:   []string{"archive", "-archivePath", "archive/path"},
		},
		{
			name:   "authentication",
			params: ArchiveParams{Authentication: &testAuth},
			want:   []string{"archive", "-allowProvisioningUpdates", "-authenticationKeyPath", "/key/path", "-authenticationKeyID", "keyID", "-authenticationKeyIssuerID", "issuerID"},
		},
		{
			name:   "workspace is detected from the extension",
			params: ArchiveParams{ProjectPath: "App.xcworkspace", Scheme: "App"},
			want:   []string{"archive", "-workspace", "App.xcworkspace", "-scheme", "App"},
		},
		{
			name: "steps-xcode-archive: clean, xcconfig, archive path, API key, additional options last",
			params: ArchiveParams{
				ProjectPath:       "App.xcworkspace",
				Scheme:            "App",
				Configuration:     "Release",
				XCConfigPath:      "/tmp/temp.xcconfig",
				ArchivePath:       "/tmp/App.xcarchive",
				Clean:             true,
				Authentication:    &testAuth,
				AdditionalOptions: []string{"-destination", "generic/platform=iOS", "-skipPackagePluginValidation"},
			},
			want: []string{
				"clean", "archive",
				"-workspace", "App.xcworkspace",
				"-scheme", "App",
				"-configuration", "Release",
				"-xcconfig", "/tmp/temp.xcconfig",
				"-archivePath", "/tmp/App.xcarchive",
				"-allowProvisioningUpdates",
				"-authenticationKeyPath", "/key/path",
				"-authenticationKeyID", "keyID",
				"-authenticationKeyIssuerID", "issuerID",
				"-destination", "generic/platform=iOS", "-skipPackagePluginValidation",
			},
		},
		{
			name: "a user -destination replaces the step's default instead of duplicating it",
			params: ArchiveParams{
				ProjectPath:       "App.xcodeproj",
				Scheme:            "App",
				Destination:       "generic/platform=iOS",
				AdditionalOptions: []string{"-destination", "generic/platform=iOS Simulator"},
			},
			want: []string{"archive", "-project", "App.xcodeproj", "-scheme", "App", "-destination", "generic/platform=iOS Simulator"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := Archive(tt.params)
			require.NoError(t, err)
			require.Equal(t, tt.want, cmd.Args())

			again, err := Archive(tt.params)
			require.NoError(t, err)
			require.Equal(t, tt.want, again.Args(), "rendering must be deterministic")
		})
	}
}

// The additional options below all occur in production; under Warn they pass through as
// the steps passed them before, and are reported. Under Fail they are refused.
func TestArchive_validation(t *testing.T) {
	tests := []struct {
		name     string
		options  []string
		wantArgs []string
		wantKind DiagnosticKind
		wantErr  string
	}{
		{
			name:     "actions in the options",
			options:  []string{"clean", "archive"},
			wantArgs: []string{"archive", "-project", "App.xcodeproj", "clean", "archive"},
			wantKind: ActionInOptions,
			wantErr:  `invalid additional option: "clean" is a build action, and archive sets its own actions`,
		},
		{
			name:     "test-only flags",
			options:  []string{"-test-iterations", "2", "-retry-tests-on-failure"},
			wantArgs: []string{"archive", "-project", "App.xcodeproj", "-test-iterations", "2", "-retry-tests-on-failure"},
			wantKind: RejectedOption,
			wantErr:  `invalid additional option: "-test-iterations 2" is not valid for archive: applies to test actions only`,
		},
		{
			name:     "free-form value flag left without its value",
			options:  []string{"-skipMacroValidation", "-destination"},
			wantArgs: []string{"archive", "-project", "App.xcodeproj", "-skipMacroValidation", "-destination"},
			wantKind: MalformedOption,
			wantErr:  `invalid additional option: "-destination" requires a value`,
		},
		{
			name:     "flag quoted together with its value",
			options:  []string{"-destination generic/platform=iOS"},
			wantArgs: []string{"archive", "-project", "App.xcodeproj", "-destination generic/platform=iOS"},
			wantKind: MalformedOption,
			wantErr:  `invalid additional option: "-destination generic/platform=iOS" contains whitespace: quote only the value, not the flag and the value together`,
		},
		{
			name:     "build setting written with a leading dash",
			options:  []string{"-ENABLE_BITCODE=NO"},
			wantArgs: []string{"archive", "-project", "App.xcodeproj", "-ENABLE_BITCODE=NO"},
			wantKind: SuspiciousUserDefault,
			wantErr:  `invalid additional option: "-ENABLE_BITCODE=NO" looks like the build setting ENABLE_BITCODE=NO written with a leading dash; xcodebuild accepts it as a user default and the setting never applies`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := Archive(ArchiveParams{ProjectPath: "App.xcodeproj", AdditionalOptions: tt.options})
			require.NoError(t, err)
			require.Equal(t, tt.wantArgs, cmd.Args())
			require.NotEmpty(t, cmd.Diagnostics())
			require.Equal(t, tt.wantKind, cmd.Diagnostics()[0].Kind)

			_, err = Archive(ArchiveParams{ProjectPath: "App.xcodeproj", AdditionalOptions: tt.options, Validation: Fail})
			require.EqualError(t, err, tt.wantErr)
		})
	}
}

// A repeated option is not merged away: both go through and xcodebuild refuses it, so
// the user fixes the configuration. Fail reports it before xcodebuild runs.
func TestArchive_repeatedOption(t *testing.T) {
	params := ArchiveParams{ProjectPath: "App.xcodeproj", XCConfigPath: "/tmp/temp.xcconfig", AdditionalOptions: []string{"-xcconfig", "mine.xcconfig"}}

	cmd, err := Archive(params)
	require.NoError(t, err)
	require.Equal(t, []string{"archive", "-project", "App.xcodeproj", "-xcconfig", "/tmp/temp.xcconfig", "-xcconfig", "mine.xcconfig"}, cmd.Args())
	require.Equal(t, []Diagnostic{{Kind: RepeatedOption, Message: `"-xcconfig /tmp/temp.xcconfig" is set by archive and again as additional option [-xcconfig mine.xcconfig]; xcodebuild refuses a repeated option`}}, cmd.Diagnostics())

	params.Validation = Fail
	_, err = Archive(params)
	require.EqualError(t, err, `invalid additional option: "-xcconfig /tmp/temp.xcconfig" is set by archive and again as additional option [-xcconfig mine.xcconfig]; xcodebuild refuses a repeated option`)

	cmd, err = Archive(ArchiveParams{ProjectPath: "App.xcodeproj", AdditionalOptions: []string{"-xcconfig", "mine.xcconfig"}})
	require.NoError(t, err, "without a derived xcconfig the user's is the only one")
	require.Equal(t, []string{"archive", "-project", "App.xcodeproj", "-xcconfig", "mine.xcconfig"}, cmd.Args())
	require.Empty(t, cmd.Diagnostics())
}

func TestArchive_overridesAreReported(t *testing.T) {
	cmd, err := Archive(ArchiveParams{ProjectPath: "App.xcodeproj", Destination: "generic/platform=iOS", AdditionalOptions: []string{"-destination", "id=SIM"}})
	require.NoError(t, err)
	require.Equal(t, []Diagnostic{{Kind: Override, Message: `"-destination generic/platform=iOS" replaced by additional option [-destination id=SIM]`}}, cmd.Diagnostics())
}
