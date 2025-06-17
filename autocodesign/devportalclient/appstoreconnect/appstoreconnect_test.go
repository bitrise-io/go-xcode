// Package appstoreconnect implements a client for the App Store Connect API.
//
// It contains type definitions, authentication and API calls, without business logic built on those API calls.
package appstoreconnect

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	got := NewClient(NewRetryableHTTPClient(), "keyID", "issuerID", []byte{}, false, NoOpAnalyticsTracker{})

	require.Equal(t, "appstoreconnect-v1", got.audience)

	wantURL, err := url.Parse("https://api.appstoreconnect.apple.com/")
	require.NoError(t, err)
	require.Equal(t, wantURL, got.BaseURL)
}

func TestNewEnterpriseClient(t *testing.T) {
	got := NewClient(NewRetryableHTTPClient(), "keyID", "issuerID", []byte{}, true, NoOpAnalyticsTracker{})

	require.Equal(t, "apple-developer-enterprise-v1", got.audience)

	wantURL, err := url.Parse("https://api.enterprise.developer.apple.com/")
	require.NoError(t, err)
	require.Equal(t, wantURL, got.BaseURL)
}

type mockAnalyticsTracker struct {
	apiRequests []apiRequestRecord
	apiErrors   []apiErrorRecord
	authErrors  []string
}

type apiRequestRecord struct {
	method     string
	host       string
	endpoint   string
	statusCode int
	duration   time.Duration
}

type apiErrorRecord struct {
	method       string
	host         string
	endpoint     string
	statusCode   int
	errorMessage string
}

func (m *mockAnalyticsTracker) TrackAPIRequest(method, host, endpoint string, statusCode int, duration time.Duration) {
	m.apiRequests = append(m.apiRequests, apiRequestRecord{
		method:     method,
		host:       host,
		endpoint:   endpoint,
		statusCode: statusCode,
		duration:   duration,
	})
}

func (m *mockAnalyticsTracker) TrackAPIError(method, host, endpoint string, statusCode int, errorMessage string) {
	m.apiErrors = append(m.apiErrors, apiErrorRecord{
		method:       method,
		host:         host,
		endpoint:     endpoint,
		statusCode:   statusCode,
		errorMessage: errorMessage,
	})
}

func (m *mockAnalyticsTracker) TrackAuthError(errorMessage string) {
	m.authErrors = append(m.authErrors, errorMessage)
}

type mockHTTPClient struct {
	resp   *http.Response
	err    error
	called bool
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.called = true
	return m.resp, m.err
}

func TestTracking(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		mockTracker := &mockAnalyticsTracker{}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"data": []}`))
			require.NoError(t, err)
		}))
		defer server.Close()

		client := &Client{
			client:  &http.Client{},
			tracker: mockTracker,
		}

		req, err := http.NewRequest("GET", server.URL+"/test", nil)
		require.NoError(t, err)
		_, err = client.Do(req, nil)
		require.NoError(t, err)

		if len(mockTracker.apiRequests) != 1 {
			t.Errorf("Expected 1 API request tracked, got %d", len(mockTracker.apiRequests))
		}

		if len(mockTracker.apiErrors) != 0 {
			t.Errorf("Expected 0 API errors tracked, got %d", len(mockTracker.apiErrors))
		}

		record := mockTracker.apiRequests[0]
		if record.method != "GET" {
			t.Errorf("Expected method GET, got %s", record.method)
		}
		if record.statusCode != 200 {
			t.Errorf("Expected status code 200, got %d", record.statusCode)
		}
	})

	t.Run("error response", func(t *testing.T) {
		mockTracker := &mockAnalyticsTracker{}
		mockHTTPClient := &mockHTTPClient{
			resp: &http.Response{
				StatusCode: 400,
				Body:       io.NopCloser(strings.NewReader(`{"errors": [{"code": "PARAMETER_ERROR.INVALID", "title": "Invalid parameter"}]}`)),
				Header:     http.Header{},
			},
		}

		client := &Client{
			client:  mockHTTPClient,
			tracker: mockTracker,
		}

		req, err := http.NewRequest("POST", "https://example.com/test", nil)
		require.NoError(t, err)
		_, err = client.Do(req, nil)
		require.Error(t, err, "Expected error due to 400 Bad Request response")

		require.True(t, mockHTTPClient.called, "Expected HTTP client to be called")

		require.Len(t, mockTracker.apiRequests, 0, "Expected 0 (successful) API requests tracked")
		require.Len(t, mockTracker.apiErrors, 1, "Expected 1 API error tracked")

		errorRecord := mockTracker.apiErrors[0]
		require.Equal(t, "POST", errorRecord.method)
		require.Equal(t, 400, errorRecord.statusCode)
	})

	t.Run("network error", func(t *testing.T) {
		mockTracker := &mockAnalyticsTracker{}

		mockHTTPClient := &mockHTTPClient{
			err: errors.New("network connection failed"),
		}

		client := &Client{
			client:  mockHTTPClient,
			tracker: mockTracker,
		}

		req, err := http.NewRequest("GET", "https://api.appstoreconnect.apple.com/test", nil)
		require.NoError(t, err)
		_, err = client.Do(req, nil)
		require.Error(t, err)

		require.Len(t, mockTracker.apiRequests, 0, "Expected 0 API requests tracked")
		require.Len(t, mockTracker.apiErrors, 1, "Expected 1 API error tracked")

		record := mockTracker.apiErrors[0]
		require.Equal(t, "GET", record.method)
		require.Equal(t, 0, record.statusCode)
		require.Equal(t, "network connection failed", record.errorMessage)
	})
}
