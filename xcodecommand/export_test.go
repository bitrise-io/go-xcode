package xcodecommand

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExportArchive(t *testing.T) {
	tests := []struct {
		name   string
		params ExportArchiveParams
		want   []string
	}{
		{
			name:   "basic export",
			params: ExportArchiveParams{ArchivePath: "sample.xcarchive", ExportPath: "/var/exported", ExportOptionsPlist: "/var/export_options.plist"},
			want:   []string{"-exportArchive", "-archivePath", "sample.xcarchive", "-exportPath", "/var/exported", "-exportOptionsPlist", "/var/export_options.plist"},
		},
		{
			name:   "export with authentication",
			params: ExportArchiveParams{ArchivePath: "sample.xcarchive", Authentication: &testAuth},
			want:   []string{"-exportArchive", "-archivePath", "sample.xcarchive", "-allowProvisioningUpdates", "-authenticationKeyPath", "/key/path", "-authenticationKeyID", "keyID", "-authenticationKeyIssuerID", "issuerID"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := ExportArchive(tt.params)
			require.NoError(t, err)
			require.Equal(t, tt.want, cmd.Args())
		})
	}
}
