package appstoreconnect

import (
	"context"
	"net/http"
	"time"
)

// trackingRoundTripper reports each HTTP attempt, retries included, as its own metric event,
// timed without the retry backoff, so alerts see real response times.
type trackingRoundTripper struct {
	wrapped http.RoundTripper
	tracker Tracker
}

type isRetryContextKey struct{}

func newTrackingRoundTripper(wrapped http.RoundTripper, tracker Tracker) *trackingRoundTripper {
	if wrapped == nil {
		wrapped = http.DefaultTransport
	}
	return &trackingRoundTripper{
		wrapped: wrapped,
		tracker: tracker,
	}
}

// A context value, not shared state: concurrent requests can hit the same endpoint.
func (t *trackingRoundTripper) markAsRetry(req *http.Request) *http.Request {
	ctx := context.WithValue(req.Context(), isRetryContextKey{}, true)
	return req.WithContext(ctx)
}

func (t *trackingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Check if this request was marked as a retry by RequestLogHook
	isRetry := req.Context().Value(isRetryContextKey{}) != nil

	startTime := time.Now()
	resp, err := t.wrapped.RoundTrip(req)
	duration := time.Since(startTime)

	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
	}

	// Track this attempt with its actual duration (no retry waits included)
	t.tracker.TrackAPIRequest(req.Method, req.URL.Host, req.URL.Path, statusCode, duration, isRetry)

	return resp, err
}
