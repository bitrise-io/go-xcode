package xcodeproj

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testdata/ios-sample-asset-catalogs holds Assets.xcassets in the app's synchronized folder and
// Branding/Icons/Icons.xcassets in nested groups.
func TestXCProj_AppIconSetPaths_xcodeOutput(t *testing.T) {
	projectDir, err := filepath.Abs(filepath.Join("testdata", "ios-sample-asset-catalogs"))
	require.NoError(t, err)
	projectPath := filepath.Join(projectDir, "sample-apps-ios-json-xcproj.xcodeproj")

	project, err := testFactory().Open(projectPath)
	require.NoError(t, err)

	got, err := project.AppIconSetPaths()
	require.NoError(t, err)
	assert.Equal(t, map[string][]string{
		xcprojAppTargetID: {filepath.Join(projectDir, "sample-apps-ios-json-xcproj", "Assets.xcassets", "AppIcon.appiconset")},
	}, got)

	content, err := os.ReadFile(filepath.Join(projectPath, xcProjFileName))
	require.NoError(t, err)
	content = []byte(strings.Replace(string(content), `"ASSETCATALOG_COMPILER_APPICON_NAME": "AppIcon"`, `"ASSETCATALOG_COMPILER_APPICON_NAME": "AlternateIcon"`, 1))
	alternateIconProject, err := testFactory().newXCProj(content, projectPath)
	require.NoError(t, err)

	got, err = alternateIconProject.AppIconSetPaths()
	require.NoError(t, err)
	assert.Equal(t, map[string][]string{
		xcprojAppTargetID: {filepath.Join(projectDir, "Branding", "Icons", "Icons.xcassets", "AlternateIcon.appiconset")},
	}, got)
}

