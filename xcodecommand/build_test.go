package xcodecommand

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// The cases port the v1 CommandBuilder table, split by action.
func TestBuild(t *testing.T) {
	tests := []struct {
		name   string
		params BuildParams
		want   []string
	}{
		{name: "destination", params: BuildParams{Destination: "id=3222ioocsdcsa1"}, want: []string{"build", "-destination", "id=3222ioocsdcsa1"}},
		{name: "scheme", params: BuildParams{Scheme: "project_scheme"}, want: []string{"build", "-scheme", "project_scheme"}},
		{name: "sdk", params: BuildParams{SDK: "iphonesimulator12"}, want: []string{"build", "-sdk", "iphonesimulator12"}},
		{name: "project", params: BuildParams{ProjectPath: "project.xcodeproj"}, want: []string{"build", "-project", "project.xcodeproj"}},
		{name: "workspace", params: BuildParams{ProjectPath: "project.xcworkspace"}, want: []string{"build", "-workspace", "project.xcworkspace"}},
		{name: "configuration", params: BuildParams{Configuration: "Debug"}, want: []string{"build", "-configuration", "Debug"}},
		{name: "disable code signing", params: BuildParams{DisableCodeSigning: true}, want: []string{"build", "CODE_SIGNING_ALLOWED=NO"}},
		{name: "xcconfig", params: BuildParams{XCConfigPath: "temp.xcconfig"}, want: []string{"build", "-xcconfig", "temp.xcconfig"}},
		{name: "clean first", params: BuildParams{Clean: true, Scheme: "App"}, want: []string{"clean", "build", "-scheme", "App"}},
		{
			name:   "everything",
			params: BuildParams{ProjectPath: "App.xcworkspace", Scheme: "App", Configuration: "Debug", Destination: "id=1", XCConfigPath: "t.xcconfig", SDK: "iphoneos", DisableCodeSigning: true, Authentication: &testAuth, AdditionalOptions: []string{"-quiet"}},
			want: []string{
				"build", "-workspace", "App.xcworkspace", "-scheme", "App", "-configuration", "Debug", "-destination", "id=1",
				"-xcconfig", "t.xcconfig", "-sdk", "iphoneos",
				"-allowProvisioningUpdates", "-authenticationKeyPath", "/key/path", "-authenticationKeyID", "keyID", "-authenticationKeyIssuerID", "issuerID",
				"CODE_SIGNING_ALLOWED=NO", "-quiet",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := Build(tt.params)
			require.NoError(t, err)
			require.Equal(t, tt.want, cmd.Args())
		})
	}
}

func TestAnalyze(t *testing.T) {
	cmd, err := Analyze(AnalyzeParams{ProjectPath: "App.xcodeproj", Scheme: "App", ResultBundlePath: "/tmp/Analyze.xcresult", DisableCodeSigning: true, AdditionalOptions: []string{"COMPILER_INDEX_STORE_ENABLE=NO"}})
	require.NoError(t, err)
	require.Equal(t, []string{"analyze", "-project", "App.xcodeproj", "-scheme", "App", "-resultBundlePath", "/tmp/Analyze.xcresult", "CODE_SIGNING_ALLOWED=NO", "COMPILER_INDEX_STORE_ENABLE=NO"}, cmd.Args())

	// steps-xcode-analyze only sets the result bundle path when the user did not; now the user's replaces it.
	cmd, err = Analyze(AnalyzeParams{ProjectPath: "App.xcodeproj", ResultBundlePath: "/tmp/Analyze.xcresult", AdditionalOptions: []string{"-resultBundlePath", "/mine.xcresult"}})
	require.NoError(t, err)
	require.Equal(t, []string{"analyze", "-project", "App.xcodeproj", "-resultBundlePath", "/mine.xcresult"}, cmd.Args())
	require.Len(t, cmd.Diagnostics(), 1)
}

func TestBuildForTesting(t *testing.T) {
	cmd, err := BuildForTesting(BuildForTestingParams{ProjectPath: "App.xcodeproj", Scheme: "App", Configuration: "Debug", Destination: "id=1", TestPlan: "FullTests", Authentication: &testAuth, AdditionalOptions: []string{"SYMROOT=/tmp/test_bundle", "-only-testing:AppTests"}})
	require.NoError(t, err)
	require.Equal(t, []string{
		"build-for-testing", "-project", "App.xcodeproj", "-scheme", "App", "-configuration", "Debug", "-destination", "id=1",
		"-allowProvisioningUpdates", "-authenticationKeyPath", "/key/path", "-authenticationKeyID", "keyID", "-authenticationKeyIssuerID", "issuerID",
		"-testPlan", "FullTests", "SYMROOT=/tmp/test_bundle", "-only-testing:AppTests",
	}, cmd.Args())
	require.Empty(t, cmd.Diagnostics(), "test selection flags are valid for build-for-testing")
}

func TestBuildFamily_userDestinationReplacesTheDefault(t *testing.T) {
	cmd, err := BuildForTesting(BuildForTestingParams{Scheme: "App", Destination: "id=SIM-1", AdditionalOptions: []string{"-destination", "id=SIM-2", "-arch", "arm64"}})
	require.NoError(t, err)
	require.Equal(t, []string{"build-for-testing", "-scheme", "App", "-destination", "id=SIM-2", "-arch", "arm64"}, cmd.Args())
	require.Equal(t, []DiagnosticKind{Override}, kinds(cmd.Diagnostics()))
}

func TestBuild_swiftPackageGetsNoContainerFlag(t *testing.T) {
	for _, path := range []string{"MyPackage/Package.swift", "MyPackage", "MyPackage/"} {
		cmd, err := Build(BuildParams{ProjectPath: path, Scheme: "CoolLibrary"})
		require.NoError(t, err)
		require.Equal(t, []string{"build", "-scheme", "CoolLibrary"}, cmd.Args(), path)
	}
}
