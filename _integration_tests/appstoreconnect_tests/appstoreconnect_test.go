package appstoreconnect_tests

import (
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/autocodesign/devportalclient/appstoreconnect"
	"github.com/stretchr/testify/require"
)

func TestListBundleIDs(t *testing.T) {
	keyID, issuerID, privateKey, enterpriseAccount := getAPIKey(t)

	logger := log.NewLogger(log.WithDebugLog(true))
	tracker := appstoreconnect.NoOpAnalyticsTracker{}
	client := appstoreconnect.NewClient(appstoreconnect.NewRetryableHTTPClient(logger, tracker), keyID, issuerID, []byte(privateKey), enterpriseAccount, logger, tracker)

	response, err := client.Provisioning.ListBundleIDs(&appstoreconnect.ListBundleIDsOptions{})
	require.NoError(t, err)
	require.True(t, len(response.Data) > 0)
}
