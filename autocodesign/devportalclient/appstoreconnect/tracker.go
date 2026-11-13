package appstoreconnect

import (
	"time"

	"github.com/bitrise-io/go-utils/v2/analytics"
	"github.com/bitrise-io/go-utils/v2/env"
)

type Tracker interface {
	// TrackAPIRequest tracks one HTTP request+response. This is called for each individual attempt in case of automatic retries.
	TrackAPIRequest(method, host, endpoint string, statusCode int, duration time.Duration, isRetry bool)

	TrackAPIError(method, host, endpoint string, statusCode int, errorMessage string)

	TrackAuthError(errorMessage string)

	// TrackUnknownEntitlement is fired when a project uses an entitlement key
	// not in ServiceTypeByKey. Feeds alerting so we can register newly
	// introduced Apple entitlements before they become customer-facing errors.
	TrackUnknownEntitlement(key string)
}

// NoOpAnalyticsTracker is a dummy implementation used in tests.
type NoOpAnalyticsTracker struct{}

func (n NoOpAnalyticsTracker) TrackAPIRequest(method, host, endpoint string, statusCode int, duration time.Duration, isRetry bool) {
}

func (n NoOpAnalyticsTracker) TrackAPIError(method, host, endpoint string, statusCode int, errorMessage string) {
}

func (n NoOpAnalyticsTracker) TrackAuthError(errorMessage string) {}

func (n NoOpAnalyticsTracker) TrackUnknownEntitlement(key string) {}

type DefaultTracker struct {
	tracker analytics.Tracker
	envRepo env.Repository
}

func NewDefaultTracker(tracker analytics.Tracker, envRepo env.Repository) *DefaultTracker {
	return &DefaultTracker{
		tracker: tracker,
		envRepo: envRepo,
	}
}

func (d *DefaultTracker) TrackAPIRequest(method, host, endpoint string, statusCode int, duration time.Duration, isRetry bool) {
	d.tracker.Enqueue("step_appstoreconnect_request", analytics.Properties{
		"build_slug":        d.envRepo.Get("BITRISE_BUILD_SLUG"),
		"step_execution_id": d.envRepo.Get("BITRISE_STEP_EXECUTION_ID"),
		"http_method":       method,
		"host":              host, // Regular, enterprise, or any future third option
		"endpoint":          endpoint,
		"status_code":       statusCode,
		"duration_ms":       duration.Truncate(time.Millisecond).Milliseconds(),
		"is_retry":          isRetry,
	})
}

func (d *DefaultTracker) TrackAPIError(method, host, endpoint string, statusCode int, errorMessage string) {
	d.tracker.Enqueue("step_appstoreconnect_error", analytics.Properties{
		"build_slug":        d.envRepo.Get("BITRISE_BUILD_SLUG"),
		"step_execution_id": d.envRepo.Get("BITRISE_STEP_EXECUTION_ID"),
		"http_method":       method,
		"host":              host, // Regular, enterprise, or any future third option
		"endpoint":          endpoint,
		"status_code":       statusCode,
		"error_message":     errorMessage,
	})
}

func (d *DefaultTracker) TrackAuthError(errorMessage string) {
	d.tracker.Enqueue("step_appstoreconnect_auth_error", analytics.Properties{
		"build_slug":        d.envRepo.Get("BITRISE_BUILD_SLUG"),
		"step_execution_id": d.envRepo.Get("BITRISE_STEP_EXECUTION_ID"),
		"error_message":     errorMessage,
	})
}

func (d *DefaultTracker) TrackUnknownEntitlement(key string) {
	d.tracker.Enqueue("step_appstoreconnect_unknown_entitlement", analytics.Properties{
		"build_slug":        d.envRepo.Get("BITRISE_BUILD_SLUG"),
		"step_execution_id": d.envRepo.Get("BITRISE_STEP_EXECUTION_ID"),
		"entitlement_key":   key,
	})
}
