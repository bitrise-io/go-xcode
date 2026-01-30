package appstoreconnect

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/stretchr/testify/require"
)

type attemptTracker struct {
	mu       sync.Mutex
	attempts []attemptRecord
}

type attemptRecord struct {
	method     string
	host       string
	endpoint   string
	statusCode int
	duration   time.Duration
	isRetry    bool
}

func (a *attemptTracker) TrackAPIRequest(method, host, endpoint string, statusCode int, duration time.Duration, isRetry bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.attempts = append(a.attempts, attemptRecord{
		method:     method,
		host:       host,
		endpoint:   endpoint,
		statusCode: statusCode,
		duration:   duration,
		isRetry:    isRetry,
	})
}

func (a *attemptTracker) TrackAPIError(method, host, endpoint string, statusCode int, errorMessage string) {
}

func (a *attemptTracker) TrackAuthError(errorMessage string) {
}

func TestTrackingRoundTripper(t *testing.T) {
	t.Run("tracks single successful request", func(t *testing.T) {
		tracker := &attemptTracker{}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		transport := newTrackingRoundTripper(http.DefaultTransport, tracker)
		client := &http.Client{Transport: transport}

		req, err := http.NewRequest("GET", server.URL+"/test", nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		tracker.mu.Lock()
		defer tracker.mu.Unlock()

		require.Len(t, tracker.attempts, 1)
		require.False(t, tracker.attempts[0].isRetry)
		require.Equal(t, http.StatusOK, tracker.attempts[0].statusCode)
	})

	t.Run("tracks multiple attempts for same request", func(t *testing.T) {
		logger := log.NewLogger(log.WithDebugLog(true))
		tracker := &attemptTracker{}
		attemptCount := 0

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attemptCount++
			if attemptCount < 3 {
				w.WriteHeader(http.StatusTooManyRequests)
			} else {
				w.WriteHeader(http.StatusOK)
			}
		}))
		defer server.Close()

		retryableClient := NewRetryableHTTPClient(logger, tracker)

		req, err := http.NewRequest("GET", server.URL+"/test", nil)
		require.NoError(t, err)

		resp, err := retryableClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		tracker.mu.Lock()
		defer tracker.mu.Unlock()

		require.Len(t, tracker.attempts, 3, "Expected 3 attempts to be tracked")

		require.False(t, tracker.attempts[0].isRetry)
		require.Equal(t, http.StatusTooManyRequests, tracker.attempts[0].statusCode)

		require.True(t, tracker.attempts[1].isRetry)
		require.Equal(t, http.StatusTooManyRequests, tracker.attempts[1].statusCode)

		require.True(t, tracker.attempts[2].isRetry)
		require.Equal(t, http.StatusOK, tracker.attempts[2].statusCode)
	})
}
