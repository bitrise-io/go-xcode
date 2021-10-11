package xcodebuild

import (
	"reflect"
	"strings"
	"testing"
)

func TestCommandBuilder_cmdSlice(t *testing.T) {
	tests := []struct {
		name    string
		builder CommandBuilder
		want    []string
	}{
		{
			name: "simulator",
			builder: CommandBuilder{
				destination: "id=3222ioocsdcsa1",
				action:      BuildAction,
			},
			want: []string{
				"-destination",
				"id=3222ioocsdcsa1",
				"build",
			},
		},
		{
			name: "generic iOS",
			builder: CommandBuilder{
				destination: "generic/platform=iOS",
				action:      BuildAction,
			},
			want: []string{
				"-destination",
				"generic/platform=iOS",
				"build",
			},
		},
		{
			name: "scheme",
			builder: CommandBuilder{
				scheme: "project_scheme",
				action: BuildAction,
			},
			want: []string{
				"-scheme",
				"project_scheme",
				"build",
			},
		},
		{
			name: "iphone simulator sdk",
			builder: CommandBuilder{
				sdk:    "iphonesimulator12",
				action: BuildAction,
			},
			want: []string{
				"build",
				"-sdk",
				"iphonesimulator12",
			},
		},
		{
			name: "analyze",
			builder: CommandBuilder{
				resultBundlePath: "/tmp/Analyze.xcresult",
				action:           AnalyzeAction,
			},
			want: []string{
				"analyze",
				"-resultBundlePath",
				"/tmp/Analyze.xcresult",
			},
		},
		{
			name: "force development team",
			builder: CommandBuilder{
				action: BuildAction,
			},
			want: []string{
				"DEVELOPMENT_TEAM=Bitrise",
				"build",
			},
		},
		{
			name: "force code sign identity",
			builder: CommandBuilder{
				forceCodeSignIdentity: "iOS Developer: Bitrise bot",
				action:                BuildAction,
			},
			want: []string{
				"CODE_SIGN_IDENTITY=iOS Developer: Bitrise bot",
				"build",
			},
		},
		{
			name: "force profile",
			builder: CommandBuilder{
				forceProvisioningProfileSpecifier: "id 2331232",
				action:                            BuildAction,
			},
			want: []string{
				"PROVISIONING_PROFILE_SPECIFIER=id 2331232",
				"PROVISIONING_PROFILE=id 2331232",
				"build",
			},
		},
		{
			name: "project",
			builder: CommandBuilder{
				projectPath: "project.xcodeproj",
				action:      BuildAction,
			},
			want: []string{
				"-project",
				"project.xcodeproj",
				"build",
			},
		},
		{
			name: "workspace",
			builder: CommandBuilder{
				projectPath: "project.xcworkspace",
				isWorkspace: true,
				action:      BuildAction,
			},
			want: []string{
				"-workspace",
				"project.xcworkspace",
				"build",
			},
		},
		{
			name: "debug configuration",
			builder: CommandBuilder{
				configuration: "debug",
				action:        BuildAction,
			},
			want: []string{
				"-configuration",
				"debug",
				"build",
			},
		},
		{
			name: "archive",
			builder: CommandBuilder{
				archivePath: "archive/path",
				action:      ArchiveAction,
			},
			want: []string{
				"archive",
				"-archivePath",
				"archive/path",
			},
		},
		{
			name: "disable index while building",
			builder: CommandBuilder{
				action: BuildAction,
			},
			want: []string{
				"COMPILER_INDEX_STORE_ENABLE=NO",
				"build",
			},
		},
		{
			name: "disable code signing",
			builder: CommandBuilder{
				disableCodesign: true,
				action:          BuildAction,
			},
			want: []string{
				"CODE_SIGNING_ALLOWED=NO",
				"build",
			},
		},
		{
			name: "xcconfig",
			builder: CommandBuilder{
				xcconfigPath: "temp.xcconfig",
				action:       BuildAction,
			},
			want: []string{
				"-xcconfig",
				"temp.xcconfig",
				"build",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.builder.args()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CommandBuilder.cmdSlice() = %v\nwant %v", strings.Join(got, "\n"), strings.Join(tt.want, "\n"))
			}
		})
	}
}
