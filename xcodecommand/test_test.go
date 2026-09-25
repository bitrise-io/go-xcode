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
