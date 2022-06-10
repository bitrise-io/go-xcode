package xcodebuild

import (
	"reflect"
	"strings"
	"testing"
)

func TestCommandBuilder_cmdSlice(t *testing.T) {
	tests := []struct {
		name    string
		builder func() *CommandBuilder
		want    []string
	}{
		{
			name: "Set destination on build action",
			builder: func() (cmdBuilder *CommandBuilder) {
				cmdBuilder = NewCommandBuilder("", "build")
				cmdBuilder.SetDestination("id=3222ioocsdcsa1")
				return
			},
			want: []string{
				"xcodebuild",
				"build",
				"-destination",
				"id=3222ioocsdcsa1",
			},
		},
		{
			name: "Set scheme on build action",
			builder: func() (cmdBuilder *CommandBuilder) {
				cmdBuilder = NewCommandBuilder("", "build")
				cmdBuilder.SetScheme("project_scheme")
				return
			},
			want: []string{
				"xcodebuild",
				"build",
				"-scheme",
				"project_scheme",
			},
		},
		{
			name: "Set SDK on build action",
			builder: func() (cmdBuilder *CommandBuilder) {
				cmdBuilder = NewCommandBuilder("", "build")
				cmdBuilder.SetSDK("iphonesimulator12")
				return
			},
			want: []string{
				"xcodebuild",
				"build",
				"-sdk",
				"iphonesimulator12",
			},
		},
		{
			name: "Set result bundle path on analyse action",
			builder: func() (cmdBuilder *CommandBuilder) {
				cmdBuilder = NewCommandBuilder("", "analyze")
				cmdBuilder.SetResultBundlePath("/tmp/Analyze.xcresult")
				return
			},
			want: []string{
				"xcodebuild",
				"analyze",
				"-resultBundlePath",
				"/tmp/Analyze.xcresult",
			},
		},
		{
			name: "Set project path on build action",
			builder: func() (cmdBuilder *CommandBuilder) {
				cmdBuilder = NewCommandBuilder("project.xcodeproj", "build")
				return
			},
			want: []string{
				"xcodebuild",
				"build",
				"-project",
				"project.xcodeproj",
			},
		},
		{
			name: "Set workspace path on build action",
			builder: func() (cmdBuilder *CommandBuilder) {
				cmdBuilder = NewCommandBuilder("project.xcworkspace", "build")
				return
			},
			want: []string{
				"xcodebuild",
				"build",
				"-workspace",
				"project.xcworkspace",
			},
		},
		{
			name: "Set configuration on build action",
			builder: func() (cmdBuilder *CommandBuilder) {
				cmdBuilder = NewCommandBuilder("", "build")
				cmdBuilder.SetConfiguration("Debug")
				return
			},
			want: []string{
				"xcodebuild",
				"build",
				"-configuration",
				"Debug",
			},
		},
		{
			name: "Set archive path on archive action",
			builder: func() (cmdBuilder *CommandBuilder) {
				cmdBuilder = NewCommandBuilder("", "archive")
				cmdBuilder.SetArchivePath("archive/path")
				return
			},
			want: []string{
				"xcodebuild",
				"archive",
				"-archivePath",
				"archive/path",
			},
		},
		{
			name: "Set authentication on archive action",
			builder: func() (cmdBuilder *CommandBuilder) {
				cmdBuilder = NewCommandBuilder("", "archive")
				cmdBuilder.SetAuthentication(AuthenticationParams{
					KeyID:     "keyID",
					IsssuerID: "issuerID",
					KeyPath:   "/key/path",
				})
				return
			},
			want: []string{
				"xcodebuild",
				"archive",
				"-allowProvisioningUpdates",
				"-authenticationKeyPath", "/key/path",
				"-authenticationKeyID", "keyID",
				"-authenticationKeyIssuerID", "issuerID",
			},
		},
		{
			name: "Disable code signing on build action",
			builder: func() (cmdBuilder *CommandBuilder) {
				cmdBuilder = NewCommandBuilder("", "build")
				cmdBuilder.SetDisableCodesign(true)
				return
			},
			want: []string{
				"xcodebuild",
				"build",
				"CODE_SIGNING_ALLOWED=NO",
			},
		},
		{
			name: "Set xcconfig on build action",
			builder: func() (cmdBuilder *CommandBuilder) {
				cmdBuilder = NewCommandBuilder("", "build")
				cmdBuilder.SetXCConfigPath("temp.xcconfig")
				return
			},
			want: []string{
				"xcodebuild",
				"build",
				"-xcconfig",
				"temp.xcconfig",
			},
		},
		{
			name: "Set test plan on build-for-testing action",
			builder: func() (cmdBuilder *CommandBuilder) {
				cmdBuilder = NewCommandBuilder("", "build-for-testing")
				cmdBuilder.SetTestPlan("FullTests")
				return
			},
			want: []string{
				"xcodebuild",
				"build-for-testing",
				"-testPlan",
				"FullTests",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.builder().cmdSlice()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CommandBuilder.cmdSlice() = %v\nwant %v", strings.Join(got, "\n"), strings.Join(tt.want, "\n"))
			}
		})
	}
}
