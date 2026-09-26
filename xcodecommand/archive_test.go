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
