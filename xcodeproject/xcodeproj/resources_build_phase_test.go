package xcodeproj

import (
	"path"
	"reflect"
	"testing"

	plist "github.com/bitrise-io/go-plist"
	"github.com/bitrise-io/go-xcode/xcodeproject/serialized"
)

func Test_parseResourcesBuildPhase(t *testing.T) {
	var raw serialized.Object
	_, err := plist.Unmarshal([]byte(rawResourcesBuildPhase), &raw)
	if err != nil {
		t.Errorf("setup: failed to parse raw object")
	}

	const id1 = "47C11A3D21FF63950084FD7F"

	tests := []struct {
		name string

		id      string
		objects serialized.Object

		want    resourcesBuildPhase
		wantErr bool
	}{
		{
			name:    "normal",
			id:      id1,
			objects: raw,
			want: resourcesBuildPhase{
				ID:    id1,
				files: []string{"47C11A4D21FF63970084FD7F", "47C11A4A21FF63970084FD7F", "47C11A4821FF63950084FD7F"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseResourcesBuildPhase(tt.id, tt.objects)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseResourcesBuildPhase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseResourcesBuildPhase() = %v, want %v", got, tt.want)
			}
		})
	}
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
	var raw serialized.Object
	_, err := plist.Unmarshal([]byte(rawFileReference), &raw)
	if err != nil {
		t.Errorf("setup: failed to parse raw object")
	}

	tests := []struct {
		name string

		id      string
		objects serialized.Object

		want    fileReference
		wantErr bool
	}{
		{
			name:    "Normal case",
			id:      "47C11A4921FF63970084FD7F",
			objects: raw,
			want: fileReference{
				id:   "47C11A4921FF63970084FD7F",
				path: "Assets.xcassets",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFileReference(tt.id, tt.objects)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFileReference() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseFileReference() = %v, want %v", got, tt.want)
			}
		})
	}
}

const rawFileReference = `
47C11A4921FF63970084FD7F /* Assets.xcassets */ = {isa = PBXFileReference; lastKnownFileType = folder.assetcatalog; path = Assets.xcassets; sourceTree = "<group>"; };
`

