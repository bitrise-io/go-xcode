package xcproj

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
)

// testdata/ios-sample.xcproj is Xcode 27.2 output for a SwiftUI app with a unit test target and a
// UI test target.
func TestDecode_xcodeOutput(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "ios-sample.xcproj"))
	require.NoError(t, err)

	project, err := Decode(content)
	require.NoError(t, err)

	assert.Equal(t, "Release", project.DefaultConfiguration)
	assert.Equal(t, []Configuration{{Name: "Debug"}, {Name: "Release"}}, project.Configurations)

	debugOptimization, ok := project.BuildSettings.String("GCC_OPTIMIZATION_LEVEL[config=Debug]")
	require.True(t, ok)
	assert.Equal(t, "0", debugOptimization)

	preprocessorDefinitions, ok := project.BuildSettings.StringSlice("GCC_PREPROCESSOR_DEFINITIONS[config=Debug]")
	require.True(t, ok)
	assert.Equal(t, []string{"DEBUG=1", "$(inherited)"}, preprocessorDefinitions)

	require.Len(t, project.Targets, 3)

	app := project.Targets[0]
	assert.Equal(t, "16F0774C306FEF4F008B9392", app.ID)
	assert.Equal(t, "sample-apps-ios-json-xcproj", app.Name)
	assert.Equal(t, TargetKindNative, app.Kind)
	assert.Equal(t, "com.apple.product-type.application", app.ProductType)
	assert.Equal(t, "Products/sample-apps-ios-json-xcproj.app", app.Product)
	assert.Empty(t, app.TestHostTarget)
	assert.Empty(t, app.Dependencies)

	bundleID, ok := app.BuildSettings.String("PRODUCT_BUNDLE_IDENTIFIER")
	require.True(t, ok)
	assert.Equal(t, "io.bitrise.sample-apps-ios-json-xcproj", bundleID)

	runpathSearchPaths, ok := app.BuildSettings.StringSlice("LD_RUNPATH_SEARCH_PATHS")
	require.True(t, ok)
	assert.Equal(t, []string{"$(inherited)", "@executable_path/Frameworks"}, runpathSearchPaths)

	for _, test := range []struct {
		target      Target
		name        string
		productType string
	}{
		{target: project.Targets[1], name: "sample-apps-ios-json-xcprojTests", productType: "com.apple.product-type.bundle.unit-test"},
		{target: project.Targets[2], name: "sample-apps-ios-json-xcprojUITests", productType: "com.apple.product-type.bundle.ui-testing"},
	} {
		assert.Equal(t, test.name, test.target.Name)
		assert.Equal(t, test.productType, test.target.ProductType)
		assert.Equal(t, app.Name, test.target.TestHostTarget)
		assert.Equal(t, []string{app.Name}, test.target.Dependencies)
	}
}

func TestDecode_JSON5OnlySyntax(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "hand-edited-json5.xcproj"))
	require.NoError(t, err)

	_, err = Decode(content)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unquoted keys")
	assert.Contains(t, err.Error(), "line 5, column 3")
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    Project
		wantErr string
	}{
		{
			name:    "configuration written as an object",
			content: `{"configurations": ["Debug", {"id": "ABC", "name": "Release", "file": "Release.xcconfig"}]}`,
			want: Project{
				Configurations: []Configuration{{Name: "Debug"}, {ID: "ABC", Name: "Release"}},
				BuildSettings:  serialized.Object{},
				Targets:        []Target{},
			},
		},
		{
			name: "dependencies written as objects",
			content: `{"targets": [{"name": "App", "id": "A1", "kind": "aggregate", "dependencies": [
				"Framework",
				{"target": "Widget", "platforms": ["ios"]},
				{"kind": "localTarget", "target": "Clip"},
				{"kind": "remoteTarget", "project": "Other.xcodeproj", "target": "Lib", "target-id": "B2"},
				{"kind": "package", "product-name": "Alamofire"},
			]}]}`,
			want: Project{
				Configurations: []Configuration{},
				BuildSettings:  serialized.Object{},
				Targets: []Target{{
					ID:            "A1",
					Name:          "App",
					Kind:          TargetKindAggregate,
					Dependencies:  []string{"Framework", "Widget", "Clip"},
					BuildSettings: serialized.Object{},
				}},
			},
		},
		{
			name:    "product type outside com.apple.product-type.",
			content: `{"targets": [{"name": "Tool", "id": "A1", "full-product-type": "com.example.product-type.tool"}]}`,
			want: Project{
				Configurations: []Configuration{},
				BuildSettings:  serialized.Object{},
				Targets: []Target{{
					ID:            "A1",
					Name:          "Tool",
					Kind:          TargetKindNative,
					ProductType:   "com.example.product-type.tool",
					BuildSettings: serialized.Object{},
				}},
			},
		},
		{
			name:    "comments",
			content: "// header\n{\n  /* block */ \"default-configuration\": \"Debug\", // trailing\n}",
			want: Project{
				DefaultConfiguration: "Debug",
				Configurations:       []Configuration{},
				BuildSettings:        serialized.Object{},
				Targets:              []Target{},
			},
		},
		{
			name:    "configuration without a name",
			content: `{"configurations": [{"id": "ABC"}]}`,
			wantErr: "configuration 0 has no name",
		},
		{
			name:    "target without a name",
			content: `{"targets": [{"id": "A1"}]}`,
			wantErr: "target 0: target has no name",
		},
		{
			name:    "target without an id",
			content: `{"targets": [{"name": "App"}]}`,
			wantErr: "target 0: target App has no id",
		},
		{
			name:    "unknown target kind",
			content: `{"targets": [{"name": "App", "id": "A1", "kind": "legacy"}]}`,
			wantErr: "target 0: target App has unknown kind: legacy",
		},
		{
			name:    "local dependency without a target",
			content: `{"targets": [{"name": "App", "id": "A1", "dependencies": [{"platforms": ["ios"]}]}]}`,
			wantErr: "target 0: target App has a dependency with no target",
		},
		{
			name:    "wrong value type",
			content: `{"targets": {}}`,
			wantErr: "failed to decode project.xcproj",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Decode([]byte(tt.content))
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
