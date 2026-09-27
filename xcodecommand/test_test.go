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
				OnlyTesting:                    []string{"AppTests"},
				SkipTesting:                    []string{"TestTarget1/TestClass1", "TestTarget2"},
				CollectTestDiagnostics:         "never",
				AdditionalOptions:              []string{"-quiet"},
			},
			want: []string{
				"-project", "App.xcodeproj", "-scheme", "App", "clean", "test", "-destination", "id=SIM",
				"-testPlan", "Full", "-xcconfig", "/tmp/temp.xcconfig",
				"-resultBundlePath", "/tmp/Test.xcresult",
				"-retry-tests-on-failure", "-test-iterations", "3", "-test-repetition-relaunch-enabled", "YES",
				"-only-testing:AppTests", "-skip-testing:TestTarget1/TestClass1", "-skip-testing:TestTarget2",
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

// The modes are the test steps' test_repetition_mode values; every mode but none repeats
// up to MaximumTestRepetitions times.
func TestTest_repetition(t *testing.T) {
	tests := []struct {
		mode     TestRepetitionMode
		max      int
		relaunch bool
		want     []string
		wantErr  string
	}{
		{mode: "", want: nil},
		{mode: TestRepetitionNone, max: 3, want: nil},
		{mode: TestRepetitionUntilFailure, max: 3, want: []string{"-run-tests-until-failure", "-test-iterations", "3"}},
		{mode: TestRepetitionRetryOnFailure, max: 3, relaunch: true, want: []string{"-retry-tests-on-failure", "-test-iterations", "3", "-test-repetition-relaunch-enabled", "YES"}},
		{mode: TestRepetitionUpUntilMaximumRuns, max: 5, want: []string{"-test-iterations", "5"}},
		{mode: "retry-on-failure", max: 3, wantErr: `test_repetition_mode "retry-on-failure" is not one of none, until_failure, retry_on_failure, up_until_maximum_repetitions`},
		{mode: TestRepetitionRetryOnFailure, max: 0, wantErr: "test_repetition_mode retry_on_failure needs a maximum_test_repetitions of at least 2, got 0"},
		{mode: TestRepetitionNone, relaunch: true, wantErr: "relaunch_tests_for_each_repetition needs a test_repetition_mode other than none"},
	}
	for _, tt := range tests {
		t.Run(string(tt.mode)+tt.wantErr, func(t *testing.T) {
			params := TestParams{TestRepetitionMode: tt.mode, MaximumTestRepetitions: tt.max, RelaunchTestsForEachRepetition: tt.relaunch}
			cmd, err := Test(params)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				_, err = TestWithoutBuilding(TestWithoutBuildingParams{TestRepetitionMode: tt.mode, MaximumTestRepetitions: tt.max, RelaunchTestsForEachRepetition: tt.relaunch})
				require.ErrorContains(t, err, tt.wantErr, "test-without-building shares the validation")
				return
			}
			require.NoError(t, err)
			require.Equal(t, append([]string{"test"}, tt.want...), cmd.Args())
		})
	}
}