func Test_resolveFileReferenceAbsolutePath(t *testing.T) {
	var objects serialized.Object
	_, err := plist.Unmarshal([]byte(rawProj), &objects)
	if err != nil {
		t.Fatalf("setup: failed to unmarshal project")
	}
	const projectID = "BA3CBE6D19F7A93800CED4D5"

	type args struct {
		id          string
		projectID   string
		projectPath string
		objects     serialized.Object
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "happy case",
			args: args{
				id:          "BA3CBE8819F7A93900CED4D5",
				projectID:   projectID,
				projectPath: path.Join("parent", "project_root"),
				objects:     objects,
			},
			want:    path.Join("parent", "ios-simple-objc", "Images.xcassets"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveObjectAbsolutePath(tt.args.id, tt.args.projectID, tt.args.projectPath, tt.args.objects)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveFileReferenceAbsolutePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("resolveFileReferenceAbsolutePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resolveFilePath(t *testing.T) {
	tests := []struct {
		name    string
		nodes   []projectEntry
		want    string
		wantErr bool
	}{
		{
			name: "simple",
			nodes: []projectEntry{
				{
					path:         "Images.xcassets",
					pathRelation: groupParent,
				},
				{
					path:         "project_root",
					pathRelation: absoluteParentPath,
				},
			},
			want:    path.Join("project_root", "Images.xcassets"),
			wantErr: false,
		},
		{
			name: "1 level with group",
			nodes: []projectEntry{
				{
					path:         "Images.xcassets",
					pathRelation: groupParent,
				},
				{
					path:         "",
					pathRelation: groupParent,
				},
				{
					path:         "project_root",
					pathRelation: absoluteParentPath,
				},
			},
			want:    path.Join("project_root", "Images.xcassets"),
			wantErr: false,
		},
		{
			name: "2 level with group root",
			nodes: []projectEntry{
				{
					path:         "Images.xcassets",
					pathRelation: groupParent,
				},
				{
					path:         "",
					pathRelation: groupParent,
				},
				{
					path:         "project_root",
					pathRelation: groupParent,
				},
			},
			want:    path.Join("project_root", "Images.xcassets"),
			wantErr: false,
		},
		{
			name: "3 levels",
			nodes: []projectEntry{
				{
					path:         "Images.xcassets",
					pathRelation: groupParent,
				},
				{
					path:         "group",
					pathRelation: groupParent,
				},
				{
					path:         "project_root",
					pathRelation: absoluteParentPath,
				},
			},
			want:    path.Join("project_root", "group", "Images.xcassets"),
			wantErr: false,
		},
		{
			name: "3 levels with absolute group",
			nodes: []projectEntry{
				{
					path:         "Images.xcassets",
					pathRelation: groupParent,
				},
				{
					path:         "group",
					pathRelation: absoluteParentPath,
				},
				{
					path:         "project_root",
					pathRelation: absoluteParentPath,
				},
			},
			want:    path.Join("group", "Images.xcassets"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveFilePath(tt.nodes)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveFilePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("resolveFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

//
// BA3CBE8819F7A93900CED4D5 /* Images.xcassets */ = {isa = PBXFileReference; lastKnownFileType = folder.assetcatalog; path = Images.xcassets; sourceTree = "<group>"; };
//
// BA3CBE7719F7A93800CED4D5 /* ios-simple-objc */ = {
// 	isa = PBXGroup;
// 	children = (
// 		BA3CBE7C19F7A93800CED4D5 /* AppDelegate.h */,
// 		BA3CBE7D19F7A93900CED4D5 /* AppDelegate.m */,
// 		BA3CBE8219F7A93900CED4D5 /* ViewController.h */,
// 		BA3CBE8319F7A93900CED4D5 /* ViewController.m */,
// 		BA3CBE8519F7A93900CED4D5 /* Main.storyboard */,
// 		BA3CBE8819F7A93900CED4D5 /* Images.xcassets */,
// 		BA3CBE8A19F7A93900CED4D5 /* LaunchScreen.xib */,
// 		BA3CBE7F19F7A93900CED4D5 /* ios_simple_objc.xcdatamodeld */,
// 		BA3CBE7819F7A93800CED4D5 /* Supporting Files */,
// 	);
// 	path = "ios-simple-objc";
// 	sourceTree = "<group>";
// };
//
// BA3CBE6C19F7A93800CED4D5 = {
// 	isa = PBXGroup;
// 	children = (
// 		BA3CBE7719F7A93800CED4D5 /* ios-simple-objc */,
// 		BA3CBE9419F7A93900CED4D5 /* ios-simple-objcTests */,
// 		BA3CBE7619F7A93800CED4D5 /* Products */,
// 	);
// 	sourceTree = "<group>";
// };
//
// mainGroup = BA3CBE6C19F7A93800CED4D5;
// projectDirPath = "";
// projectRoot = "";

func Test_findInProjectTree(t *testing.T) {
	var objects serialized.Object
	_, err := plist.Unmarshal([]byte(rawProj), &objects)
	if err != nil {
		t.Errorf("setup: failed to parse raw object")
	}

	projectRoot, err := objects.Object("BA3CBE6D19F7A93800CED4D5")
	if err != nil {
		t.Fatalf("setup failed: project root not found")
	}
	mainGroupID, err := projectRoot.String("mainGroup")
	if err != nil {
		t.Fatalf("setup failed: main group not found")
	}

	type args struct {
		target    string
		currentID string
		object    serialized.Object
		visited   *[]string
	}
	tests := []struct {
		name    string
		args    args
		want    []projectEntry
		wantErr bool
	}{
		{
			name: "happy case",
			args: args{
				target:    "BA3CBE8819F7A93900CED4D5",
				currentID: mainGroupID,
				object:    objects,
				visited:   &[]string{},
			},
			want: []projectEntry{
				{
					id:           "BA3CBE8819F7A93900CED4D5",
					path:         "Images.xcassets",
					pathRelation: groupParent,
				},
				{
					id:           "BA3CBE7719F7A93800CED4D5",
					path:         "ios-simple-objc",
					pathRelation: groupParent,
				},
				{
					id:           "BA3CBE6C19F7A93800CED4D5",
					path:         "",
					pathRelation: groupParent,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findInProjectTree(tt.args.target, tt.args.currentID, tt.args.object, tt.args.visited)
			if (err != nil) != tt.wantErr {
				t.Errorf("findInProjectTree() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findInProjectTree() = %v+, want %v", got, tt.want)
			}
		})
	}
}
