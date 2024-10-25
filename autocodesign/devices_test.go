package autocodesign

import (
	"fmt"
	"testing"

	"github.com/bitrise-io/go-xcode/v2/autocodesign/devportalclient/appstoreconnect"
	"github.com/bitrise-io/go-xcode/v2/devportalservice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_registerMissingDevices_alreadyRegistered(t *testing.T) {
	bitriseDevices := []devportalservice.TestDevice{{
		DeviceID:   "71153a920968f2842d360",
		DeviceType: "models.IOS",
	}}
	devportalDevices := []appstoreconnect.Device{{
		Attributes: appstoreconnect.DeviceAttributes{
			UDID: "71153a920968f2842d360",
		},
		ID: "12",
	}}
	client := new(MockDevPortalClient)
	got, err := registerMissingTestDevices(client, bitriseDevices, devportalDevices)
	require.NoError(t, err, "registerMissingDevices() error")
	require.Equal(t, []appstoreconnect.Device(nil), got, "registerMissingDevices()")
}

func Test_registerMissingDevices_newDevice(t *testing.T) {
	bitriseDevice := devportalservice.TestDevice{
		DeviceID:   "71153a920968f2842d360",
		DeviceType: "models.IOS",
	}
	bitriseDevices := []devportalservice.TestDevice{bitriseDevice}
	devportalDevices := []appstoreconnect.Device{}
	retDevice := appstoreconnect.Device{
		Attributes: appstoreconnect.DeviceAttributes{
			DeviceClass: appstoreconnect.Iphone,
		},
		ID: "12",
	}
	want := []appstoreconnect.Device{retDevice}

	client := new(MockDevPortalClient)
	client.On("RegisterDevice", bitriseDevice).Return(&retDevice, nil).Once()

	got, err := registerMissingTestDevices(client, bitriseDevices, devportalDevices)
	require.NoError(t, err, "registerMissingDevices() error")
	require.Equal(t, want, got, "registerMissingDevices()")
	client.AssertExpectations(t)
}

func Test_registerMissingDevices_invalidUDID(t *testing.T) {
	bitriseDevice := devportalservice.TestDevice{
		DeviceID:   "71153a920968f2842d360",
		DeviceType: "models.IOS",
	}
	bitriseInvalidDevice := devportalservice.TestDevice{
		DeviceID:   "invalid-udid",
		DeviceType: "models.IOS",
	}
	bitriseDevices := []devportalservice.TestDevice{bitriseInvalidDevice, bitriseDevice}

	registeredDevice := appstoreconnect.Device{
		Attributes: appstoreconnect.DeviceAttributes{
			UDID: "71153a920968f2842d360",
		},
		ID: "12",
	}
	devportalDevices := []appstoreconnect.Device{}

	want := []appstoreconnect.Device{registeredDevice}

	client := new(MockDevPortalClient)
	client.On("RegisterDevice", bitriseInvalidDevice).Return(nil, appstoreconnect.DeviceRegistrationError{}).Once()
	client.On("RegisterDevice", bitriseDevice).Return(&registeredDevice, nil).Once()

	got, err := registerMissingTestDevices(client, bitriseDevices, devportalDevices)

	require.NoError(t, err, "registerMissingDevices()")
	require.Equal(t, want, got, "registerMissingDevices()")
	client.AssertExpectations(t)
}

func Test_listRelevantDevPortalDevices_filtersDevicesForPlatform(t *testing.T) {
	tests := []struct {
		platform      Platform
		deviceClass   appstoreconnect.DeviceClass
		devicesLength int
	}{
		{platform: IOS, deviceClass: appstoreconnect.AppleWatch, devicesLength: 1},
		{platform: IOS, deviceClass: appstoreconnect.Ipad, devicesLength: 1},
		{platform: IOS, deviceClass: appstoreconnect.Iphone, devicesLength: 1},
		{platform: IOS, deviceClass: appstoreconnect.Ipod, devicesLength: 1},
		{platform: IOS, deviceClass: appstoreconnect.AppleTV, devicesLength: 0},
		{platform: IOS, deviceClass: appstoreconnect.Mac, devicesLength: 0},

		{platform: TVOS, deviceClass: appstoreconnect.AppleWatch, devicesLength: 0},
		{platform: TVOS, deviceClass: appstoreconnect.Ipad, devicesLength: 0},
		{platform: TVOS, deviceClass: appstoreconnect.Iphone, devicesLength: 0},
		{platform: TVOS, deviceClass: appstoreconnect.Ipod, devicesLength: 0},
		{platform: TVOS, deviceClass: appstoreconnect.AppleTV, devicesLength: 1},
		{platform: TVOS, deviceClass: appstoreconnect.Mac, devicesLength: 0},

		{platform: MacOS, deviceClass: appstoreconnect.AppleWatch, devicesLength: 0},
		{platform: MacOS, deviceClass: appstoreconnect.Ipad, devicesLength: 0},
		{platform: MacOS, deviceClass: appstoreconnect.Iphone, devicesLength: 0},
		{platform: MacOS, deviceClass: appstoreconnect.Ipod, devicesLength: 0},
		{platform: MacOS, deviceClass: appstoreconnect.AppleTV, devicesLength: 0},
		{platform: MacOS, deviceClass: appstoreconnect.Mac, devicesLength: 0},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Given platform is %s when device is %s then valid devices length should be %d", tt.platform, tt.deviceClass, tt.devicesLength), func(t *testing.T) {
			devices := []appstoreconnect.Device{
				{
					Attributes: appstoreconnect.DeviceAttributes{
						DeviceClass: tt.deviceClass,
					},
				},
			}
			got := filterDevPortalDevices(devices, tt.platform)
			assert.Equal(t, tt.devicesLength, len(got))
		})
	}
}
