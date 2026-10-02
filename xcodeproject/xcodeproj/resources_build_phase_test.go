package xcodeproj

import (
	"path"
	"testing"

	"github.com/bitrise-io/go-plist"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func unmarshalObjects(t *testing.T, raw string) serialized.Object {
	t.Helper()
	var objects serialized.Object
	_, err := plist.Unmarshal([]byte(raw), &objects)
	require.NoError(t, err)
	return objects
}

func Test_parseResourcesBuildPhase(t *testing.T) {
	const id = "47C11A3D21FF63950084FD7F"

	got, err := parseResourcesBuildPhase(id, unmarshalObjects(t, rawResourcesBuildPhase))
	require.NoError(t, err)
	assert.Equal(t, resourcesBuildPhase{
		ID:    id,
		files: []string{"47C11A4D21FF63970084FD7F", "47C11A4A21FF63970084FD7F", "47C11A4821FF63950084FD7F"},
	}, got)
}

const rawResourcesBuildPhase = `
/* Begin PBXResourcesBuildPhase section */
		47C11A3D21FF63950084FD7F /* Resources */ = {
			isa = PBXResourcesBuildPhase;
			buildActionMask = 2147483647;
			files = (
				47C11A4D21FF63970084FD7F /* LaunchScreen.storyboard in Resources */,
				47C11A4A21FF63970084FD7F /* Assets.xcassets in Resources */,
				47C11A4821FF63950084FD7F /* Main.storyboard in Resources */,
			);
			runOnlyForDeploymentPostprocessing = 0;
		};
		47F01785221C4C1E00DF0B8B /* Resources */ = {
			isa = PBXResourcesBuildPhase;
			buildActionMask = 2147483647;
			files = (
				47F0178F221C4C1E00DF0B8B /* MainInterface.storyboard in Resources */,
			);
			runOnlyForDeploymentPostprocessing = 0;
		};
/* End PBXResourcesBuildPhase section */
`

func Test_parseFileReference(t *testing.T) {
	got, err := parseFileReference("47C11A4921FF63970084FD7F", unmarshalObjects(t, rawFileReference))
	require.NoError(t, err)
	assert.Equal(t, fileReference{id: "47C11A4921FF63970084FD7F", path: "Assets.xcassets"}, got)
}

const rawFileReference = `
47C11A4921FF63970084FD7F /* Assets.xcassets */ = {isa = PBXFileReference; lastKnownFileType = folder.assetcatalog; path = Assets.xcassets; sourceTree = "<group>"; };
`

func Test_resolveObjectAbsolutePath(t *testing.T) {
	project := parseFixture(t, "ios-simple-objc.pbxproj")

	got, err := resolveObjectAbsolutePath("BA3CBE8819F7A93900CED4D5", project.projectID, path.Join("parent", "project_root"), projectObjects(t, project))
	require.NoError(t, err)
	assert.Equal(t, path.Join("parent", "ios-simple-objc", "Images.xcassets"), got)
}

func Test_resolveFilePath(t *testing.T) {
	tests := []struct {
		name  string
		nodes []projectEntry
		want  string
	}{
		{
			name: "simple",
			nodes: []projectEntry{
				{path: "Images.xcassets", pathRelation: groupParent},
				{path: "project_root", pathRelation: absoluteParentPath},
			},
			want: path.Join("project_root", "Images.xcassets"),
		},
		{
			name: "1 level with group",
			nodes: []projectEntry{
				{path: "Images.xcassets", pathRelation: groupParent},
				{path: "", pathRelation: groupParent},
				{path: "project_root", pathRelation: absoluteParentPath},
			},
			want: path.Join("project_root", "Images.xcassets"),
		},
		{
			name: "2 level with group root",
			nodes: []projectEntry{
				{path: "Images.xcassets", pathRelation: groupParent},
				{path: "", pathRelation: groupParent},
				{path: "project_root", pathRelation: groupParent},
			},
			want: path.Join("project_root", "Images.xcassets"),
		},
		{
			name: "3 levels",
			nodes: []projectEntry{
				{path: "Images.xcassets", pathRelation: groupParent},
				{path: "group", pathRelation: groupParent},
				{path: "project_root", pathRelation: absoluteParentPath},
			},
			want: path.Join("project_root", "group", "Images.xcassets"),
		},
		{
			name: "3 levels with absolute group",
			nodes: []projectEntry{
				{path: "Images.xcassets", pathRelation: groupParent},
				{path: "group", pathRelation: absoluteParentPath},
				{path: "project_root", pathRelation: absoluteParentPath},
			},
			want: path.Join("group", "Images.xcassets"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveFilePath(tt.nodes)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_findInProjectTree(t *testing.T) {
	project := parseFixture(t, "ios-simple-objc.pbxproj")
	objects := projectObjects(t, project)

	rawProject, ok := objects.Object(project.projectID)
	require.True(t, ok)
	mainGroupID, ok := rawProject.String("mainGroup")
	require.True(t, ok)

	got, err := findInProjectTree("BA3CBE8819F7A93900CED4D5", mainGroupID, objects, map[string]bool{})
	require.NoError(t, err)
	assert.Equal(t, []projectEntry{
		{id: "BA3CBE8819F7A93900CED4D5", path: "Images.xcassets", pathRelation: groupParent},
		{id: "BA3CBE7719F7A93800CED4D5", path: "ios-simple-objc", pathRelation: groupParent},
		{id: "BA3CBE6C19F7A93800CED4D5", path: "", pathRelation: groupParent},
	}, got)
}