func TestXCProj_AppIconSetPaths(t *testing.T) {
	tests := []struct {
		name string
		// files may hold {{projectDir}}, which is replaced with the project directory.
		files        string
		iconSetPaths []string
		// symlinks maps a link path to its target, both relative to the project directory.
		symlinks map[string]string
		want     []string
		wantErr  string
	}{
		{
			name:         "root-level file reference with a multi-part path",
			files:        `[{"path": "Branding/Icons/Icons.xcassets", "target-membership": ["App/resources"]}]`,
			iconSetPaths: []string{"Branding/Icons/Icons.xcassets/AppIcon.appiconset"},
			want:         []string{"Branding/Icons/Icons.xcassets/AppIcon.appiconset"},
		},
		{
			name: "absolute path inside a group",
			files: `[{"kind": "group", "path": "Branding", "children": [
				{"path": "{{projectDir}}/Icons.xcassets", "target-membership": ["App/resources"]},
			]}]`,
			iconSetPaths: []string{"Icons.xcassets/AppIcon.appiconset"},
			want:         []string{"Icons.xcassets/AppIcon.appiconset"},
		},
		{
			name: "project path base inside a group",
			files: `[{"kind": "group", "path": "Branding", "children": [
				{"path": "<PROJECT>/Icons.xcassets", "target-membership": ["App/resources"]},
			]}]`,
			iconSetPaths: []string{"Icons.xcassets/AppIcon.appiconset"},
			want:         []string{"Icons.xcassets/AppIcon.appiconset"},
		},
		{
			name: "group at the project root",
			files: `[{"kind": "group", "path": "Branding", "children": [
				{"kind": "group", "path": "<PROJECT>/", "children": [
					{"path": "Icons.xcassets", "target-membership": ["App/resources"]},
				]},
			]}]`,
			iconSetPaths: []string{"Icons.xcassets/AppIcon.appiconset"},
			want:         []string{"Icons.xcassets/AppIcon.appiconset"},
		},
		{
			name:         "escaped path",
			files:        `[{"path": "\\<Icons\\\\.xcassets", "target-membership": ["App/resources"]}]`,
			iconSetPaths: []string{`<Icons\.xcassets/AppIcon.appiconset`},
			want:         []string{`<Icons\.xcassets/AppIcon.appiconset`},
		},
		{
			name:         "asset catalog nested in a synchronized folder",
			files:        `[{"kind": "folder", "path": "App", "target-membership": ["App"]}]`,
			iconSetPaths: []string{"App/Resources/Assets.xcassets/AppIcon.appiconset"},
			want:         []string{"App/Resources/Assets.xcassets/AppIcon.appiconset"},
		},
		{
			name: "asset catalogs of other targets and build phases",
			files: `[
				{"path": "Copied.xcassets", "target-membership": ["App/copy"]},
				{"path": "Other.xcassets", "target-membership": ["Other/resources"]},
				{"kind": "folder", "path": "OtherFolder", "target-membership": ["Other"]},
			]`,
			iconSetPaths: []string{
				"Copied.xcassets/AppIcon.appiconset",
				"Other.xcassets/AppIcon.appiconset",
				"OtherFolder/Assets.xcassets/AppIcon.appiconset",
			},
		},
		{
			name: "path base that needs build settings",
			files: `[{"kind": "group", "path": "<USER:BRANDING_DIR>/Branding", "children": [
				{"path": "Icons.xcassets", "target-membership": ["App/resources"]},
			]}]`,
			wantErr: "failed to resolve path <USER:BRANDING_DIR>/Branding: unsupported path base",
		},
		{
			name:    "home-relative synchronized folder",
			files:   `[{"kind": "folder", "path": "~/App", "target-membership": ["App"]}]`,
			wantErr: "failed to resolve path ~/App: unsupported path base",
		},
		{
			name: "unresolvable path that leads to no catalog of the target",
			files: `[
				{"kind": "group", "path": "<SDK>/", "children": [
					{"path": "Other.xcassets", "target-membership": ["Other/resources"]},
					{"path": "Info.plist", "target-membership": ["App/resources"]},
				]},
				{"kind": "group", "path": "<USER:BRANDING_DIR>/", "children": [
					{"path": "<PROJECT>/Icons.xcassets", "target-membership": ["App/resources"]},
				]},
			]`,
			iconSetPaths: []string{"Icons.xcassets/AppIcon.appiconset"},
			want:         []string{"Icons.xcassets/AppIcon.appiconset"},
		},
		{
			name:         "non-native target",
			files:        `[{"kind": "folder", "path": "Aggregate", "target-membership": ["Aggregate"]}]`,
			iconSetPaths: []string{"Aggregate/Assets.xcassets/AppIcon.appiconset"},
		},
		{
			name:         "app icon set missing",
			files:        `[{"kind": "folder", "path": "App", "target-membership": ["App"]}]`,
			iconSetPaths: []string{"App/Assets.xcassets/Other.appiconset"},
			wantErr:      "not found app icon set (AppIcon)",
		},
		{
			name:  "symbolic links in a synchronized folder",
			files: `[{"kind": "folder", "path": "App", "target-membership": ["App"]}]`,
			iconSetPaths: []string{
				"App/Assets.xcassets/AppIcon.appiconset",
				"Linked/Linked.xcassets/AppIcon.appiconset",
			},
			symlinks: map[string]string{
				"App/Nested/Cycle":    "App",
				"App/Linked.xcassets": "Linked/Linked.xcassets",
			},
			want: []string{"App/Assets.xcassets/AppIcon.appiconset"},
		},
		{
			name: "synchronized folder missing",
			files: `[
				{"kind": "folder", "path": "App", "target-membership": ["App"]},
				{"path": "Icons.xcassets", "target-membership": ["App/resources"]},
			]`,
			iconSetPaths: []string{"Icons.xcassets/AppIcon.appiconset"},
			want:         []string{"Icons.xcassets/AppIcon.appiconset"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectPath := writeProjectFiles(t, nil)
			projectDir := filepath.Dir(projectPath)
			content := fmt.Appendf(nil, `{
				"configurations": ["Debug"],
				"targets": [
					{"name": "App", "id": "A1", "build-settings": {"ASSETCATALOG_COMPILER_APPICON_NAME": "AppIcon"}},
					{"name": "Other", "id": "O1"},
					{"name": "Aggregate", "id": "G1", "kind": "aggregate", "build-settings": {"ASSETCATALOG_COMPILER_APPICON_NAME": "AppIcon"}},
				],
				"files": %s,
			}`, strings.ReplaceAll(tt.files, "{{projectDir}}", projectDir))
			require.NoError(t, os.WriteFile(filepath.Join(projectPath, xcProjFileName), content, 0644))
			for _, iconSetPath := range tt.iconSetPaths {
				require.NoError(t, os.MkdirAll(filepath.Join(projectDir, iconSetPath), 0755))
			}
			for link, target := range tt.symlinks {
				linkPath := filepath.Join(projectDir, link)
				require.NoError(t, os.MkdirAll(filepath.Dir(linkPath), 0755))
				require.NoError(t, os.Symlink(filepath.Join(projectDir, target), linkPath))
			}

			project, err := testFactory().Open(projectPath)
			require.NoError(t, err)

			got, err := project.AppIconSetPaths()
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			want := map[string][]string{}
			for _, iconSetPath := range tt.want {
				want["A1"] = append(want["A1"], filepath.Join(projectDir, iconSetPath))
			}
			assert.Equal(t, want, got)
		})
	}
}
