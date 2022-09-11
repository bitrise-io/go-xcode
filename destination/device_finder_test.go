package destination

import (
	"testing"

	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/destination/mocks"
	"github.com/bitrise-io/go-xcode/v2/destination/testdata"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

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
				Name:   "iPhone 8",
				ID:     "D64FA78C-5A25-4BF3-9EE8-855761042DEE",
				Status: "Shutdown",
				OS:     "16.0",
			},
		},
		{
			name: "device not available",
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
				Name:   "iPhone 8",
				ID:     "D64FA78C-5A25-4BF3-9EE8-855761042DEE",
				Status: "Shutdown",
				OS:     "16.0",
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
				Name:   "iPhone 8",
				ID:     "D64FA78C-5A25-4BF3-9EE8-855761042DEE",
				Status: "Shutdown",
				OS:     "16.0",
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
				Name:   "iPhone 8",
				ID:     "D64FA78C-5A25-4BF3-9EE8-855761042DEE",
				Status: "Shutdown",
				OS:     "16.0",
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
				Name:   "Apple Watch Series 7 - 45mm",
				ID:     "4F40330B-622F-4B44-8918-0BBE62720CC4",
				Status: "Shutdown",
				OS:     "9.0",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := deviceFinder{
				logger:         logger,
				commandFactory: commandFactory,
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

func Test_deviceFinder_FindDevice_realXcrun(t *testing.T) {
	commandFactory := command.NewFactory(env.NewRepository())
	logger := log.NewLogger()
	logger.EnableDebugLog(true)

	d := deviceFinder{
		logger:         logger,
		commandFactory: commandFactory,
	}

	got, err := d.FindDevice(Simulator{
		Platform: "iOS Simulator",
		OS:       "latest",
		Name:     "iPhone Xs",
	})

	require.NoError(t, err)
	require.NotEmpty(t, got.ID)

	t.Logf("got: %+v", got)
}
