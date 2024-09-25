package xcodeproj

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	plist "github.com/bitrise-io/go-plist"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
)

func Test_assetCatalog(t *testing.T) {
	var objects serialized.Object
	_, err := plist.Unmarshal([]byte(rawProj), &objects)
	if err != nil {
		t.Fatalf("setup: failed to unmarshal project")
	}
	proj, err := parseProj("BA3CBE6D19F7A93800CED4D5", objects)
	if err != nil {
		t.Fatalf("setup: failed to parse project")
	}

	tests := []struct {
		name    string
		target  Target
		objects serialized.Object
		want    []fileReference
		wantErr bool
	}{
		{
			name:    "good path",
			target:  proj.Targets[0],
			objects: objects,
			want: []fileReference{{
				id:   "BA3CBE8819F7A93900CED4D5",
				path: "Images.xcassets",
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := assetCatalogs(tt.target, proj.ID, tt.objects)
			if (err != nil) != tt.wantErr {
				t.Errorf("AssetCatalogs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AssetCatalogs() = %v, want %v", got, tt.want)
			}
		})
	}
}

type internalProject struct {
	objects serialized.Object
	proj    Proj
}

func createDummyProject(rawProj string, rootObjectID string, projectPath string, iconSetPaths [][]string) (internalProject, error) {
	for _, pathElements := range iconSetPaths {
		path := filepath.Join(append([]string{projectPath}, pathElements...)...)
		if err := os.MkdirAll(path, 0755); err != nil {
			return internalProject{}, fmt.Errorf("failed top create dir %v", err)
		}
	}

	var objects serialized.Object
	_, err := plist.Unmarshal([]byte(rawProj), &objects)
	if err != nil {
		return internalProject{}, fmt.Errorf("failed to unmarshal project, %v", err)
	}
	// PBXProject object ID
	proj, err := parseProj(rootObjectID, objects)
	if err != nil {
		return internalProject{}, fmt.Errorf("failed to parse project, %v", err)
	}

	return internalProject{
		objects: objects,
		proj:    proj,
	}, nil
}

func Test_appIconSetPaths(t *testing.T) {
	tests := []struct {
		name         string
		rawProj      string
		rootObjectID string
		projPath     string
		iconSetPaths [][]string
		want         []string
		wantErr      bool
	}{
		{
			name:         "single asset catlog",
			rawProj:      rawProj,
			rootObjectID: "BA3CBE6D19F7A93800CED4D5",
			projPath:     "ios-simple-objc.xcodeproj",
			iconSetPaths: [][]string{
				{"ios-simple-objc", "Images.xcassets", "AppIcon.appiconset"},
			},
			want: []string{"ios-simple-objc", "Images.xcassets", "AppIcon.appiconset"},
		},
		{
			name:         "asset catalog missing",
			rawProj:      rawProj,
			rootObjectID: "BA3CBE6D19F7A93800CED4D5",
			projPath:     "ios-simple-objc.xcodeproj",
			iconSetPaths: [][]string{},
			want:         []string{},
			wantErr:      true,
		},
		{
			name:         "2 asset catalogs",
			rawProj:      rawCatalystProj,
			rootObjectID: "13917C0A243F43D00087912B",
			projPath:     "Catalyst Sample.xcodeproj",
			iconSetPaths: [][]string{
				{"Catalyst Sample", "Assets.xcassets", "AppIcon.appiconset"},
				{"Catalyst Sample", "Preview Content", "Preview Assets.appiconset"},
			},
			want: []string{"Catalyst Sample", "Assets.xcassets", "AppIcon.appiconset"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectDir, err := ioutil.TempDir("", "ios-dummy-project")
			if err != nil {
				t.Errorf("setup: failed to create temp dir, %v", err)
			}
			defer func() {
				if err := os.RemoveAll(projectDir); err != nil {
					t.Logf("failed to clean up after test, error: %s", err)
				}
			}()
			internalProject, err := createDummyProject(tt.rawProj, tt.rootObjectID, projectDir, tt.iconSetPaths)
			if err != nil {
				t.Errorf("setup: %v", err)
			}
			var want TargetsToAppIconSets
			if len(tt.want) != 0 {
				want = TargetsToAppIconSets{
					internalProject.proj.Targets[0].ID: []string{filepath.Join(append([]string{projectDir}, tt.want...)...)},
				}
			}

			got, err := appIconSetPaths(internalProject.proj, filepath.Join(projectDir, tt.projPath), internalProject.objects)

			if (err != nil) != tt.wantErr {
				t.Errorf("appIconSetPaths() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("appIconSetPaths() = %v, want %v", got, want)
			}
		})
	}
}
