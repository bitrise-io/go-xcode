package destination

import (
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/destination/testdata"
	"github.com/bitrise-io/go-xcode/v2/mocks"
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

	got, err := d.ListDevices()
	require.NoError(t, err)

	wantDeviceTypes := []DeviceType{
		{
			Name:          "iPhone 6s",
			Identifier:    "com.apple.CoreSimulator.SimDeviceType.iPhone-6s",
			ProductFamily: "iPhone",
		},
	}
	wantRuntimes := []DeviceRuntime{{
		Identifier:  "com.apple.CoreSimulator.SimRuntime.iOS-16-0",
		Platform:    "iOS",
		Version:     "16.0",
		IsAvailable: true,
		Name:        "iOS 16.0",
		SupportedDeviceTypes: []DeviceType{
			{Name: "iPhone 8", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-8", ProductFamily: "iPhone"},
			{Name: "iPhone 8 Plus", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-8-Plus", ProductFamily: "iPhone"},
			{Name: "iPhone X", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-X", ProductFamily: "iPhone"},
			{Name: "iPhone Xs", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-XS", ProductFamily: "iPhone"},
			{Name: "iPhone Xs Max", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-XS-Max", ProductFamily: "iPhone"},
			{Name: "iPhone Xʀ", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-XR", ProductFamily: "iPhone"},
			{Name: "iPhone 11", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-11", ProductFamily: "iPhone"},
			{Name: "iPhone 11 Pro", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-11-Pro", ProductFamily: "iPhone"},
			{Name: "iPhone 11 Pro Max", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-11-Pro-Max", ProductFamily: "iPhone"},
			{Name: "iPhone SE (2nd generation)", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-SE--2nd-generation-", ProductFamily: "iPhone"},
			{Name: "iPhone 12 mini", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-12-mini", ProductFamily: "iPhone"},
			{Name: "iPhone 12", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-12", ProductFamily: "iPhone"},
			{Name: "iPhone 12 Pro", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-12-Pro", ProductFamily: "iPhone"},
			{Name: "iPhone 12 Pro Max", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-12-Pro-Max", ProductFamily: "iPhone"},
			{Name: "iPhone 13 Pro", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-13-Pro", ProductFamily: "iPhone"},
			{Name: "iPhone 13 Pro Max", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-13-Pro-Max", ProductFamily: "iPhone"},
			{Name: "iPhone 13 mini", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-13-mini", ProductFamily: "iPhone"},
			{Name: "iPhone 13", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-13", ProductFamily: "iPhone"},
			{Name: "iPhone SE (3rd generation)", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-SE-3rd-generation", ProductFamily: "iPhone"},
			{Name: "iPad Pro (9.7-inch)", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPad-Pro--9-7-inch-", ProductFamily: "iPad"},
			{Name: "iPad Pro (12.9-inch) (1st generation)", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPad-Pro", ProductFamily: "iPad"},
			{Name: "iPad (5th generation)", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPad--5th-generation-", ProductFamily: "iPad"},
			{Name: "iPad Pro (12.9-inch) (2nd generation)", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPad-Pro--12-9-inch---2nd-generation-", ProductFamily: "iPad"},
			{Name: "iPad Pro (10.5-inch)", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPad-Pro--10-5-inch-", ProductFamily: "iPad"},
			{Name: "iPad (6th generation)", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPad--6th-generation-", ProductFamily: "iPad"},
			{Name: "iPad (7th generation)", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPad--7th-generation-", ProductFamily: "iPad"},
			{Name: "iPad Pro (11-inch) (1st generation)", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPad-Pro--11-inch-", ProductFamily: "iPad"},
			{Name: "iPad Pro (12.9-inch) (3rd generation)", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPad-Pro--12-9-inch---3rd-generation-", ProductFamily: "iPad"},
			{Name: "iPad Pro (11-inch) (2nd generation)", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPad-Pro--11-inch---2nd-generation-", ProductFamily: "iPad"},
			{Name: "iPad Pro (12.9-inch) (4th generation)", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPad-Pro--12-9-inch---4th-generation-", ProductFamily: "iPad"},
			{Name: "iPad mini (5th generation)", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPad-mini--5th-generation-", ProductFamily: "iPad"},
			{Name: "iPad Air (3rd generation)", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPad-Air--3rd-generation-", ProductFamily: "iPad"},
			{Name: "iPad (8th generation)", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPad--8th-generation-", ProductFamily: "iPad"},
			{Name: "iPad (9th generation)", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPad-9th-generation", ProductFamily: "iPad"},
			{Name: "iPad Air (4th generation)", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPad-Air--4th-generation-", ProductFamily: "iPad"},
			{Name: "iPad Pro (11-inch) (3rd generation)", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPad-Pro-11-inch-3rd-generation", ProductFamily: "iPad"},
			{Name: "iPad Pro (12.9-inch) (5th generation)", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPad-Pro-12-9-inch-5th-generation", ProductFamily: "iPad"},
			{Name: "iPad Air (5th generation)", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPad-Air-5th-generation", ProductFamily: "iPad"},
			{Name: "iPad mini (6th generation)", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPad-mini-6th-generation", ProductFamily: "iPad"},
		},
	}}
	wantDevices := map[string][]Device{
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
