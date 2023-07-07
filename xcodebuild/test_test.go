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
		scheme                    string
		destination               string
		generateCodeCoverage      bool
		disableIndexWhileBuilding bool
		customBuildActions        []string
		customOptions             []string
		want                      []string
	}{
		{
			name:                      "simulator",
			projectPath:               "",
			scheme:                    "",
			destination:               "id 2323asd2s",
			generateCodeCoverage:      false,
			disableIndexWhileBuilding: false,
			customBuildActions:        []string{""},
			customOptions:             nil,
			want: []string{
				"xcodebuild",
				"",
				"test",
				"-destination",
				"id 2323asd2s",
			},
		},
		{
			name:                      "generic iOS",
			projectPath:               "",
			scheme:                    "",
			destination:               "generic/platform=iOS",
			generateCodeCoverage:      false,
			disableIndexWhileBuilding: false,
			customBuildActions:        []string{""},
			customOptions:             nil,
			want: []string{
				"xcodebuild",
				"",
				"test",
				"-destination",
				"generic/platform=iOS",
			},
		},
		{
			name:                      "scheme",
			projectPath:               "",
			scheme:                    "ios_scheme",
			destination:               "",
			generateCodeCoverage:      false,
			disableIndexWhileBuilding: false,
			customBuildActions:        []string{""},
			customOptions:             nil,
			want: []string{
				"xcodebuild",
				"-scheme",
				"ios_scheme",
				"",
				"test",
			},
		},
		{
			name:                      "generate code coverage",
			projectPath:               "",
			scheme:                    "",
			destination:               "",
			generateCodeCoverage:      true,
			disableIndexWhileBuilding: false,
			customBuildActions:        []string{""},
			customOptions:             nil,
			want: []string{
				"xcodebuild",
				"GCC_INSTRUMENT_PROGRAM_FLOW_ARCS=YES",
				"GCC_GENERATE_TEST_COVERAGE_FILES=YES",
				"",
				"test",
			},
		},
		{
			name:                      "workspace",
			projectPath:               "ios/project.xcworkspace",
			scheme:                    "",
			destination:               "",
			generateCodeCoverage:      false,
			disableIndexWhileBuilding: false,
			customBuildActions:        []string{""},
			customOptions:             nil,
			want: []string{
				"xcodebuild",
				"-workspace",
				"ios/project.xcworkspace",
				"",
				"test",
			},
		},
		{
			name:                      "project",
			projectPath:               "ios/project.xcodeproj",
			scheme:                    "",
			destination:               "",
			generateCodeCoverage:      false,
			disableIndexWhileBuilding: false,
			customBuildActions:        []string{""},
			customOptions:             nil,
			want: []string{
				"xcodebuild",
				"-project",
				"ios/project.xcodeproj",
				"",
				"test",
			},
		},
		{
			name:                      "disable index while building",
			projectPath:               "",
			scheme:                    "",
			destination:               "",
			generateCodeCoverage:      false,
			disableIndexWhileBuilding: true,
			customBuildActions:        []string{""},
			customOptions:             nil,
			want: []string{
				"xcodebuild",
				"",
				"test",
				"COMPILER_INDEX_STORE_ENABLE=NO",
			},
		},
		{
			name:                      "SPM project",
			projectPath:               "Package.swift",
			scheme:                    "CoolLibrary",
			destination:               "",
			generateCodeCoverage:      false,
			disableIndexWhileBuilding: false,
			customBuildActions:        []string{""},
			customOptions:             nil,
			want: []string{
				"xcodebuild",
				"-scheme",
				"CoolLibrary",
				"",
				"test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &TestCommandModel{
				projectPath:               tt.projectPath,
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
