package xcodecommand

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildForTesting(t *testing.T) {
	tests := []struct {
		name   string
		params BuildForTestingParams
		want   []string
	}{
		{
			name:   "steps-xcode-build-for-test: destination, API key, test plan, additional options last",
			params: BuildForTestingParams{ProjectPath: "App.xcodeproj", Scheme: "App", Configuration: "Debug", Destination: "id=1", TestPlan: "FullTests", Authentication: &testAuth, AdditionalOptions: additional("SYMROOT=/tmp/test_bundle", "-only-testing:AppTests")},
			want: []string{
				"build-for-testing", "-project", "App.xcodeproj", "-scheme", "App", "-configuration", "Debug", "-destination", "id=1",
				"-allowProvisioningUpdates", "-authenticationKeyPath", "/key/path", "-authenticationKeyID", "keyID", "-authenticationKeyIssuerID", "issuerID",
				"-testPlan", "FullTests", "SYMROOT=/tmp/test_bundle", "-only-testing:AppTests",
			},
		},
		{
			// The step's -destination is a default: the user's replaces it.
			name:   "a user -destination replaces the step's default",
			params: BuildForTestingParams{Scheme: "App", Destination: "generic/platform=iOS", AdditionalOptions: additional("-destination", "generic/platform=iOS Simulator")},
			want:   []string{"build-for-testing", "-scheme", "App", "-destination", "generic/platform=iOS Simulator"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := BuildForTesting(tt.params)
			require.NoError(t, err)
			require.Equal(t, tt.want, cmd.Args())
		})
	}
}
