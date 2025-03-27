package xcodebuild

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExportCommandModel_cmdSlice(t *testing.T) {
	tests := []struct {
		name               string
		archivePath        string
		exportDir          string
		exportOptionsPlist string
		authentication     *AuthenticationParams
		want               []string
	}{
		{
			name:               "basic export",
			archivePath:        "sample.xcarchive",
			exportDir:          "/var/exported",
			exportOptionsPlist: "/var/export_options.plist",
			want: []string{"xcodebuild",
				"-exportArchive",
				"-archivePath", "sample.xcarchive",
				"-exportPath", "/var/exported",
				"-exportOptionsPlist", "/var/export_options.plist",
			},
		},
		{
			name:        "export with authentication",
			archivePath: "sample.xcarchive",
			authentication: &AuthenticationParams{
				KeyID:     "keyID",
				IsssuerID: "issuerID",
				KeyPath:   "/key/path",
			},
			want: []string{"xcodebuild",
				"-exportArchive",
				"-archivePath", "sample.xcarchive",
				"-allowProvisioningUpdates",
				"-authenticationKeyPath", "/key/path",
				"-authenticationKeyID", "keyID",
				"-authenticationKeyIssuerID", "issuerID",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewExportCommand()
			c.SetArchivePath(tt.archivePath)
			c.SetExportDir(tt.exportDir)
			c.SetExportOptionsPlist(tt.exportOptionsPlist)
			if tt.authentication != nil {
				c.SetAuthentication(*tt.authentication)
			}

			got := c.cmdSlice()
			require.Equal(t, tt.want, got)

			got2 := c.cmdSlice()
			require.Equal(t, tt.want, got2, "Second run should return the same result")
		})
	}
}
