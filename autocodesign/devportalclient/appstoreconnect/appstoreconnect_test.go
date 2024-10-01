// Package appstoreconnect implements a client for the App Store Connect API.
//
// It contains type definitions, authentication and API calls, without business logic built on those API calls.
package appstoreconnect

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	got := NewClient(NewRetryableHTTPClient(), "keyID", "issuerID", []byte{})

	require.Equal(t, "appstoreconnect-v1", got.audience)

	wantURL, err := url.Parse("https://api.appstoreconnect.apple.com/")
	require.NoError(t, err)
	require.Equal(t, wantURL, got.BaseURL)
}

func TestNewEnterpriseClient(t *testing.T) {
	got := NewEnterpriseClient(NewRetryableHTTPClient(), "keyID", "issuerID", []byte{})

	require.Equal(t, "apple-developer-enterprise-v1", got.audience)

	wantURL, err := url.Parse("https://api.enterprise.developer.apple.com/")
	require.NoError(t, err)
	require.Equal(t, wantURL, got.BaseURL)
}
