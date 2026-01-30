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

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	logger := log.NewLogger(log.WithDebugLog(true))
	tracker := NoOpAnalyticsTracker{}
	got := NewClient(NewRetryableHTTPClient(logger, tracker), "keyID", "issuerID", []byte{}, false, logger, tracker)

	require.Equal(t, "appstoreconnect-v1", got.audience)

	wantURL, err := url.Parse("https://api.appstoreconnect.apple.com/")
	require.NoError(t, err)
	require.Equal(t, wantURL, got.BaseURL)
}

func TestNewEnterpriseClient(t *testing.T) {
	logger := log.NewLogger(log.WithDebugLog(true))
	tracker := NoOpAnalyticsTracker{}
	got := NewClient(NewRetryableHTTPClient(logger, tracker), "keyID", "issuerID", []byte{}, true, logger, tracker)

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
	isRetry    bool
}

type apiErrorRecord struct {
	method       string
	host         string
	endpoint     string
	statusCode   int
	errorMessage string
}

func (m *mockAnalyticsTracker) TrackAPIRequest(method, host, endpoint string, statusCode int, duration time.Duration, isRetry bool) {
	m.apiRequests = append(m.apiRequests, apiRequestRecord{
		method:     method,
		host:       host,
		endpoint:   endpoint,
		statusCode: statusCode,
		duration:   duration,
		isRetry:    isRetry,
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

func (m *mockHTTPClient) RoundTrip(req *http.Request) (*http.Response, error) {
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

		httpClient := &http.Client{}
		httpClient.Transport = newTrackingRoundTripper(httpClient.Transport, mockTracker)

		client := &Client{
			client:  httpClient,
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
		mockTransport := &mockHTTPClient{
			resp: &http.Response{
				StatusCode: 400,
				Body:       io.NopCloser(strings.NewReader(`{"errors": [{"code": "PARAMETER_ERROR.INVALID", "title": "Invalid parameter"}]}`)),
				Header:     http.Header{},
			},
		}

		httpClient := &http.Client{}
		httpClient.Transport = newTrackingRoundTripper(mockTransport, mockTracker)

		client := &Client{
			client:  httpClient,
			tracker: mockTracker,
		}

		req, err := http.NewRequest("POST", "https://example.com/test", nil)
		require.NoError(t, err)
		_, err = client.Do(req, nil)
		require.Error(t, err, "Expected error due to 400 Bad Request response")

		require.True(t, mockTransport.called, "Expected HTTP client to be called")

		require.Len(t, mockTracker.apiRequests, 1, "Expected 1 (failed) API requests tracked")
		require.Len(t, mockTracker.apiErrors, 1, "Expected 1 API error tracked")

		errorRecord := mockTracker.apiErrors[0]
		require.Equal(t, "POST", errorRecord.method)
		require.Equal(t, 400, errorRecord.statusCode)
	})

	t.Run("network error", func(t *testing.T) {
		mockTracker := &mockAnalyticsTracker{}

		mockTransport := &mockHTTPClient{
			err: errors.New("network connection failed"),
		}

		httpClient := &http.Client{}
		httpClient.Transport = newTrackingRoundTripper(mockTransport, mockTracker)

		client := &Client{
			client:  httpClient,
			tracker: mockTracker,
		}

		req, err := http.NewRequest("GET", "https://api.appstoreconnect.apple.com/test", nil)
		require.NoError(t, err)
		_, err = client.Do(req, nil)
		require.Error(t, err)

		require.Len(t, mockTracker.apiRequests, 1, "Expected 1 API request tracked (even though it failed)")
		require.Len(t, mockTracker.apiErrors, 1, "Expected 1 API error tracked")

		record := mockTracker.apiErrors[0]
		require.Equal(t, "GET", record.method)
		require.Equal(t, 0, record.statusCode)
		require.Contains(t, record.errorMessage, "network connection failed")
	})
}
