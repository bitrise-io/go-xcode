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

// Mock types for testing Client.Do analytics tracking behavior
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

func TestClientDoAnalyticsTracking(t *testing.T) {
	mockTracker := &mockAnalyticsTracker{}

	t.Run("successful request", func(t *testing.T) {
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
		mockTracker.apiRequests = nil
		mockTracker.apiErrors = nil

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte(`{"errors": [{"code": "PARAMETER_ERROR.INVALID", "title": "Invalid parameter"}]}`))
			require.NoError(t, err)
		}))
		defer server.Close()

		// Use a mock HTTP client to avoid JWT token generation issues
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
		require.NoError(t, err)

		if !mockHTTPClient.called {
			t.Errorf("Expected HTTP client to be called")
		}

		if len(mockTracker.apiRequests) != 1 {
			t.Errorf("Expected 1 API request tracked, got %d", len(mockTracker.apiRequests))
		}

		if len(mockTracker.apiErrors) != 1 {
			t.Errorf("Expected 1 API error tracked, got %d", len(mockTracker.apiErrors))
		}

		requestRecord := mockTracker.apiRequests[0]
		if requestRecord.method != "POST" {
			t.Errorf("Expected method POST, got %s", requestRecord.method)
		}
		if requestRecord.statusCode != 400 {
			t.Errorf("Expected status code 400, got %d", requestRecord.statusCode)
		}

		errorRecord := mockTracker.apiErrors[0]
		if errorRecord.method != "POST" {
			t.Errorf("Expected method POST, got %s", errorRecord.method)
		}
		if errorRecord.statusCode != 400 {
			t.Errorf("Expected status code 400, got %d", errorRecord.statusCode)
		}
	})

	t.Run("network error", func(t *testing.T) {
		mockTracker.apiRequests = nil
		mockTracker.apiErrors = nil

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
		require.NoError(t, err)

		if len(mockTracker.apiRequests) != 0 {
			t.Errorf("Expected 0 API requests tracked, got %d", len(mockTracker.apiRequests))
		}

		if len(mockTracker.apiErrors) != 1 {
			t.Errorf("Expected 1 API error tracked, got %d", len(mockTracker.apiErrors))
		}

		record := mockTracker.apiErrors[0]
		if record.method != "GET" {
			t.Errorf("Expected method GET, got %s", record.method)
		}
		if record.statusCode != 0 {
			t.Errorf("Expected status code 0 for network error, got %d", record.statusCode)
		}
		if record.errorMessage != "network connection failed" {
			t.Errorf("Expected error message 'network connection failed', got %s", record.errorMessage)
		}
	})
}
