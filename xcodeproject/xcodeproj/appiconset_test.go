package xcodeproj

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// catalyst-objects.plist holds only the objects dictionary of the CatalystSample project.
func catalystPBXProj(t *testing.T) []byte {
	t.Helper()
	objects, err := os.ReadFile(filepath.Join("testdata", "catalyst-objects.plist"))
	require.NoError(t, err)
	return []byte("// !$*UTF8*$!\n{\n\tobjects = " + string(objects) + ";\n\trootObject = 13917C0A243F43D00087912B;\n}\n")
}

func TestProject_AppIconSetPaths(t *testing.T) {
	simpleObjc, err := os.ReadFile(filepath.Join("testdata", "ios-simple-objc.pbxproj"))
	require.NoError(t, err)

	tests := []struct {
		name         string
		content      []byte
		projPath     string
		iconSetPaths [][]string
		want         []string
		wantErr      bool
	}{
		{
			name:         "single asset catalog",
			content:      simpleObjc,
			projPath:     "ios-simple-objc.xcodeproj",
			iconSetPaths: [][]string{{"ios-simple-objc", "Images.xcassets", "AppIcon.appiconset"}},
			want:         []string{"ios-simple-objc", "Images.xcassets", "AppIcon.appiconset"},
		},
		{
			name:     "asset catalog missing",
			content:  simpleObjc,
			projPath: "ios-simple-objc.xcodeproj",
			wantErr:  true,
		},
		{
			name:     "2 asset catalogs",
			content:  catalystPBXProj(t),
			projPath: "Catalyst Sample.xcodeproj",
			iconSetPaths: [][]string{
				{"Catalyst Sample", "Assets.xcassets", "AppIcon.appiconset"},
				{"Catalyst Sample", "Preview Content", "Preview Assets.appiconset"},
			},
			want: []string{"Catalyst Sample", "Assets.xcassets", "AppIcon.appiconset"},
		},
		{
			name:         "glob and regexp metacharacters in the path",
			content:      simpleObjc,
			projPath:     filepath.Join("Pro [ject] (1)+", "ios-simple-objc.xcodeproj"),
			iconSetPaths: [][]string{{"Pro [ject] (1)+", "ios-simple-objc", "Images.xcassets", "AppIcon.appiconset"}},
			want:         []string{"Pro [ject] (1)+", "ios-simple-objc", "Images.xcassets", "AppIcon.appiconset"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectDir := t.TempDir()
			for _, elements := range tt.iconSetPaths {
				require.NoError(t, os.MkdirAll(filepath.Join(append([]string{projectDir}, elements...)...), 0755))
			}

			factory := NewFactory(log.NewLogger(), nil, fileutil.NewFileManager(), nil, pathutil.NewPathModifier(), pathutil.NewPathProvider(), nil, nil)
			project, err := factory.newPBXProj(tt.content, filepath.Join(projectDir, tt.projPath))
			require.NoError(t, err)

			got, err := project.AppIconSetPaths()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			want := map[string][]string{
				project.Targets()[0].ID: {filepath.Join(append([]string{projectDir}, tt.want...)...)},
			}
			assert.Equal(t, want, got)
		})
	}
}
