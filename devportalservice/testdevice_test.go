package devportalservice

import (
	"path"
	"testing"
	"time"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeviceParsing(t *testing.T) {
	content := "00000000–0000000000000001,00000000–0000000000000002,00000000–0000000000000003"
	pth := path.Join(t.TempDir(), "devices.txt")
	err := fileutil.WriteStringToFile(pth, content)
	require.NoError(t, err)

	currentTime := time.Now()
	got := ParseTestDevicesFromFile(pth, currentTime)
	expected := []TestDevice{
		{
			DeviceID:   "00000000–0000000000000001",
			Title:      "Device 1",
			CreatedAt:  currentTime,
			UpdatedAt:  currentTime,
			DeviceType: "unknown",
		},
		{
			DeviceID:   "00000000–0000000000000002",
			Title:      "Device 2",
			CreatedAt:  currentTime,
			UpdatedAt:  currentTime,
			DeviceType: "unknown",
		},
		{
			DeviceID:   "00000000–0000000000000003",
			Title:      "Device 3",
			CreatedAt:  currentTime,
			UpdatedAt:  currentTime,
			DeviceType: "unknown",
		},
	}
	assert.Equal(t, expected, got)
}
