package xcodeproj

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func pbxFile(t *testing.T, p *Project) *pbxprojFile {
	t.Helper()

	file, ok := projectFileOf(p).(*pbxprojFile)
	require.True(t, ok, "%s is not a project.pbxproj project", p.Path())

	return file
}

func TestProject_perObjectModify(t *testing.T) {
	tests := []struct {
		name          string
		fixture       string
		target        string
		configuration string
		want          string
	}{
		{
			name:          "no target attributes",
			fixture:       "without-target-attributes.pbxproj",
			target:        "TargetWithouthTargetAttributes",
			configuration: "Debug",
			want:          "without-target-attributes-modified.pbxproj",
		},
		{
			name:          "changes 2 objects, as target attributes are included in the project",
			fixture:       "ios-sample.pbxproj",
			target:        "XcodeProj",
			configuration: "Debug",
			want:          "ios-sample-changed.pbxproj",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project := parseFixture(t, tt.fixture)
			require.NoError(t, project.ForceCodeSign(forceCodeSignOptions(tt.target, tt.configuration)))

			got, err := pbxFile(t, project).perObjectModify()
			require.NoError(t, err)

			want, err := os.ReadFile(filepath.Join("testdata", tt.want))
			require.NoError(t, err)
			require.Equal(t, string(want), string(got))
		})
	}
}

func TestProject_perObjectModify_noChangesReturnsOriginalBytes(t *testing.T) {
	project := parseFixture(t, "ios-sample.pbxproj")

	got, err := pbxFile(t, project).perObjectModify()
	require.NoError(t, err)

	want, err := os.ReadFile(filepath.Join("testdata", "ios-sample.pbxproj"))
	require.NoError(t, err)
	require.Equal(t, string(want), string(got))
}

func openPBXProj(t *testing.T, projectPath string) *Project {
	t.Helper()

	opened, err := testFactory().Open(projectPath)
	require.NoError(t, err)

	p := opened
	pbxFile(t, p)

	return p
}

func TestProject_Save_fallsBackToFullRewrite(t *testing.T) {
	projectPath := filepath.Join(t.TempDir(), "App.xcodeproj")
	require.NoError(t, os.MkdirAll(projectPath, 0755))

	content, err := os.ReadFile(filepath.Join("testdata", "minimal.pbxproj"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "project.pbxproj"), content, 0644))

	project := openPBXProj(t, projectPath)

	objects, ok := pbxFile(t, project).rawProj.Object("objects")
	require.True(t, ok)
	objects["BB00000000000000000000A1"] = map[string]any{"isa": "PBXGroup", "children": []any{}, "sourceTree": "<group>"}

	_, err = pbxFile(t, project).perObjectModify()
	require.ErrorContains(t, err, "new object added")

	require.NoError(t, project.Save())

	reopened := openPBXProj(t, projectPath)
	require.Len(t, reopened.Targets(), 1)

	reopenedObjects, ok := pbxFile(t, reopened).rawProj.Object("objects")
	require.True(t, ok)
	assert.True(t, reopenedObjects.Has("BB00000000000000000000A1"))
}
