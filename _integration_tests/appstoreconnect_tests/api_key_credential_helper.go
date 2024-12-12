package appstoreconnect_tests

import (
	"io"
	"os"
	"testing"

	"github.com/bitrise-io/go-xcode/v2/_integration_tests"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/require"
)

func getAPIKey(t *testing.T) (string, string, []byte, bool) {
	if os.Getenv("TEST_API_KEY") != "" {
		return getLocalAPIKey(t)
	}
	return getRemoteAPIKey(t)
}

func getLocalAPIKey(t *testing.T) (string, string, []byte, bool) {
	keyID := os.Getenv("TEST_API_KEY_ID")
	require.NotEmpty(t, keyID)
	issuerID := os.Getenv("TEST_API_KEY_ISSUER_ID")
	require.NotEmpty(t, issuerID)
	privateKey := os.Getenv("TEST_API_KEY")
	require.NotEmpty(t, privateKey)
	isEnterpriseAPIKey := os.Getenv("TEST_API_KEY_IS_ENTERPRISE") == "true"

	return keyID, issuerID, []byte(privateKey), isEnterpriseAPIKey
}

func getRemoteAPIKey(t *testing.T) (string, string, []byte, bool) {
	serviceAccountJSON := os.Getenv("GCS_SERVICE_ACCOUNT_JSON")
	require.NotEmpty(t, serviceAccountJSON)
	projectID := os.Getenv("GCS_PROJECT_ID")
	require.NotEmpty(t, projectID)
	bucketName := os.Getenv("GCS_BUCKET_NAME")
	require.NotEmpty(t, bucketName)

	secretAccessor, err := _integration_tests.NewSecretAccessor(serviceAccountJSON, projectID)
	require.NoError(t, err)

	bucketAccessor, err := _integration_tests.NewBucketAccessor(serviceAccountJSON, bucketName)
	require.NoError(t, err)

	keyID, err := secretAccessor.GetSecret("BITRISE_APPSTORECONNECT_API_KEY_ID")
	require.NoError(t, err)

	issuerID, err := secretAccessor.GetSecret("BITRISE_APPSTORECONNECT_API_KEY_ISSUER_ID")
	require.NoError(t, err)

	keyURL, err := secretAccessor.GetSecret("BITRISE_APPSTORECONNECT_API_KEY_URL")
	require.NoError(t, err)

	keyDownloadURL, err := bucketAccessor.GetExpiringURL(keyURL)
	require.NoError(t, err)

	client := retryablehttp.NewClient()
	resp, err := client.Get(keyDownloadURL)
	require.NoError(t, err)

	privateKey, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return keyID, issuerID, privateKey, false
}
