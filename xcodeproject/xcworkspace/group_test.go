package xcworkspace

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroup_unmarshal(t *testing.T) {
	var group Group
	require.NoError(t, xml.Unmarshal([]byte(`<Group location = "group:Group" name = "Group">
   <Group location = "group:SubProject" name = "SubProject">
      <FileRef location = "group:ViewController.swift"></FileRef>
   </Group>
   <FileRef location = "group:SubProject/SubProject.xcodeproj"></FileRef>
</Group>`), &group))

	refType, pth, err := group.TypeAndPath()
	require.NoError(t, err)
	assert.Equal(t, GroupFileRefType, refType)
	assert.Equal(t, "Group", pth)

	require.Len(t, group.FileRefs, 1)
	assert.Equal(t, "group:SubProject/SubProject.xcodeproj", group.FileRefs[0].Location)

	require.Len(t, group.Groups, 1)
	assert.Equal(t, "group:SubProject", group.Groups[0].Location)
	require.Len(t, group.Groups[0].FileRefs, 1)
	assert.Equal(t, "group:ViewController.swift", group.Groups[0].FileRefs[0].Location)
}
