package destination

import (
	"encoding/json"
	"errors"
	"slices"
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/destination/testdata"
	"github.com/bitrise-io/go-xcode/v2/mocks"
	"github.com/bitrise-io/go-xcode/v2/xcodeversion"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_deviceFinder_FindDevice_WhenCreateFails_ThenError(t *testing.T) {
	// Given
	var (
		commandFactory = new(mocks.CommandFactory)
		listCmd        = new(mocks.Command)
	)
	listCmd.On("PrintableCommandArgs").Return("xcrun simctl list --json")
	listCmd.On("RunAndReturnTrimmedOutput").Return(testdata.DeviceList, nil)
	commandFactory.On("Create", "xcrun", []string{"simctl", "list", "--json"}, mock.Anything).Return(listCmd)

	var (
		createCmd  = new(mocks.Command)
		createArgs = []string{
			"simctl",
			"create",
			"iPhone Xs",
			"com.apple.CoreSimulator.SimDeviceType.iPhone-XS",
			"com.apple.CoreSimulator.SimRuntime.iOS-16-0",
		}
	)
	createCmd.On("PrintableCommandArgs").Return("xcrun simctl create")
	createCmd.On("RunAndReturnTrimmedCombinedOutput").Return("", errors.New("create-err"))
	commandFactory.On("Create", "xcrun", createArgs, mock.Anything).Return(createCmd)

	logger := log.NewLogger()
	d := deviceFinder{
		logger:         logger,
		commandFactory: commandFactory,
	}

	requiredDevice := Simulator{
		Platform: "iOS Simulator",
		OS:       "latest",
		Name:     "iPhone Xs",
	}

	// When
	_, err := d.FindDevice(requiredDevice)
	// Then
	require.Error(t, err)
}

func Test_deviceFinder_FindDevice(t *testing.T) {
	command := new(mocks.Command)
	command.On("PrintableCommandArgs").Return("xcrun simctl list --json")
	command.On("RunAndReturnTrimmedOutput").Return(testdata.DeviceList, nil)

	debugCmd := new(mocks.Command)
	debugCmd.On("PrintableCommandArgs").Return("xcrun simctl list")
	debugCmd.On("Run").Return(nil)

	commandFactory := new(mocks.CommandFactory)
	commandFactory.On("Create", "xcrun", []string{"simctl", "list", "--json"}, mock.Anything).Return(command)
	commandFactory.On("Create", "xcrun", []string{"simctl", "list"}, mock.Anything).Return(debugCmd)

	logger := log.NewLogger()

	tests := []struct {
		name         string
		xcodeVersion xcodeversion.Version
		wantedDevice Simulator
		want         Device
		wantErr      bool
	}{
		{
			name: "latest",
			wantedDevice: Simulator{
				Platform: "iOS Simulator",
				OS:       "latest",
				Name:     "iPhone 8",
			},
			want: Device{
				UDID:           "D64FA78C-5A25-4BF3-9EE8-855761042DEE",
				State:          "Shutdown",
				TypeIdentifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-8",
				Type:           "iPhone 8",
				Platform:       "iOS Simulator",
				Name:           "iPhone 8",
				OS:             "16.0",
				IsAvailable:    true,
			},
		},
		{
			name:         "latest for an earler Xcode version",
			xcodeVersion: xcodeversion.Version{Major: 13},
			wantedDevice: Simulator{
				Platform: "iOS Simulator",
				OS:       "latest",
				Name:     "iPhone 8",
			},
			want: Device{
				UDID:           "3F9A1206-31E2-417B-BBC7-6330B52B8358",
				State:          "Shutdown",
				TypeIdentifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-8",
				Type:           "iPhone 8",
				Platform:       "iOS Simulator",
				Name:           "iPhone 8",
				OS:             "13.7",
				IsAvailable:    true,
			},
		},
		{
			name: "arch flag specified (Rosetta Simulator)",
			wantedDevice: Simulator{
				Platform: "iOS Simulator",
				OS:       "latest",
				Name:     "iPhone 8",
				Arch:     "x86_64",
			},
			want: Device{
				UDID:           "D64FA78C-5A25-4BF3-9EE8-855761042DEE",
				State:          "Shutdown",
				TypeIdentifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-8",
				Type:           "iPhone 8",
				Platform:       "iOS Simulator",
				Name:           "iPhone 8",
				OS:             "16.0",
				Arch:           "x86_64",
				IsAvailable:    true,
			},
		},
		{
			name: "device type not available",
			wantedDevice: Simulator{
				Platform: "iOS Simulator",
				OS:       "latest",
				Name:     "iPhone NotExists",
			},
			wantErr: true,
		},
		{
			name: "default device, already created",
			wantedDevice: Simulator{
				Platform: "iOS Simulator",
				OS:       "13.7",
				Name:     "Bitrise iOS default",
			},
			want: Device{
				UDID:           "20FDD0DF-0369-43FF-98E6-DBB8C820341E",
				State:          "Shutdown",
				TypeIdentifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-6s",
				Type:           "iPhone 6s",
				Platform:       "iOS Simulator",
				Name:           "Bitrise iOS default",
				OS:             "13.7",
				IsAvailable:    true,
			},
		},
		{
			name: "default device, but not yet created",
			wantedDevice: Simulator{
				Platform: "iOS Simulator",
				OS:       "latest",
				Name:     "Bitrise iOS default",
			},
			want: Device{
				UDID:           "D64FA78C-5A25-4BF3-9EE8-855761042DEE",
				State:          "Shutdown",
				TypeIdentifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-8",
				Type:           "iPhone 8",
				Platform:       "iOS Simulator",
				Name:           "iPhone 8",
				OS:             "16.0",
				IsAvailable:    true,
			},
		},
		{
			name: "device type not available",
			wantedDevice: Simulator{
				Platform: "iOS Simulator",
				OS:       "latest",
				Name:     "iPhone NotExists",
			},
			wantErr: true,
		},
		{
			name: "runtime not available",
			wantedDevice: Simulator{
				Platform: "iOS Simulator",
				OS:       "15.0",
				Name:     "iPhone 8",
			},
			wantErr: true,
		},
		{
			name: "explicit OS version",
			wantedDevice: Simulator{
				Platform: "iOS Simulator",
				OS:       "16.0",
				Name:     "iPhone 8",
			},
			want: Device{
				UDID:           "D64FA78C-5A25-4BF3-9EE8-855761042DEE",
				State:          "Shutdown",
				TypeIdentifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-8",
				Type:           "iPhone 8",
				Platform:       "iOS Simulator",
				Name:           "iPhone 8",
				OS:             "16.0",
				IsAvailable:    true,
			},
		},
		{
			name: "explicit OS version (only major)",
			wantedDevice: Simulator{
				Platform: "iOS Simulator",
				OS:       "16",
				Name:     "iPhone 8",
			},
			want: Device{
				UDID:           "D64FA78C-5A25-4BF3-9EE8-855761042DEE",
				State:          "Shutdown",
				TypeIdentifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-8",
				Type:           "iPhone 8",
				Platform:       "iOS Simulator",
				Name:           "iPhone 8",
				OS:             "16.0",
				IsAvailable:    true,
			},
		},
		{
			name: "explicit OS version with unused bugfix version",
			wantedDevice: Simulator{
				Platform: "iOS Simulator",
				OS:       "16.0.1",
				Name:     "iPhone 8",
			},
			want: Device{
				UDID:           "D64FA78C-5A25-4BF3-9EE8-855761042DEE",
				State:          "Shutdown",
				TypeIdentifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-8",
				Type:           "iPhone 8",
				Platform:       "iOS Simulator",
				Name:           "iPhone 8",
				OS:             "16.0",
				IsAvailable:    true,
			},
		},
		{
			name: "watch",
			wantedDevice: Simulator{
				Platform: "watchOS Simulator",
				OS:       "latest",
				Name:     "Apple Watch Series 7 - 45mm",
			},
			want: Device{
				UDID:           "4F40330B-622F-4B44-8918-0BBE62720CC4",
				State:          "Shutdown",
				TypeIdentifier: "com.apple.CoreSimulator.SimDeviceType.Apple-Watch-Series-7-45mm",
				Type:           "Apple Watch Series 7 - 45mm",
				Platform:       "watchOS Simulator",
				Name:           "Apple Watch Series 7 - 45mm",
				OS:             "9.0",
				IsAvailable:    true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := deviceFinder{
				logger:         logger,
				commandFactory: commandFactory,
				xcodeVersion:   tt.xcodeVersion,
			}

			got, err := d.FindDevice(tt.wantedDevice)

			if tt.wantErr {
				require.Error(t, err)
				t.Logf("Expected error: %s", err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_deviceFinder_FindDevice_MultipleRuntimesWithSameVersion(t *testing.T) {
	const (
		duoTypeID   = "com.apple.CoreSimulator.SimDeviceType.iPhone-Duo"
		xcodeDevice = "iOS Simulator"
	)

	tests := []struct {
		name           string
		wantedDevice   Simulator
		want           Device
		wantCreateArgs []string // the missing device is expected to be created with these simctl arguments
		wantErr        bool
	}{
		{
			name:         "single iPhone Duo-only runtime: device created for the runtime",
			wantedDevice: Simulator{Platform: xcodeDevice, OS: "27.1", Name: "iPhone Duo"},
			want: Device{
				UDID: "5BC7AC73-1349-4535-BDCF-28679722AA8B", State: "Shutdown", TypeIdentifier: duoTypeID, IsAvailable: true,
				Type: "iPhone Duo", Platform: xcodeDevice, Name: "iPhone Duo", OS: "27.1",
			},
		},
		{
			name:         "single iPhone Duo-only runtime: default device",
			wantedDevice: Simulator{Platform: xcodeDevice, OS: "27.1", Name: defaultDeviceName},
			want: Device{
				UDID: "03F49682-6BB2-4AD7-A71D-28FA34A0FB31", State: "Shutdown", TypeIdentifier: duoTypeID, IsAvailable: true,
				Type: "iPhone Duo", Platform: xcodeDevice, Name: defaultDeviceName, OS: "27.1",
			},
		},
		{
			name:         "single iPhone Duo-only runtime: other device types are incompatible",
			wantedDevice: Simulator{Platform: xcodeDevice, OS: "27.1", Name: "iPhone 17"},
			wantErr:      true,
		},
		{
			name:         "iPhone Duo-only and generic runtime: device created for the generic runtime",
			wantedDevice: Simulator{Platform: xcodeDevice, OS: "27.2", Name: "iPhone 18 Pro"},
			want: Device{
				UDID: "96CF0E04-7A22-4A25-A43E-1C455DA678FF", State: "Shutdown", TypeIdentifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-18-Pro", IsAvailable: true,
				Type: "iPhone 18 Pro", Platform: xcodeDevice, Name: "iPhone 18 Pro", OS: "27.2",
			},
		},
		{
			name:         "iPhone Duo-only and generic runtime: device created for the iPhone Duo-only runtime",
			wantedDevice: Simulator{Platform: xcodeDevice, OS: "latest", Name: "iPhone Duo"},
			want: Device{
				UDID: "524E5A02-EC53-4129-BB01-E6832E4FE238", State: "Shutdown", TypeIdentifier: duoTypeID, IsAvailable: true,
				Type: "iPhone Duo", Platform: xcodeDevice, Name: "iPhone Duo", OS: "27.2",
			},
		},
		{
			name:         "iPhone Duo-only and generic runtime: default device",
			wantedDevice: Simulator{Platform: xcodeDevice, OS: "latest", Name: defaultDeviceName},
			want: Device{
				UDID: "A54C9770-D487-455C-8BA2-9BDEFE5DD4E2", State: "Shutdown", TypeIdentifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-13", IsAvailable: true,
				Type: "iPhone 13", Platform: xcodeDevice, Name: defaultDeviceName, OS: "27.2",
			},
		},
		{
			name:         "iPhone Duo-only and generic runtime: device type unsupported by both runtimes",
			wantedDevice: Simulator{Platform: xcodeDevice, OS: "27.2", Name: "iPhone 8"},
			wantErr:      true,
		},
		{
			name:         "iPhone Duo-only and generic runtime: missing device is created for the runtime supporting it",
			wantedDevice: Simulator{Platform: xcodeDevice, OS: "27.2", Name: "iPhone 16 Pro"},
			wantCreateArgs: []string{
				"simctl", "create", "iPhone 16 Pro",
				"com.apple.CoreSimulator.SimDeviceType.iPhone-16-Pro",
				"com.apple.CoreSimulator.SimRuntime.iOS-27-2",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listCmd := new(mocks.Command)
			listCmd.On("PrintableCommandArgs").Return("xcrun simctl list --json")
			listCmd.On("RunAndReturnTrimmedOutput").Return(testdata.DeviceListXcode27, nil)

			debugCmd := new(mocks.Command)
			debugCmd.On("PrintableCommandArgs").Return("xcrun simctl list")
			debugCmd.On("Run").Return(nil)

			commandFactory := new(mocks.CommandFactory)
			commandFactory.Test(t)
			commandFactory.On("Create", "xcrun", []string{"simctl", "list", "--json"}, mock.Anything).Return(listCmd)
			commandFactory.On("Create", "xcrun", []string{"simctl", "list"}, mock.Anything).Return(debugCmd)

			if tt.wantCreateArgs != nil {
				createCmd := new(mocks.Command)
				createCmd.On("PrintableCommandArgs").Return("xcrun simctl create")
				createCmd.On("RunAndReturnTrimmedCombinedOutput").Return("", errors.New("create-err"))
				commandFactory.On("Create", "xcrun", tt.wantCreateArgs, mock.Anything).Return(createCmd)
			}

			d := deviceFinder{
				logger:         log.NewLogger(),
				commandFactory: commandFactory,
				xcodeVersion:   xcodeversion.Version{Major: 27, Minor: 2},
			}

			got, err := d.FindDevice(tt.wantedDevice)

			if tt.wantErr {
				require.Error(t, err)
				t.Logf("Expected error: %s", err)
			} else {
				require.NoError(t, err)
			}
			if tt.wantCreateArgs != nil {
				commandFactory.AssertCalled(t, "Create", "xcrun", tt.wantCreateArgs, mock.Anything)
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_deviceFinder_deviceForDestination_Xcode27EdgeCases(t *testing.T) {
	const (
		genericRuntimeID = "com.apple.CoreSimulator.SimRuntime.iOS-27-2"
		duoRuntimeID     = "com.apple.CoreSimulator.SimRuntime.iOS-27-2-iPhone19-4"
		duoTypeID        = "com.apple.CoreSimulator.SimDeviceType.iPhone-Duo"
	)

	parseList := func(t *testing.T) *DeviceList {
		var list DeviceList
		require.NoError(t, json.Unmarshal([]byte(testdata.DeviceListXcode27), &list))
		return &list
	}
	removeDevices := func(list *DeviceList, name string) {
		for runtimeID, devices := range list.Devices {
			list.Devices[runtimeID] = slices.DeleteFunc(devices, func(d Device) bool { return d.Name == name })
		}
	}
	duoDefaultDevice := Device{Name: defaultDeviceName, TypeIdentifier: duoTypeID, IsAvailable: true, UDID: "DUO-DEFAULT", State: "Shutdown"}
	find := func(list *DeviceList, os, name string) (Device, error) {
		d := deviceFinder{
			logger:       log.NewLogger(),
			xcodeVersion: xcodeversion.Version{Major: 27, Minor: 2},
			list:         list,
		}
		return d.deviceForDestination(Simulator{Platform: "iOS Simulator", OS: os, Name: name})
	}

	t.Run("default device not created: falls back to the generic runtime although the Duo-only one is listed first", func(t *testing.T) {
		list := parseList(t)
		removeDevices(list, defaultDeviceName)

		for _, os := range []string{"latest", "27.2"} {
			got, err := find(list, os, defaultDeviceName)
			require.NoError(t, err, os)
			require.Equal(t, "96CF0E04-7A22-4A25-A43E-1C455DA678FF", got.UDID, os)
			require.Equal(t, "iPhone 18 Pro", got.Type, os)
		}
	})

	t.Run("default device not created: the Duo-only runtime is used when it is the only one", func(t *testing.T) {
		list := parseList(t)
		removeDevices(list, defaultDeviceName)

		got, err := find(list, "27.1", defaultDeviceName)
		require.NoError(t, err)
		require.Equal(t, "5BC7AC73-1349-4535-BDCF-28679722AA8B", got.UDID)
		require.Equal(t, "iPhone Duo", got.Type)
	})

	t.Run("missing iPhone Duo is created for the Duo-only runtime", func(t *testing.T) {
		list := parseList(t)
		removeDevices(list, "iPhone Duo")

		_, err := find(list, "27.2", "iPhone Duo")

		var missingErr *missingDeviceErr
		require.ErrorAs(t, err, &missingErr)
		require.Equal(t, duoRuntimeID, missingErr.runtimeID)
	})

	t.Run("default device created on both runtimes: the generic one wins", func(t *testing.T) {
		list := parseList(t)
		list.Devices[duoRuntimeID] = append(list.Devices[duoRuntimeID], duoDefaultDevice)

		got, err := find(list, "27.2", defaultDeviceName)
		require.NoError(t, err)
		require.Equal(t, "A54C9770-D487-455C-8BA2-9BDEFE5DD4E2", got.UDID)
	})

	t.Run("default device unavailable on the generic runtime: the Duo-only one is used", func(t *testing.T) {
		list := parseList(t)
		for i, device := range list.Devices[genericRuntimeID] {
			if device.Name == defaultDeviceName {
				list.Devices[genericRuntimeID][i].IsAvailable = false
				list.Devices[genericRuntimeID][i].AvailabilityError = "runtime profile not found"
			}
		}
		list.Devices[duoRuntimeID] = append(list.Devices[duoRuntimeID], duoDefaultDevice)

		got, err := find(list, "27.2", defaultDeviceName)
		require.NoError(t, err)
		require.Equal(t, "DUO-DEFAULT", got.UDID)
	})

	t.Run("runtime without a device list entry is skipped", func(t *testing.T) {
		list := parseList(t)
		delete(list.Devices, duoRuntimeID)

		got, err := find(list, "27.2", "iPhone 18 Pro")
		require.NoError(t, err)
		require.Equal(t, "96CF0E04-7A22-4A25-A43E-1C455DA678FF", got.UDID)

		delete(list.Devices, genericRuntimeID)
		_, err = find(list, "27.2", "iPhone 18 Pro")
		require.ErrorContains(t, err, "no device exists for runtime")
	})
}

func Test_deviceFinder_runtimesForPlatformVersion(t *testing.T) {
	var (
		iPhone17  = DeviceType{Name: "iPhone 17", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-17", ProductFamily: "iPhone"}
		iPhone13  = DeviceType{Name: "iPhone 13", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-13", ProductFamily: "iPhone"}
		iPhoneDuo = DeviceType{Name: "iPhone Duo", Identifier: "com.apple.CoreSimulator.SimDeviceType.iPhone-Duo", ProductFamily: "iPhone"}

		generic      = DeviceRuntime{Identifier: "com.apple.CoreSimulator.SimRuntime.iOS-27-2", Platform: "iOS", Version: "27.2", IsAvailable: true, Name: "iOS 27.2", SupportedDeviceTypes: []DeviceType{iPhone17, iPhone13}}
		genericPatch = DeviceRuntime{Identifier: "com.apple.CoreSimulator.SimRuntime.iOS-27-2-1", Platform: "iOS", Version: "27.2.1", IsAvailable: true, Name: "iOS 27.2.1", SupportedDeviceTypes: []DeviceType{iPhone17, iPhone13}}
		duoOnly      = DeviceRuntime{Identifier: "com.apple.CoreSimulator.SimRuntime.iOS-27-2-iPhone19-4", Platform: "iOS", Version: "27.2", IsAvailable: true, Name: "iOS 27.2", SupportedDeviceTypes: []DeviceType{iPhoneDuo}}
		older        = DeviceRuntime{Identifier: "com.apple.CoreSimulator.SimRuntime.iOS-18-6", Platform: "iOS", Version: "18.6", IsAvailable: true, Name: "iOS 18.6", SupportedDeviceTypes: []DeviceType{iPhone13}}
		unavailable  = DeviceRuntime{Identifier: "com.apple.CoreSimulator.SimRuntime.iOS-27-3", Platform: "iOS", Version: "27.3", IsAvailable: false, Name: "iOS 27.3", SupportedDeviceTypes: []DeviceType{iPhone17, iPhone13}}
		watch        = DeviceRuntime{Identifier: "com.apple.CoreSimulator.SimRuntime.watchOS-27-2", Platform: "watchOS", Version: "27.2", IsAvailable: true, Name: "watchOS 27.2"}
	)

	tests := []struct {
		name          string
		runtimes      []DeviceRuntime
		xcodeVersion  xcodeversion.Version
		wantedVersion string
		want          []DeviceRuntime
		wantErr       bool
	}{
		{
			name:          "latest returns every runtime of the newest major.minor version, the generic one first",
			runtimes:      []DeviceRuntime{older, duoOnly, unavailable, watch, generic},
			wantedVersion: "latest",
			want:          []DeviceRuntime{generic, duoOnly},
		},
		{
			name:          "latest groups by major.minor, ignoring the patch version",
			runtimes:      []DeviceRuntime{duoOnly, genericPatch},
			wantedVersion: "latest",
			want:          []DeviceRuntime{genericPatch, duoOnly},
		},
		{
			name:          "latest respects the Xcode version",
			runtimes:      []DeviceRuntime{older, duoOnly, generic},
			xcodeVersion:  xcodeversion.Version{Major: 16},
			wantedVersion: "latest",
			want:          []DeviceRuntime{older},
		},
		{
			name:          "latest falls back to the first runtime if the Xcode version supports none",
			runtimes:      []DeviceRuntime{duoOnly, generic},
			xcodeVersion:  xcodeversion.Version{Major: 26},
			wantedVersion: "latest",
			want:          []DeviceRuntime{duoOnly},
		},
		{
			name:          "explicit version returns every runtime of the major.minor version, the generic ones first",
			runtimes:      []DeviceRuntime{older, duoOnly, genericPatch, unavailable, watch, generic},
			wantedVersion: "27.2",
			want:          []DeviceRuntime{genericPatch, generic, duoOnly},
		},
		{
			name:          "explicit version with a single match",
			runtimes:      []DeviceRuntime{older, duoOnly, generic},
			wantedVersion: "18.6",
			want:          []DeviceRuntime{older},
		},
		{
			name:          "explicit version without a match",
			runtimes:      []DeviceRuntime{older, duoOnly, generic},
			wantedVersion: "17.0",
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := deviceFinder{
				logger:       log.NewLogger(),
				xcodeVersion: tt.xcodeVersion,
				list:         &DeviceList{Runtimes: tt.runtimes},
			}

			got, err := d.runtimesForPlatformVersion("iOS", tt.wantedVersion)

			if tt.wantErr {
				require.Error(t, err)
				t.Logf("Expected error: %s", err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.want, got)
		})
	}
}
