package xcodeproj

import (
	"testing"

	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func projectObjects(t *testing.T, p *xcodeProject) serialized.Object {
	t.Helper()
	objects, ok := pbxFile(t, p).rawProj.Object("objects")
	require.True(t, ok)
	return objects
}

func Test_assetCatalogs(t *testing.T) {
	project := parseFixture(t, "ios-simple-objc.pbxproj")

	got, err := assetCatalogs(pbxFile(t, project).buildPhaseIDs[project.Targets()[0].ID], projectObjects(t, project))
	require.NoError(t, err)
	assert.Equal(t, []fileReference{{id: "BA3CBE8819F7A93900CED4D5", path: "Images.xcassets"}}, got)
}
