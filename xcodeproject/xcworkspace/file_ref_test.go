package xcworkspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseLocation(t *testing.T) {
	tests := []struct {
		location string
		wantType locationType
		wantPath string
		wantErr  string
	}{
		{location: "group:../SubProjectTests/Info.plist", wantType: groupLocation, wantPath: "../SubProjectTests/Info.plist"},
		{location: "container:App.xcodeproj", wantType: containerLocation, wantPath: "App.xcodeproj"},
		{location: "absolute:/opt/Pods/Pods.xcodeproj", wantType: absoluteLocation, wantPath: "/opt/Pods/Pods.xcodeproj"},
		{location: "developer:Applications/Simulator.app", wantType: developerLocation, wantPath: "Applications/Simulator.app"},
		{location: "self:", wantType: selfLocation, wantPath: ""},
		{location: "bogus:Something", wantErr: "unknown file reference type"},
		{location: "App.xcodeproj", wantErr: "unknown file reference location"},
	}
	for _, tt := range tests {
		t.Run(tt.location, func(t *testing.T) {
			refType, pth, err := parseLocation(tt.location)
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
