package devportalservice

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/stretchr/testify/require"
)

func TestGetAppleDeveloperConnection(t *testing.T) {
	tests := []struct {
		name string

		responseContent string
		err             error

		want    *AppleDeveloperConnection
		wantErr bool
	}{
		{
			name:            "No Apple Developer Connection set for the build",
			responseContent: "{}",
			want:            &AppleDeveloperConnection{},
			wantErr:         false,
		},
		{
			name:            "No Apple Developer Connection set for the build, test devices available",
			responseContent: testDevicesResponseBody,
			want:            &testConnectionOnlyDevices,
			wantErr:         false,
		},
		{
			name:            "Apple ID-based Apple Developer Connection set for the build",
			responseContent: testAppleIDConnectionResponseBody,
			want:            &testConnectionWithAppleIDConnection,
			wantErr:         false,
		},
		{
			name:            "API key Apple Developer Connection set for the build",
			responseContent: testAPIKeyConnectionResponseBody,
			want:            &testConnectionWithAPIKeyConnection,
			wantErr:         false,
		},
		{
			name:            "Apple ID-based and API key Apple Developer Connection set for the build, test device available",
			responseContent: testAppleIDAndAPIKeyConnectionResponseBody,
			want:            &testConnectionWithAppleIDAndAPIKeyConnection,
			wantErr:         false,
		},
	}

	logger := log.NewLogger()
	for _, tt := range tests {
		t.Run(tt.name+"_http", func(t *testing.T) {
			response :=
				&http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(tt.responseContent)),
				}

			c := NewBitriseClient(logger, nil, newMockHTTPClient(response, nil), "dummy url", "dummy token")
			got, err := c.GetAppleDeveloperConnection()
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})

		t.Run(tt.name+"_file", func(t *testing.T) {
			fileManager := NewFileManager(t)
			fileManager.On("OpenReaderIfExists", "dummy url").Return(io.NopCloser(strings.NewReader(tt.responseContent)), nil)
			c := NewBitriseClient(logger, fileManager, nil, "file://dummy url", "dummy token")
			got, err := c.GetAppleDeveloperConnection()
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestFastlaneLoginSession(t *testing.T) {
	tests := []struct {
		name string

		response *http.Response
		err      error

		want    string
		wantErr bool
	}{
		{
			name: "No Apple Developer Connection set for the build",
			response: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("{}")),
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Apple ID-based Apple Developer Connection set for the build",
			response: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(testAppleIDConnectionResponseBody)),
			},
			want:    testFastlaneSession,
			wantErr: false,
		},
	}

	logger := log.NewLogger()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewBitriseClient(logger, nil, newMockHTTPClient(tt.response, nil), "dummy url", "dummy token")
			conn, err := c.GetAppleDeveloperConnection()
			require.NoError(t, err)

			if tt.want == "" {
				require.Nil(t, conn.AppleIDConnection)
				return
			}

			got, err := conn.AppleIDConnection.FastlaneLoginSession()
			if (err != nil) != tt.wantErr {
				t.Errorf("FastlaneLoginSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FastlaneLoginSession() = %v, want %v", got, tt.want)
			}
		})
	}
}

type mockHTTPClient struct {
	response *http.Response
	err      error
}

func newMockHTTPClient(response *http.Response, err error) mockHTTPClient {
	return mockHTTPClient{response: response, err: err}
}

func (c mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.response, c.err
}

func Test_validateTestDevice(t *testing.T) {
	tests := []struct {
		name                  string
		deviceList            []TestDevice
		wantValidDevices      []TestDevice
		wantDuplicatedDevices []TestDevice
	}{
		{
			name: "One device",
			deviceList: []TestDevice{{
				DeviceID: "612cb2257",
			}},
			wantValidDevices: []TestDevice{{
				DeviceID: "612cb2257",
			}},
			wantDuplicatedDevices: nil,
		},
		{
			name: "One device with whitespace",
			deviceList: []TestDevice{{
				DeviceID: "612c b2257 ",
			}},
			wantValidDevices: []TestDevice{{
				DeviceID: "612cb2257",
			}},
			wantDuplicatedDevices: nil,
		},
		{
			name: "One device with separator",
			deviceList: []TestDevice{{
				DeviceID: "00008020-00213C3D2201002F",
			}},
			wantValidDevices: []TestDevice{{
				DeviceID: "00008020-00213C3D2201002F",
			}},
			wantDuplicatedDevices: nil,
		},
		{
			name: "Duplicate devices",
			deviceList: []TestDevice{
				{
					DeviceID: "00008020-00213C3D2201002F",
				},
				{
					DeviceID: "0000802000213C3D2201002F",
				},
			},
			wantValidDevices: []TestDevice{{
				DeviceID: "00008020-00213C3D2201002F",
			}},
			wantDuplicatedDevices: []TestDevice{{
				DeviceID: "0000802000213C3D2201002F",
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValidDevices, gotDuplicatedDevices := validateTestDevice(tt.deviceList)
			require.Equal(t, tt.wantValidDevices, gotValidDevices, "normalizeTestDevices() validDevices")
			require.Equal(t, tt.wantDuplicatedDevices, gotDuplicatedDevices, "normalizeTestDevices() duplicateDevices")
		})
	}
}

func TestTestDevice_IsEqualUDID(t *testing.T) {
	tests := []struct {
		name            string
		UDID, otherUDID string
		want            bool
	}{
		{
			name:      "Same device",
			UDID:      "00008020-00213C3D2201002F",
			otherUDID: "00008020-00213C3D2201002F",
			want:      true,
		},
		{
			name:      "Same device with different casing and separators",
			UDID:      "00008020-00213C3D2201002F",
			otherUDID: "00008020 00213c3d2201002f",
			want:      true,
		},
		{
			name:      "Different devices",
			UDID:      "00008020-00213C3D2201002F",
			otherUDID: "00008020-00213C3D2201002G",
			want:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEqualUDID(tt.UDID, tt.otherUDID); got != tt.want {
				t.Errorf("TestDevice.EqualsTo() = %v, want %v", got, tt.want)
			}
		})
	}
}
