package xcodecommand

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func build(p BuildParams) func() (Command, error) { return func() (Command, error) { return Build(p) } }

// The build rows port the v1 CommandBuilder table; analyze and build-for-testing share its
// rendering.
func TestBuild(t *testing.T) {
	tests := []struct {
		name  string
		build func() (Command, error)
		want  []string
	}{
		{name: "destination", build: build(BuildParams{Destination: "id=3222ioocsdcsa1"}), want: []string{"build", "-destination", "id=3222ioocsdcsa1"}},
		{name: "scheme", build: build(BuildParams{Scheme: "project_scheme"}), want: []string{"build", "-scheme", "project_scheme"}},
		{name: "sdk", build: build(BuildParams{SDK: "iphonesimulator12"}), want: []string{"build", "-sdk", "iphonesimulator12"}},
		{name: "project", build: build(BuildParams{ProjectPath: "project.xcodeproj"}), want: []string{"build", "-project", "project.xcodeproj"}},
		{name: "workspace", build: build(BuildParams{ProjectPath: "project.xcworkspace"}), want: []string{"build", "-workspace", "project.xcworkspace"}},
		{name: "configuration", build: build(BuildParams{Configuration: "Debug"}), want: []string{"build", "-configuration", "Debug"}},
		{name: "disable code signing", build: build(BuildParams{DisableCodeSigning: true}), want: []string{"build", "CODE_SIGNING_ALLOWED=NO"}},
		{name: "xcconfig", build: build(BuildParams{XCConfigPath: "temp.xcconfig"}), want: []string{"build", "-xcconfig", "temp.xcconfig"}},
		{name: "clean first", build: build(BuildParams{Clean: true, Scheme: "App"}), want: []string{"clean", "build", "-scheme", "App"}},
		{
			name:  "everything",
			build: build(BuildParams{ProjectPath: "App.xcworkspace", Scheme: "App", Configuration: "Debug", Destination: "id=1", XCConfigPath: "t.xcconfig", SDK: "iphoneos", DisableCodeSigning: true, Authentication: &testAuth, AdditionalOptions: []string{"-quiet"}}),
			want: []string{
				"build", "-workspace", "App.xcworkspace", "-scheme", "App", "-configuration", "Debug", "-destination", "id=1",
				"-xcconfig", "t.xcconfig", "-sdk", "iphoneos",
				"-allowProvisioningUpdates", "-authenticationKeyPath", "/key/path", "-authenticationKeyID", "keyID", "-authenticationKeyIssuerID", "issuerID",
				"CODE_SIGNING_ALLOWED=NO", "-quiet",
			},
		},
		{name: "workspace path with a trailing slash", build: build(BuildParams{ProjectPath: "App.xcworkspace/"}), want: []string{"build", "-workspace", "App.xcworkspace/"}},
		{name: "Swift package manifest gets no container flag", build: build(BuildParams{ProjectPath: "MyPackage/Package.swift", Scheme: "CoolLibrary"}), want: []string{"build", "-scheme", "CoolLibrary"}},
		{name: "Swift package directory gets no container flag", build: build(BuildParams{ProjectPath: "MyPackage/", Scheme: "CoolLibrary"}), want: []string{"build", "-scheme", "CoolLibrary"}},
		{
			name: "analyze",
			build: func() (Command, error) {
				return Analyze(AnalyzeParams{ProjectPath: "App.xcodeproj", Scheme: "App", ResultBundlePath: "/tmp/Analyze.xcresult", DisableCodeSigning: true, AdditionalOptions: []string{"COMPILER_INDEX_STORE_ENABLE=NO"}})
			},
			want: []string{"analyze", "-project", "App.xcodeproj", "-scheme", "App", "-resultBundlePath", "/tmp/Analyze.xcresult", "CODE_SIGNING_ALLOWED=NO", "COMPILER_INDEX_STORE_ENABLE=NO"},
		},
		{
			name: "build-for-testing",
			build: func() (Command, error) {
				return BuildForTesting(BuildForTestingParams{ProjectPath: "App.xcodeproj", Scheme: "App", Configuration: "Debug", Destination: "id=1", TestPlan: "FullTests", Authentication: &testAuth, AdditionalOptions: []string{"SYMROOT=/tmp/test_bundle", "-only-testing:AppTests"}})
			},
			want: []string{
				"build-for-testing", "-project", "App.xcodeproj", "-scheme", "App", "-configuration", "Debug", "-destination", "id=1",
				"-allowProvisioningUpdates", "-authenticationKeyPath", "/key/path", "-authenticationKeyID", "keyID", "-authenticationKeyIssuerID", "issuerID",
				"-testPlan", "FullTests", "SYMROOT=/tmp/test_bundle", "-only-testing:AppTests",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := tt.build()
			require.NoError(t, err)
			require.Equal(t, tt.want, cmd.Args())
		})
	}
}
