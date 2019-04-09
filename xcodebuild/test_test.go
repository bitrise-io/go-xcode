package xcodebuild

import (
	"reflect"
	"strings"
	"testing"
)

func TestTestCommandModel_cmdSlice(t *testing.T) {
	tests := []struct {
		name                      string
		projectPath               string
		isWorkspace               bool
		scheme                    string
		destination               string
		generateCodeCoverage      bool
		disableIndexWhileBuilding bool
		customBuildActions        []string
		customOptions             []string
		want                      []string
	}{
		{
			name:                      "test simulator",
			projectPath:               "ios/project.xcprojec",
			isWorkspace:               false,
			scheme:                    "project",
			destination:               "id 2323asd2s",
			generateCodeCoverage:      false,
			disableIndexWhileBuilding: true,
			customBuildActions:        []string{""},
			customOptions:             []string{""},
			want: []string{
				"xcodebuild",
				"-project",
				"ios/project.xcprojec",
				"-scheme",
				"project",
				"",
				"test",
				"-destination",
				"id 2323asd2s",
				"COMPILER_INDEX_STORE_ENABLE=NO",
				"",
			},
		},
		{
			name:                      "test generate code coverage",
			projectPath:               "ios/project.xcprojec",
			isWorkspace:               false,
			scheme:                    "project",
			destination:               "id 2323asd2s",
			generateCodeCoverage:      true,
			disableIndexWhileBuilding: true,
			customBuildActions:        []string{""},
			customOptions:             []string{""},
			want: []string{
				"xcodebuild",
				"-project",
				"ios/project.xcprojec",
				"-scheme",
				"project",
				"GCC_INSTRUMENT_PROGRAM_FLOW_ARCS=YES",
				"GCC_GENERATE_TEST_COVERAGE_FILES=YES",
				"",
				"test",
				"-destination",
				"id 2323asd2s",
				"COMPILER_INDEX_STORE_ENABLE=NO",
				"",
			},
		},
		{
			name:                      "test workspace",
			projectPath:               "ios/project.xcworkspaxe",
			isWorkspace:               true,
			scheme:                    "project",
			destination:               "id 2323asd2s",
			generateCodeCoverage:      false,
			disableIndexWhileBuilding: true,
			customBuildActions:        []string{""},
			customOptions:             []string{""},
			want: []string{
				"xcodebuild",
				"-workspace",
				"ios/project.xcworkspaxe",
				"-scheme",
				"project",
				"",
				"test",
				"-destination",
				"id 2323asd2s",
				"COMPILER_INDEX_STORE_ENABLE=NO",
				"",
			},
		},
		{
			name:                      "test generic iOS",
			projectPath:               "ios/project.xcproject",
			isWorkspace:               false,
			scheme:                    "project",
			destination:               "generic/platform=iOS",
			generateCodeCoverage:      false,
			disableIndexWhileBuilding: true,
			customBuildActions:        []string{""},
			customOptions:             []string{""},
			want: []string{
				"xcodebuild",
				"-project",
				"ios/project.xcproject",
				"-scheme",
				"project",
				"",
				"test",
				"-destination",
				"generic/platform=iOS",
				"COMPILER_INDEX_STORE_ENABLE=NO",
				"",
			},
		},
		{
			name:                      "test generic iOS index while building",
			projectPath:               "ios/project.xcprojec",
			isWorkspace:               false,
			scheme:                    "project",
			destination:               "id 2323asd2s",
			generateCodeCoverage:      false,
			disableIndexWhileBuilding: false,
			customBuildActions:        []string{""},
			customOptions:             []string{""},
			want: []string{
				"xcodebuild",
				"-project",
				"ios/project.xcprojec",
				"-scheme",
				"project",
				"",
				"test",
				"-destination",
				"id 2323asd2s",
				"",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &TestCommandModel{
				projectPath:               tt.projectPath,
				isWorkspace:               tt.isWorkspace,
				scheme:                    tt.scheme,
				destination:               tt.destination,
				generateCodeCoverage:      tt.generateCodeCoverage,
				disableIndexWhileBuilding: tt.disableIndexWhileBuilding,
				customBuildActions:        tt.customBuildActions,
				customOptions:             tt.customOptions,
			}
			if got := c.cmdSlice(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TestCommandModel.cmdSlice() = %v\nwant %v", strings.Join(got, "\n"), strings.Join(tt.want, "\n"))
			}
		})
	}
}
