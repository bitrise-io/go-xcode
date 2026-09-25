package xcodecommand

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// The full case mirrors the xcode-test-without-building step's argument order.
func TestTestWithoutBuilding(t *testing.T) {
	cmd, err := TestWithoutBuilding(TestWithoutBuildingParams{
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
	})
	require.NoError(t, err)
	require.Equal(t, []string{
		"test-without-building", "-xctestrun", "/tmp/App_iphonesimulator.xctestrun", "-destination", "id=SIM",
		"-resultBundlePath", "/tmp/Test-App.xcresult",
		"-retry-tests-on-failure", "-test-iterations", "3", "-test-repetition-relaunch-enabled", "YES",
		"-only-testing:AppTests/Login", "-only-testing:AppUITests", "-skip-testing:AppTests/Login/testSlow",
		"-collect-test-diagnostics", "on-failure",
		"-parallel-testing-enabled", "NO",
	}, cmd.Args())
	require.Empty(t, cmd.Diagnostics())
}

func TestTestWithoutBuilding_minimal(t *testing.T) {
	cmd, err := TestWithoutBuilding(TestWithoutBuildingParams{XCTestRun: "a.xctestrun", Destination: "id=SIM", ResultBundlePath: "/tmp/r.xcresult"})
	require.NoError(t, err)
	require.Equal(t, []string{"test-without-building", "-xctestrun", "a.xctestrun", "-destination", "id=SIM", "-resultBundlePath", "/tmp/r.xcresult"}, cmd.Args())
}

// The test selection flags are what this command is about: entries from the step inputs
// and from xcodebuild_options must all reach xcodebuild, which applies -only-testing
// before -skip-testing.
func TestTestWithoutBuilding_testSelection(t *testing.T) {
	cmd, err := TestWithoutBuilding(TestWithoutBuildingParams{
		XCTestRun:         "a.xctestrun",
		OnlyTesting:       []string{"AppTests"},
		SkipTesting:       []string{"AppTests/Flaky"},
		AdditionalOptions: []string{"-only-testing:AppUITests", "-skip-testing:AppTests/Slow", "-only-test-configuration", "Debug", "-skip-test-configuration", "Release"},
	})
	require.NoError(t, err)
	require.Equal(t, []string{
		"test-without-building", "-xctestrun", "a.xctestrun",
		"-only-testing:AppTests", "-skip-testing:AppTests/Flaky",
		"-only-testing:AppUITests", "-skip-testing:AppTests/Slow", "-only-test-configuration", "Debug", "-skip-test-configuration", "Release",
	}, cmd.Args())
	require.Empty(t, cmd.Diagnostics(), "test selection flags are appendable, never repeated or rejected")
}
