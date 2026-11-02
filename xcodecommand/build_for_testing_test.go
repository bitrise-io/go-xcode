package xcodecommand

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildForTesting(t *testing.T) {
	cmd, err := BuildForTesting(BuildForTestingParams{ProjectPath: "App.xcodeproj", Scheme: "App", Configuration: "Debug", Destination: "id=1", TestPlan: "FullTests", Authentication: &testAuth, AdditionalOptions: []string{"SYMROOT=/tmp/test_bundle", "-only-testing:AppTests"}})
	require.NoError(t, err)
	require.Equal(t, []string{
		"build-for-testing", "-project", "App.xcodeproj", "-scheme", "App", "-configuration", "Debug", "-destination", "id=1",
		"-allowProvisioningUpdates", "-authenticationKeyPath", "/key/path", "-authenticationKeyID", "keyID", "-authenticationKeyIssuerID", "issuerID",
		"-testPlan", "FullTests", "SYMROOT=/tmp/test_bundle", "-only-testing:AppTests",
	}, cmd.Args())
}
