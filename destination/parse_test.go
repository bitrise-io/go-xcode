package destination

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/destination/testdata"
	"github.com/bitrise-io/go-xcode/v2/mocks"
	"github.com/bitrise-io/go-xcode/v2/xcodeversion"
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

func Test_deviceFinder_deviceForDestination_DeviceSpecificRuntime(t *testing.T) {
	var list DeviceList
	require.NoError(t, json.Unmarshal([]byte(testdata.DeviceListWithDeviceSpecificRuntime), &list))

	const (
		genericRuntimeID        = "com.apple.CoreSimulator.SimRuntime.iOS-27-2"
		deviceSpecificRuntimeID = "com.apple.CoreSimulator.SimRuntime.iOS-27-2-iPhone19-4"
	)
	iPhone17 := Device{
		Name:           "iPhone 17",
		TypeIdentifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-17",
		IsAvailable:    true,
		UDID:           "FB72DF49-C0D1-4F75-ADC3-D42850028AD6",
		State:          "Shutdown",
		Type:           "iPhone 17",
		Platform:       "iOS Simulator",
		OS:             "27.2",
	}

	tests := []struct {
		name          string
		wantedDevice  Simulator
		want          Device
		wantMissingIn string
	}{
		{
			name:         "existing device, latest",
			wantedDevice: Simulator{Platform: "iOS Simulator", OS: "latest", Name: "iPhone 17"},
			want:         iPhone17,
		},
		{
			name:          "missing device is created in the generic runtime, explicit version",
			wantedDevice:  Simulator{Platform: "iOS Simulator", OS: "27.2", Name: "iPhone 12 Pro Max"},
			wantMissingIn: genericRuntimeID,
		},
		{
			name:          "missing device is created in the generic runtime, latest",
			wantedDevice:  Simulator{Platform: "iOS Simulator", OS: "latest", Name: "iPhone 12 Pro Max"},
			wantMissingIn: genericRuntimeID,
		},
		{
			name:         "default device falls back to a device of the generic runtime",
			wantedDevice: Simulator{Platform: "iOS Simulator", OS: "latest", Name: "Bitrise iOS default"},
			want:         iPhone17,
		},
		{
			name:         "existing device of the device-specific runtime",
			wantedDevice: Simulator{Platform: "iOS Simulator", OS: "27.2", Name: "iPhone Duo"},
			want: Device{
				Name:           "iPhone Duo",
				TypeIdentifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-Duo",
				IsAvailable:    true,
				UDID:           "6F0E6D53-7F1B-4F0B-9C2E-1D2A4B3C5D01",
				State:          "Shutdown",
				Type:           "iPhone Duo",
				Platform:       "iOS Simulator",
				OS:             "27.2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := deviceFinder{
				logger: log.NewLogger(),
				list:   &list,
			}

			got, err := d.deviceForDestination(tt.wantedDevice)

			if tt.wantMissingIn != "" {
				var missingErr *missingDeviceErr
				require.ErrorAs(t, err, &missingErr)
				require.Equal(t, tt.wantMissingIn, missingErr.runtimeID)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}

	t.Run("newer patch of the generic runtime", func(t *testing.T) {
		listWithPatch := list
		listWithPatch.Runtimes = slices.Clone(list.Runtimes)
		for i, r := range listWithPatch.Runtimes {
			if r.Identifier == genericRuntimeID {
				listWithPatch.Runtimes[i].Version = "27.2.1"
			}
		}
		listWithPatch.Devices = map[string][]Device{
			genericRuntimeID:        {},
			deviceSpecificRuntimeID: {},
		}
		d := deviceFinder{
			logger: log.NewLogger(),
			list:   &listWithPatch,
		}

		_, err := d.deviceForDestination(Simulator{Platform: "iOS Simulator", OS: "27.2", Name: "iPhone 12 Pro Max"})
		var missingErr *missingDeviceErr
		require.ErrorAs(t, err, &missingErr)
		require.Equal(t, genericRuntimeID, missingErr.runtimeID)

		// The device-specific runtime has an older patch version, but it is the only one that supports the device.
		_, err = d.deviceForDestination(Simulator{Platform: "iOS Simulator", OS: "27.2", Name: "iPhone Duo"})
		require.ErrorAs(t, err, &missingErr)
		require.Equal(t, deviceSpecificRuntimeID, missingErr.runtimeID)
	})

	t.Run("newer patch of the device-specific runtime", func(t *testing.T) {
		listWithPatch := list
		listWithPatch.Runtimes = slices.Clone(list.Runtimes)
		for i, r := range listWithPatch.Runtimes {
			if r.Identifier == deviceSpecificRuntimeID {
				listWithPatch.Runtimes[i].Version = "27.2.1"
			}
		}
		d := deviceFinder{
			logger: log.NewLogger(),
			list:   &listWithPatch,
		}

		for _, os := range []string{"latest", "27.2"} {
			got, err := d.deviceForDestination(Simulator{Platform: "iOS Simulator", OS: os, Name: "iPhone 17"})
			require.NoError(t, err)
			require.Equal(t, iPhone17, got)

			got, err = d.deviceForDestination(Simulator{Platform: "iOS Simulator", OS: os, Name: "Bitrise iOS default"})
			require.NoError(t, err)
			require.Equal(t, iPhone17, got)

			_, err = d.deviceForDestination(Simulator{Platform: "iOS Simulator", OS: os, Name: "iPhone 12 Pro Max"})
			var missingErr *missingDeviceErr
			require.ErrorAs(t, err, &missingErr)
			require.Equal(t, genericRuntimeID, missingErr.runtimeID)

			got, err = d.deviceForDestination(Simulator{Platform: "iOS Simulator", OS: os, Name: "iPhone Duo"})
			require.NoError(t, err)
			require.Equal(t, "6F0E6D53-7F1B-4F0B-9C2E-1D2A4B3C5D01", got.UDID)
		}
	})

	t.Run("device type not supported by any runtime", func(t *testing.T) {
		listWithoutSupport := list
		listWithoutSupport.Runtimes = slices.Clone(list.Runtimes)
		for i, r := range listWithoutSupport.Runtimes {
			listWithoutSupport.Runtimes[i].SupportedDeviceTypes = slices.DeleteFunc(slices.Clone(r.SupportedDeviceTypes), func(dt DeviceType) bool {
				return dt.Name == "iPhone 12 Pro Max"
			})
		}
		d := deviceFinder{
			logger: log.NewLogger(),
			list:   &listWithoutSupport,
		}

		_, err := d.deviceForDestination(Simulator{Platform: "iOS Simulator", OS: "27.2", Name: "iPhone 12 Pro Max"})
		require.EqualError(t, err, "no iOS 27.2 runtime supports device type iPhone 12 Pro Max (checked runtimes: "+genericRuntimeID+", "+deviceSpecificRuntimeID+"). Select a different device or OS version")
	})

	t.Run("latest runtime is device-specific only", func(t *testing.T) {
		// Like Xcode 27.1: the latest runtime only supports iPhone Duo, and there is no generic runtime for that version.
		listLatestDuo := list
		listLatestDuo.Runtimes = slices.DeleteFunc(slices.Clone(list.Runtimes), func(r DeviceRuntime) bool { return r.Identifier == genericRuntimeID })
		d := deviceFinder{
			logger: log.NewLogger(),
			list:   &listLatestDuo,
		}

		_, err := d.deviceForDestination(Simulator{Platform: "iOS Simulator", OS: "latest", Name: "iPhone 17"})
		require.EqualError(t, err, "no iOS latest (27.2) runtime supports device type iPhone 17 (checked runtimes: "+deviceSpecificRuntimeID+"). Select a different device or OS version")
	})

	t.Run("default device does not exist", func(t *testing.T) {
		listWithoutDevices := list
		listWithoutDevices.Devices = map[string][]Device{
			genericRuntimeID:        {},
			deviceSpecificRuntimeID: list.Devices[deviceSpecificRuntimeID],
		}
		d := deviceFinder{
			logger: log.NewLogger(),
			list:   &listWithoutDevices,
		}

		_, err := d.deviceForDestination(Simulator{Platform: "iOS Simulator", OS: "27.2", Name: "Bitrise iOS default"})
		require.EqualError(t, err, "no available device found for Bitrise iOS default (checked runtimes: "+genericRuntimeID+")")
	})

	t.Run("latest when Xcode supports no runtime", func(t *testing.T) {
		d := deviceFinder{
			logger:       log.NewLogger(),
			list:         &list,
			xcodeVersion: xcodeversion.Version{Major: 26, Minor: 0},
		}

		got, err := d.deviceForDestination(Simulator{Platform: "iOS Simulator", OS: "latest", Name: "iPhone 17"})
		require.NoError(t, err)
		require.Equal(t, iPhone17, got)
	})

	t.Run("runtime with an invalid version is skipped", func(t *testing.T) {
		listWithInvalid := list
		listWithInvalid.Runtimes = append(slices.Clone(list.Runtimes), DeviceRuntime{
			Identifier:  "com.apple.CoreSimulator.SimRuntime.iOS-invalid",
			Platform:    "iOS",
			Version:     "invalid",
			IsAvailable: true,
		})
		d := deviceFinder{
			logger: log.NewLogger(),
			list:   &listWithInvalid,
		}

		got, err := d.deviceForDestination(Simulator{Platform: "iOS Simulator", OS: "27.2", Name: "iPhone 17"})
		require.NoError(t, err)
		require.Equal(t, iPhone17, got)
	})

	t.Run("missing device of the device-specific runtime", func(t *testing.T) {
		listWithoutDuo := list
		listWithoutDuo.Devices = map[string][]Device{
			genericRuntimeID:        list.Devices[genericRuntimeID],
			deviceSpecificRuntimeID: {},
		}
		d := deviceFinder{
			logger: log.NewLogger(),
			list:   &listWithoutDuo,
		}

		_, err := d.deviceForDestination(Simulator{Platform: "iOS Simulator", OS: "27.2", Name: "iPhone Duo"})

		var missingErr *missingDeviceErr
		require.ErrorAs(t, err, &missingErr)
		require.Equal(t, deviceSpecificRuntimeID, missingErr.runtimeID)
	})
}

func Test_deviceFinder_isDeviceSpecificRuntime(t *testing.T) {
	generic := DeviceRuntime{Identifier: "com.apple.CoreSimulator.SimRuntime.iOS-27-2", Platform: "iOS", Version: "27.2", IsAvailable: true}
	unavailableGeneric := DeviceRuntime{Identifier: "com.apple.CoreSimulator.SimRuntime.iOS-27-2", Platform: "iOS", Version: "27.2", IsAvailable: false}
	deviceSpecific := DeviceRuntime{Identifier: "com.apple.CoreSimulator.SimRuntime.iOS-27-2-iPhone19-4", Platform: "iOS", Version: "27.2", IsAvailable: true}
	deviceSpecificPatch := DeviceRuntime{Identifier: "com.apple.CoreSimulator.SimRuntime.iOS-27-2-iPhone19-4", Platform: "iOS", Version: "27.2.1", IsAvailable: true}
	genericPatch := DeviceRuntime{Identifier: "com.apple.CoreSimulator.SimRuntime.iOS-17-0", Platform: "iOS", Version: "17.0.1", IsAvailable: true}

	tests := []struct {
		name     string
		runtimes []DeviceRuntime
		runtime  DeviceRuntime
		want     bool
	}{
		{name: "generic runtime", runtimes: []DeviceRuntime{generic, deviceSpecific}, runtime: generic, want: false},
		{name: "device-specific runtime", runtimes: []DeviceRuntime{generic, deviceSpecific}, runtime: deviceSpecific, want: true},
		{name: "device-specific runtime with a newer patch version", runtimes: []DeviceRuntime{generic, deviceSpecificPatch}, runtime: deviceSpecificPatch, want: true},
		{name: "device-specific runtime without a generic runtime", runtimes: []DeviceRuntime{deviceSpecific}, runtime: deviceSpecific, want: false},
		{name: "device-specific runtime with an unavailable generic runtime", runtimes: []DeviceRuntime{unavailableGeneric, deviceSpecific}, runtime: deviceSpecific, want: false},
		{name: "generic runtime with a patch version", runtimes: []DeviceRuntime{genericPatch}, runtime: genericPatch, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := deviceFinder{list: &DeviceList{Runtimes: tt.runtimes}}
			require.Equal(t, tt.want, d.isDeviceSpecificRuntime(tt.runtime))
		})
	}
}
