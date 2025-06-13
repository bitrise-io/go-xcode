package appstoreconnect_tests

import (
	"testing"

	"github.com/bitrise-io/go-xcode/v2/autocodesign/devportalclient/appstoreconnect"
	"github.com/stretchr/testify/require"
)

func TestListBundleIDs(t *testing.T) {
	keyID, issuerID, privateKey, enterpriseAccount := getAPIKey(t)

	client := appstoreconnect.NewClient(appstoreconnect.NewRetryableHTTPClient(), keyID, issuerID, []byte(privateKey), enterpriseAccount, &appstoreconnect.NoOpAnalyticsTracker{})

	response, err := client.Provisioning.ListBundleIDs(&appstoreconnect.ListBundleIDsOptions{})
	require.NoError(t, err)
	require.True(t, len(response.Data) > 0)
}
