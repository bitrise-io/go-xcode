package appstoreconnectclient

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/autocodesign/devportalclient/appstoreconnect"
	"github.com/bitrise-io/go-xcode/v2/devportalservice"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockDeviceClient struct {
	mock.Mock
}

func (c *MockDeviceClient) Do(req *http.Request) (*http.Response, error) {
	fmt.Printf("do called: %#v - %#v\n", req.Method, req.URL.Path)

	switch {
	case req.URL.Path == "/v1/devices" && req.Method == "POST":
		return c.RegisterDevice(req)
	}

	return nil, fmt.Errorf("invalid endpoint called: %s, method: %s", req.URL.Path, req.Method)
}

func (c *MockDeviceClient) RegisterDevice(req *http.Request) (*http.Response, error) {
	args := c.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestDeviceClient_RegisterDevice_WhenInvaludUUID(t *testing.T) {
	logger := log.NewLogger(log.WithDebugLog(true))
	mockClient := MockDeviceClient{}
	mockClient.On("RegisterDevice", mock.Anything).Return(&http.Response{}, &appstoreconnect.ErrorResponse{
		Response: &http.Response{
			StatusCode: http.StatusConflict,
		},
	})

	client := appstoreconnect.NewClient(&mockClient, "keyID", "issueID", []byte("privateKey"), false, logger, appstoreconnect.NoOpAnalyticsTracker{})
	deviceClient := NewDeviceClient(client)

	got, err := deviceClient.RegisterDevice(devportalservice.TestDevice{
		DeviceID: "aadd",
	})

	require.IsType(t, appstoreconnect.DeviceRegistrationError{}, err)
	require.Nil(t, got)
}
