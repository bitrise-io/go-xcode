package xcodecommand

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFail_ignoresInformationalDiagnostics(t *testing.T) {
	// A redundant switch and a replaced default are informational: Fail still succeeds.
	cmd, err := Archive(ArchiveParams{
		ProjectPath:       "App.xcodeproj",
		Destination:       "generic/platform=iOS",
		Authentication:    &testAuth,
		AdditionalOptions: []string{"-allowProvisioningUpdates", "-destination", "id=SIM"},
		Validation:        Fail,
	})
	require.NoError(t, err)
	require.Equal(t, []DiagnosticKind{Override, RedundantOption}, kinds(cmd.Diagnostics()), "reported in derived-flag order")
	require.Equal(t, []string{
		"archive", "-project", "App.xcodeproj",
		"-authenticationKeyPath", "/key/path", "-authenticationKeyID", "keyID", "-authenticationKeyIssuerID", "issuerID",
		"-allowProvisioningUpdates", "-destination", "id=SIM",
	}, cmd.Args())
}

func kinds(diagnostics []Diagnostic) []DiagnosticKind {
	var out []DiagnosticKind
	for _, d := range diagnostics {
		out = append(out, d.Kind)
	}
	return out
}
