package xcodecommand

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTestWithoutBuilding(t *testing.T) {
	tests := []struct {
		name   string
		params TestWithoutBuildingParams
		want   []string
	}{
		{
			name:   "minimal",
			params: TestWithoutBuildingParams{XCTestRun: "a.xctestrun", Destination: "id=SIM", ResultBundlePath: "/tmp/r.xcresult"},
			want:   []string{"test-without-building", "-xctestrun", "a.xctestrun", "-destination", "id=SIM", "-resultBundlePath", "/tmp/r.xcresult"},
		},
		{
			name: "every field, in the xcode-test-without-building step's argument order",
			params: TestWithoutBuildingParams{
				XCTestRun:                      "/tmp/App_iphonesimulator.xctestrun",
				Destination:                    "id=SIM",
				ResultBundlePath:               "/tmp/Test-App.xcresult",
				TestRepetitionMode:             TestRepetitionRetryOnFailure,
				MaximumTestRepetitions:         3,
				RelaunchTestsForEachRepetition: true,
				OnlyTesting:                    []string{"AppTests/Login", "AppUITests"},
				SkipTesting:                    []string{"AppTests/Login/testSlow"},
				CollectTestDiagnostics:         "on-failure",
				AdditionalOptions:              []string{"-parallel-testing-enabled", "NO"},
			},
			want: []string{
				"test-without-building", "-xctestrun", "/tmp/App_iphonesimulator.xctestrun", "-destination", "id=SIM",
				"-resultBundlePath", "/tmp/Test-App.xcresult",
				"-retry-tests-on-failure", "-test-iterations", "3", "-test-repetition-relaunch-enabled", "YES",
				"-only-testing:AppTests/Login", "-only-testing:AppUITests", "-skip-testing:AppTests/Login/testSlow",
				"-collect-test-diagnostics", "on-failure",
				"-parallel-testing-enabled", "NO",
			},
		},
		{
			// Entries from the step inputs and from xcodebuild_options all reach xcodebuild,
			// which applies -only-testing before -skip-testing: they are appendable.
			name: "test selection from inputs and options",
			params: TestWithoutBuildingParams{
				XCTestRun:         "a.xctestrun",
				OnlyTesting:       []string{"AppTests"},
				SkipTesting:       []string{"AppTests/Flaky"},
				AdditionalOptions: []string{"-only-testing:AppUITests", "-skip-testing:AppTests/Slow", "-only-test-configuration", "Debug", "-skip-test-configuration", "Release"},
			},
			want: []string{
				"test-without-building", "-xctestrun", "a.xctestrun",
				"-only-testing:AppTests", "-skip-testing:AppTests/Flaky",
				"-only-testing:AppUITests", "-skip-testing:AppTests/Slow", "-only-test-configuration", "Debug", "-skip-test-configuration", "Release",
			},
		},
		{
			name: "a scheme, rendered as test renders it",
			params: TestWithoutBuildingParams{
				ProjectPath:      "App.xcworkspace",
				Scheme:           "App",
				TestPlan:         "Smoke",
				Destination:      "id=SIM",
				ResultBundlePath: "/tmp/r.xcresult",
			},
			want: []string{
				"-workspace", "App.xcworkspace", "-scheme", "App", "test-without-building",
				"-destination", "id=SIM", "-testPlan", "Smoke", "-resultBundlePath", "/tmp/r.xcresult",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := TestWithoutBuilding(tt.params)
			require.NoError(t, err)
			require.Equal(t, tt.want, cmd.Args())
			require.Empty(t, cmd.Diagnostics())
		})
	}
}

func TestTestWithoutBuilding_XCTestRunOrScheme(t *testing.T) {
	// xcodebuild: "Cannot use -xctestrun with -project, -workspace or -scheme options."
	_, err := TestWithoutBuilding(TestWithoutBuildingParams{XCTestRun: "a.xctestrun", ProjectPath: "App.xcodeproj", Scheme: "App"})
	require.EqualError(t, err, "test-without-building runs an xctestrun or a scheme, not both")

	_, err = TestWithoutBuilding(TestWithoutBuildingParams{ProjectPath: "App.xcodeproj", Destination: "id=SIM"})
	require.EqualError(t, err, "test-without-building needs an xctestrun or a scheme")
}

func TestTestWithoutBuilding_Rejections(t *testing.T) {
	// Each form rejects the other's flags: the xctestrun form a -scheme or a
	// -testProductsPath, the scheme form an -xctestrun.
	cmd, err := TestWithoutBuilding(TestWithoutBuildingParams{XCTestRun: "a.xctestrun", AdditionalOptions: []string{"-scheme", "App", "-testProductsPath", "App.xctestproducts"}})
	require.NoError(t, err)
	require.Equal(t, []DiagnosticKind{RejectedOption, RejectedOption}, kinds(cmd.Diagnostics()))

	cmd, err = TestWithoutBuilding(TestWithoutBuildingParams{ProjectPath: "App.xcodeproj", Scheme: "App", AdditionalOptions: []string{"-xctestrun", "a.xctestrun"}})
	require.NoError(t, err)
	require.Equal(t, []DiagnosticKind{RejectedOption}, kinds(cmd.Diagnostics()))

	_, err = TestWithoutBuilding(TestWithoutBuildingParams{ProjectPath: "App.xcodeproj", Scheme: "App", AdditionalOptions: []string{"-xctestrun", "a.xctestrun"}, Validation: Fail})
	require.Error(t, err)
}
