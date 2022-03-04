package xcodebuild

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolvePackagesCommandModel_cmdSlice(t *testing.T) {
	type fields struct {
		projectPath   string
		scheme        string
		configuration string
		customOptions []string
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "workspace",
			fields: fields{
				projectPath: "test.xcworkspace",
			},
			want: []string{
				"xcodebuild",
				"-workspace", "test.xcworkspace",
				"-resolvePackageDependencies",
			},
		},
		{
			name: "project",
			fields: fields{
				projectPath:   "test.xcodeproj",
				scheme:        "Test",
				configuration: "Debug",
			},
			want: []string{
				"xcodebuild",
				"-project", "test.xcodeproj",
				"-scheme", "Test",
				"-configuration", "Debug",
				"-resolvePackageDependencies",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ResolvePackagesCommandModel{
				projectPath:   tt.fields.projectPath,
				scheme:        tt.fields.scheme,
				configuration: tt.fields.configuration,
				customOptions: tt.fields.customOptions,
			}

			got := m.cmdSlice()

			require.Equal(t, got, tt.want)
		})
	}
}
