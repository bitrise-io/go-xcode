package destination

import (
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/destination/mocks"
	"github.com/bitrise-io/go-xcode/v2/destination/testdata"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_deviceFinder_parseDeviceList(t *testing.T) {
	commandFactory := new(mocks.CommandFactory)
	command := new(mocks.Command)

	command.On("PrintableCommandArgs").Return("xcrun simctl list")
	command.On("RunAndReturnTrimmedOutput").Once().Return(testdata.DeviceList, nil)
	commandFactory.On("Create", "xcrun", []string{"simctl", "list", "--json"}, mock.Anything).Return(command)

	d := deviceFinder{
		logger:         log.NewLogger(),
		commandFactory: commandFactory,
	}

	got, err := d.parseDeviceList()
	require.NoError(t, err)

	wantDeviceTypes := []deviceType{
		{
			Name:          "iPhone 6s",
			Identifier:    "com.apple.CoreSimulator.SimDeviceType.iPhone-6s",
			ProductFamily: "iPhone",
		},
	}
	wantRuntimes := []deviceRuntime{{
		Identifier:  "com.apple.CoreSimulator.SimRuntime.iOS-16-0",
		Platform:    "iOS",
		Version:     "16.0",
		IsAvailable: true,
		Name:        "iOS 16.0",
	}}
	wantDevices := map[string][]device{
		"com.apple.CoreSimulator.SimRuntime.iOS-16-0": {{
			Name:           "iPhone 11",
			TypeIdentifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-11",
			UDID:           "C4807E16-889C-42F1-BE22-6C4F1A25D807",
			State:          "Shutdown",
			IsAvailable:    true,
		}},
	}

	for _, d := range wantDeviceTypes {
		require.Contains(t, got.DeviceTypes, d)
	}

	for _, r := range wantRuntimes {
		require.Contains(t, got.Runtimes, r)
	}

	for runtime, deviceList := range wantDevices {
		gotList, ok := got.Devices[runtime]
		require.True(t, ok)

		for _, device := range deviceList {
			require.Contains(t, gotList, device)
		}
	}
}
