package autocodesign

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/bitrise-io/go-xcode/autocodesign/devportalclient/appstoreconnect"
	"github.com/bitrise-io/go-xcode/autocodesign/devportalclient/appstoreconnectclient"
	"github.com/bitrise-io/go-xcode/devportalservice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_registerMissingDevices_alreadyRegistered(t *testing.T) {
	mockClient := &MockClientRegisterDevice{}
	successClient := appstoreconnect.NewClient(mockClient, "keyID", "issueID", []byte("privateKey"))

	args := struct {
		client           DevPortalClient
		bitriseDevices   []devportalservice.TestDevice
		devportalDevices []appstoreconnect.Device
	}{
		client: appstoreconnectclient.NewAPIDeviceClient(successClient),
		bitriseDevices: []devportalservice.TestDevice{{
			DeviceID:   "71153a920968f2842d360",
			DeviceType: "ios",
		}},
		devportalDevices: []appstoreconnect.Device{{
			Attributes: appstoreconnect.DeviceAttributes{
				UDID: "71153a920968f2842d360",
			},
			ID: "12",
		}},
	}
	want := []appstoreconnect.Device(nil)

	got, err := registerMissingTestDevices(args.client, args.bitriseDevices, args.devportalDevices)

	require.NoError(t, err, "registerMissingDevices() error")
	require.Equal(t, want, got, "registerMissingDevices()")
	mockClient.AssertNotCalled(t, "PostDevice")
	mockClient.AssertExpectations(t)
}

func Test_registerMissingDevices_newDevice(t *testing.T) {
	mockClient := &MockClientRegisterDevice{}
	mockClient.
		On("PostDevice", mock.AnythingOfType("*http.Request")).
		Return(newResponse(t, http.StatusOK,
			map[string]interface{}{
				"data": map[string]interface{}{
					"id": "12",
					"attributes": map[string]interface{}{
						"deviceClass": "IPHONE",
					},
				},
			},
		), nil)
	successClient := appstoreconnect.NewClient(mockClient, "keyID", "issueID", []byte("privateKey"))

	args := struct {
		client           DevPortalClient
		bitriseDevices   []devportalservice.TestDevice
		devportalDevices []appstoreconnect.Device
	}{
		client: appstoreconnectclient.NewAPIDeviceClient(successClient),
		bitriseDevices: []devportalservice.TestDevice{{
			DeviceID:   "71153a920968f2842d360",
			DeviceType: "ios",
		}},
		devportalDevices: []appstoreconnect.Device{},
	}
	want := []appstoreconnect.Device{{
		Attributes: appstoreconnect.DeviceAttributes{
			DeviceClass: appstoreconnect.Iphone,
		},
		ID: "12",
	}}

	got, err := registerMissingTestDevices(args.client, args.bitriseDevices, args.devportalDevices)

	require.NoError(t, err, "registerMissingDevices()")
	require.Equal(t, want, got, "registerMissingDevices()")
	mockClient.AssertExpectations(t)
}

func Test_registerMissingDevices_invalidUDID(t *testing.T) {
	mockClient := &MockClientRegisterDevice{}
	mockClient.
		On("PostDevice", mock.AnythingOfType("*http.Request")).
		Return(newResponse(t, http.StatusConflict,
			map[string]interface{}{
				"errors": []interface{}{map[string]interface{}{"detail": "ENTITY_ERROR.ATTRIBUTE.INVALID: An attribute in the provided entity has invalid value: An invalid value 'xxx' was provided for the parameter 'udid'."}},
			}), nil)
	failureClient := appstoreconnect.NewClient(mockClient, "keyID", "issueID", []byte("privateKey"))

	args := struct {
		client           DevPortalClient
		bitriseDevices   []devportalservice.TestDevice
		devportalDevices []appstoreconnect.Device
	}{
		client: appstoreconnectclient.NewAPIDeviceClient(failureClient),
		bitriseDevices: []devportalservice.TestDevice{
			{
				DeviceID:   "invalid-udid",
				DeviceType: "ios",
			},
			{
				DeviceID:   "71153a920968f2842d360",
				DeviceType: "ios",
			},
		},
		devportalDevices: []appstoreconnect.Device{{
			Attributes: appstoreconnect.DeviceAttributes{
				UDID: "71153a920968f2842d360",
			},
			ID: "12",
		}},
	}
	want := []appstoreconnect.Device(nil)

	got, err := registerMissingTestDevices(args.client, args.bitriseDevices, args.devportalDevices)
	require.NoError(t, err, "registerMissingDevices()")
	require.Equal(t, want, got, "registerMissingDevices()")
	mockClient.AssertExpectations(t)
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

type MockClientRegisterDevice struct {
	mock.Mock
}

func (c *MockClientRegisterDevice) Do(req *http.Request) (*http.Response, error) {
	fmt.Printf("do called: %#v - %#v\n", req.Method, req.URL.Path)

	switch {
	case req.URL.Path == "/v1/devices" && req.Method == "POST":
		return c.PostDevice(req)
	}

	return nil, fmt.Errorf("invalid endpoint called: %s, method: %s", req.URL.Path, req.Method)
}

func (c *MockClientRegisterDevice) PostDevice(req *http.Request) (*http.Response, error) {
	args := c.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}
