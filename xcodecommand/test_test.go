package xcodecommand

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// The full case mirrors the xcode-test step's argument order.
func TestTest(t *testing.T) {
	tests := []struct {
		name   string
		params TestParams
		want   []string
	}{
		{
			name:   "project",
			params: TestParams{ProjectPath: "ios/project.xcodeproj", Scheme: "App", Destination: "id=SIM"},
			want:   []string{"-project", "ios/project.xcodeproj", "-scheme", "App", "test", "-destination", "id=SIM"},
		},
		{
			name:   "workspace",
			params: TestParams{ProjectPath: "ios/project.xcworkspace", Scheme: "App", Destination: "id=SIM"},
			want:   []string{"-workspace", "ios/project.xcworkspace", "-scheme", "App", "test", "-destination", "id=SIM"},
		},
		{
			name:   "Swift package gets no container flag",
			params: TestParams{ProjectPath: "MyPackage/Package.swift", Scheme: "CoolLibrary", Destination: "id=SIM"},
			want:   []string{"-scheme", "CoolLibrary", "test", "-destination", "id=SIM"},
		},
		{
			name: "everything, in the xcode-test step's order",
			params: TestParams{
				ProjectPath:                    "App.xcodeproj",
				Scheme:                         "App",
				Destination:                    "id=SIM",
				TestPlan:                       "Full",
				ResultBundlePath:               "/tmp/Test.xcresult",
				TestRepetitionMode:             TestRepetitionRetryOnFailure,
				MaximumTestRepetitions:         3,
				RelaunchTestsForEachRepetition: true,
				XCConfigPath:                   "/tmp/temp.xcconfig",
				Clean:                          true,
				SkipTesting:                    []string{"TestTarget1/TestClass1", "TestTarget2"},
				CollectTestDiagnostics:         "never",
				AdditionalOptions:              []string{"-quiet"},
			},
			want: []string{
				"-project", "App.xcodeproj", "-scheme", "App", "clean", "test", "-destination", "id=SIM",
				"-testPlan", "Full", "-resultBundlePath", "/tmp/Test.xcresult",
				"-retry-tests-on-failure", "-test-iterations", "3", "-test-repetition-relaunch-enabled", "YES",
				"-xcconfig", "/tmp/temp.xcconfig",
				"-skip-testing:TestTarget1/TestClass1", "-skip-testing:TestTarget2",
				"-collect-test-diagnostics", "never", "-quiet",
			},
		},
		{
			name:   "until failure",
			params: TestParams{Scheme: "App", TestRepetitionMode: TestRepetitionUntilFailure, MaximumTestRepetitions: 5},
			want:   []string{"-scheme", "App", "test", "-run-tests-until-failure", "-test-iterations", "5"},
		},
		{
			name:   "no repetition mode adds no iteration flags",
			params: TestParams{Scheme: "App", TestRepetitionMode: TestRepetitionNone, MaximumTestRepetitions: 5},
			want:   []string{"-scheme", "App", "test"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := Test(tt.params)
			require.NoError(t, err)
			require.Equal(t, tt.want, cmd.Args())
			require.Empty(t, cmd.Diagnostics())
		})
	}
}

func TestTest_additionalOptions(t *testing.T) {
	t.Run("user -skip-testing entries join the quarantined ones", func(t *testing.T) {
		cmd, err := Test(TestParams{Scheme: "App", SkipTesting: []string{"AppTests/Flaky"}, AdditionalOptions: []string{"-skip-testing:AppTests/Manual"}})
		require.NoError(t, err)
		require.Equal(t, []string{"-scheme", "App", "test", "-skip-testing:AppTests/Flaky", "-skip-testing:AppTests/Manual"}, cmd.Args())
		require.Empty(t, cmd.Diagnostics())
	})

	t.Run("a user -destination joins the simulator's for a multi-destination run", func(t *testing.T) {
		cmd, err := Test(TestParams{Scheme: "App", Destination: "id=SIM-1", AdditionalOptions: []string{"-destination", "id=SIM-2"}})
		require.NoError(t, err)
		require.Equal(t, []string{"-scheme", "App", "test", "-destination", "id=SIM-1", "-destination", "id=SIM-2"}, cmd.Args())
		require.Empty(t, cmd.Diagnostics())
	})

	t.Run("a user -collect-test-diagnostics replaces the step's", func(t *testing.T) {
		cmd, err := Test(TestParams{Scheme: "App", CollectTestDiagnostics: "on-failure", AdditionalOptions: []string{"-collect-test-diagnostics", "never"}})
		require.NoError(t, err)
		require.Equal(t, []string{"-scheme", "App", "test", "-collect-test-diagnostics", "never"}, cmd.Args())
		require.Equal(t, []DiagnosticKind{Override}, kinds(cmd.Diagnostics()))
	})

	t.Run("a repeated -resultBundlePath is left for xcodebuild to refuse", func(t *testing.T) {
		cmd, err := Test(TestParams{Scheme: "App", ResultBundlePath: "/tmp/Test.xcresult", AdditionalOptions: []string{"-resultBundlePath", "/mine.xcresult"}})
		require.NoError(t, err)
		require.Equal(t, []string{"-scheme", "App", "test", "-resultBundlePath", "/tmp/Test.xcresult", "-resultBundlePath", "/mine.xcresult"}, cmd.Args())
		require.Equal(t, []DiagnosticKind{RepeatedOption}, kinds(cmd.Diagnostics()))
	})

	t.Run("test selection flags are valid for test", func(t *testing.T) {
		cmd, err := Test(TestParams{Scheme: "App", AdditionalOptions: []string{"-only-testing:AppTests", "-parallel-testing-enabled", "NO", "-enableCodeCoverage", "YES"}})
		require.NoError(t, err)
		require.Empty(t, cmd.Diagnostics())
	})
}
