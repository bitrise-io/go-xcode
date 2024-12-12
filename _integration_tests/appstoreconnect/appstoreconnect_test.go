package appstoreconnect_test

import (
	"os"
	"testing"

	"github.com/bitrise-io/go-xcode/v2/autocodesign/devportalclient/appstoreconnect"
	"github.com/stretchr/testify/require"
)

func TestListBundleIDs(t *testing.T) {
	keyID := os.Getenv("TEST_KEY_ID")
	require.NotEmpty(t, keyID)
	issuerID := os.Getenv("TEST_ISSUER_ID")
	require.NotEmpty(t, issuerID)
	privateKey := os.Getenv("TEST_PRIVATE_KEY")
	require.NotEmpty(t, privateKey)
	enterpriseAccount := os.Getenv("TEST_ENTERPRISE_ACCOUNT") == "true"

	client := appstoreconnect.NewClient(appstoreconnect.NewRetryableHTTPClient(), keyID, issuerID, []byte(privateKey), enterpriseAccount)

	response, err := client.Provisioning.ListBundleIDs(&appstoreconnect.ListBundleIDsOptions{})
	require.NoError(t, err)
	require.True(t, len(response.Data) > 0)
}
