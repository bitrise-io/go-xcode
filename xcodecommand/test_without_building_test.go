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

	// Even under Fail nothing is refused.
	_, err = TestWithoutBuilding(TestWithoutBuildingParams{XCTestRun: "a.xctestrun", OnlyTesting: []string{"AppTests"}, AdditionalOptions: []string{"-only-testing:AppUITests"}, Validation: Fail})
	require.NoError(t, err)
}

func TestTestWithoutBuilding_additionalOptions(t *testing.T) {
	t.Run("a user -destination joins the step's", func(t *testing.T) {
		cmd, err := TestWithoutBuilding(TestWithoutBuildingParams{XCTestRun: "a.xctestrun", Destination: "id=SIM-1", AdditionalOptions: []string{"-destination", "id=SIM-2"}})
		require.NoError(t, err)
		require.Equal(t, []string{"test-without-building", "-xctestrun", "a.xctestrun", "-destination", "id=SIM-1", "-destination", "id=SIM-2"}, cmd.Args())
		require.Empty(t, cmd.Diagnostics())
	})
	t.Run("a user -collect-test-diagnostics replaces the step's", func(t *testing.T) {
		cmd, err := TestWithoutBuilding(TestWithoutBuildingParams{XCTestRun: "a.xctestrun", CollectTestDiagnostics: "on-failure", AdditionalOptions: []string{"-collect-test-diagnostics", "never"}})
		require.NoError(t, err)
		require.Equal(t, []string{"test-without-building", "-xctestrun", "a.xctestrun", "-collect-test-diagnostics", "never"}, cmd.Args())
		require.Equal(t, []DiagnosticKind{Override}, kinds(cmd.Diagnostics()))
	})
	t.Run("a repeated -xctestrun is left for xcodebuild to refuse", func(t *testing.T) {
		cmd, err := TestWithoutBuilding(TestWithoutBuildingParams{XCTestRun: "a.xctestrun", AdditionalOptions: []string{"-xctestrun", "b.xctestrun"}})
		require.NoError(t, err)
		require.Equal(t, []DiagnosticKind{RepeatedOption}, kinds(cmd.Diagnostics()))
	})
	t.Run("a mode-switching flag is rejected", func(t *testing.T) {
		cmd, err := TestWithoutBuilding(TestWithoutBuildingParams{XCTestRun: "a.xctestrun", AdditionalOptions: []string{"-showBuildSettings"}})
		require.NoError(t, err)
		require.Equal(t, []DiagnosticKind{RejectedOption}, kinds(cmd.Diagnostics()))
	})
}
