package xcworkspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileRef_TypeAndPath(t *testing.T) {
	tests := []struct {
		location string
		wantType FileRefType
		wantPath string
		wantErr  string
	}{
		{location: "group:../SubProjectTests/Info.plist", wantType: GroupFileRefType, wantPath: "../SubProjectTests/Info.plist"},
		{location: "container:App.xcodeproj", wantType: ContainerFileRefType, wantPath: "App.xcodeproj"},
		{location: "absolute:/opt/Pods/Pods.xcodeproj", wantType: AbsoluteFileRefType, wantPath: "/opt/Pods/Pods.xcodeproj"},
		{location: "developer:Applications/Simulator.app", wantType: DeveloperFileRefType, wantPath: "Applications/Simulator.app"},
		{location: "self:", wantType: SelfFileRefType, wantPath: ""},
		{location: "bogus:Something", wantErr: "unknown file reference type"},
		{location: "App.xcodeproj", wantErr: "unknown file reference location"},
	}
	for _, tt := range tests {
		t.Run(tt.location, func(t *testing.T) {
			refType, pth, err := FileRef{Location: tt.location}.TypeAndPath()
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantType, refType)
			assert.Equal(t, tt.wantPath, pth)
		})
	}
}
