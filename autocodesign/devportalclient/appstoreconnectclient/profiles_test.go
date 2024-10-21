package appstoreconnectclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/bitrise-io/go-xcode/v2/autocodesign"
	"github.com/bitrise-io/go-xcode/v2/autocodesign/devportalclient/appstoreconnect"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_checkBundleIDEntitlements(t *testing.T) {
	tests := []struct {
		name                 string
		bundleIDEntitlements []appstoreconnect.BundleIDCapability
		appEntitlements      autocodesign.Entitlements
		wantErr              bool
	}{
		{
			name:                 "Check known entitlements, which does not need to be registered on the Developer Portal",
			bundleIDEntitlements: []appstoreconnect.BundleIDCapability{},
			appEntitlements: autocodesign.Entitlements(map[string]interface{}{
				"keychain-access-groups":                             "",
				"com.apple.developer.ubiquity-kvstore-identifier":    "",
				"com.apple.developer.icloud-container-identifiers":   "",
				"com.apple.developer.ubiquity-container-identifiers": "",
			}),
			wantErr: false,
		},
		{
			name:                 "Needed to register entitlements",
			bundleIDEntitlements: []appstoreconnect.BundleIDCapability{},
			appEntitlements: autocodesign.Entitlements(map[string]interface{}{
				"com.apple.developer.applesignin": "",
			}),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkBundleIDEntitlements(tt.bundleIDEntitlements, tt.appEntitlements)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkBundleIDEntitlements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if mErr, ok := err.(autocodesign.NonmatchingProfileError); !ok {
					t.Errorf("checkBundleIDEntitlements() error = %v, it is not expected type", mErr)
				}
			}
		})
	}
}

func TestEnsureProfile_ExpiredProfile(t *testing.T) {
	// Arrange
	mockClient := &MockClient{}

	mockClient.
		On("PostProfilesFailed", mock.AnythingOfType("*http.Request")).
		Return(newResponse(t, http.StatusConflict,
			map[string]interface{}{
				"errors": []interface{}{map[string]interface{}{"detail": "ENTITY_ERROR: There is a problem with the request entity: Multiple profiles found with the name 'Bitrise iOS development - (io.bitrise.testapp)'.  Please remove the duplicate profiles and try again."}},
			}), nil)

	mockClient.
		On("GetBundleIDProfiles", mock.AnythingOfType("*http.Request")).
		Return(newResponse(t, http.StatusOK,
			map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"attributes": map[string]interface{}{"name": "Bitrise iOS development - (io.bitrise.testapp)"},
						"id":         "1",
					},
				}},
		), nil)

	mockClient.
		On("DeleteProfiles", mock.AnythingOfType("*http.Request")).
		Return(newResponse(t, http.StatusOK, map[string]interface{}{}), nil)

	mockClient.
		On("PostProfilesSuccess", mock.AnythingOfType("*http.Request")).
		Return(newResponse(t, http.StatusOK, map[string]interface{}{}), nil)

	client := appstoreconnect.NewClient(mockClient, "keyID", "issueID", []byte("privateKey"), false)
	profileClient := NewProfileClient(client)
	bundleID := appstoreconnect.BundleID{
		Attributes: appstoreconnect.BundleIDAttributes{Identifier: "io.bitrise.testapp"},
		Relationships: appstoreconnect.BundleIDRelationships{
			Profiles: appstoreconnect.RelationshipsLinks{
				Links: appstoreconnect.Links{
					Related: "https://api.appstoreconnect.apple.com/v1/bundleID/profiles",
				},
			},
			Capabilities: appstoreconnect.RelationshipsLinks{
				Links: appstoreconnect.Links{
					Related: "https://api.appstoreconnect.apple.com/v1/bundleID/capabilities",
				},
			},
		},
	}

	profile, err := profileClient.CreateProfile("Bitrise iOS development - (io.bitrise.testapp)", appstoreconnect.IOSAppDevelopment, bundleID, []string{}, []string{})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, profile)
	mockClient.AssertExpectations(t)
}

type MockClient struct {
	mock.Mock
	postProfileSuccess bool
}

func (c *MockClient) Do(req *http.Request) (*http.Response, error) {
	fmt.Printf("do called: %#v - %#v\n", req.Method, req.URL.Path)

	switch {
	case req.URL.Path == "/v1/profiles" && req.Method == "GET":
		return c.GetProfiles(req)
	case req.URL.Path == "/v1/profiles" && req.Method == "POST":
		// First profile create request fails by 'Multiple profiles found' error
		if !c.postProfileSuccess {
			c.postProfileSuccess = true
			return c.PostProfilesFailed(req)
		}
		// After deleting the expired profile, creating a new one succeed
		return c.PostProfilesSuccess(req)
	case req.URL.Path == "/v1//bundleID/capabilities" && req.Method == "GET":
		return c.GetBundleIDCapabilities(req)
	case req.URL.Path == "/v1//bundleID/profiles" && req.Method == "GET":
		return c.GetBundleIDProfiles(req)
	case req.URL.Path == "/v1/profiles/1" && req.Method == "DELETE":
		return c.DeleteProfiles(req)
	}

	return nil, fmt.Errorf("invalid endpoint called: %s, method: %s", req.URL.Path, req.Method)
}

func (c *MockClient) GetProfiles(req *http.Request) (*http.Response, error) {
	args := c.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func (c *MockClient) PostProfilesFailed(req *http.Request) (*http.Response, error) {
	args := c.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func (c *MockClient) GetBundleIDCapabilities(req *http.Request) (*http.Response, error) {
	args := c.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func (c *MockClient) GetBundleIDProfiles(req *http.Request) (*http.Response, error) {
	args := c.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func (c *MockClient) DeleteProfiles(req *http.Request) (*http.Response, error) {
	args := c.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func (c *MockClient) PostProfilesSuccess(req *http.Request) (*http.Response, error) {
	args := c.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func newResponse(t *testing.T, status int, body map[string]interface{}) *http.Response {
	resp := http.Response{
		StatusCode: status,
		Header:     http.Header{},
		Body:       io.NopCloser(nil),
	}

	if body != nil {
		var buff bytes.Buffer
		require.NoError(t, json.NewEncoder(&buff).Encode(body))
		resp.Body = io.NopCloser(&buff)
		resp.ContentLength = int64(buff.Len())
	}

	return &resp
}

func Test_wrapInProfileError(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantErrType error
	}{
		{
			err: appstoreconnect.ErrorResponse{
				Errors: []appstoreconnect.ErrorResponseError{{}},
			},
			wantErrType: appstoreconnect.ErrorResponse{},
		},
		{
			err: &appstoreconnect.ErrorResponse{
				Response: &http.Response{
					StatusCode: http.StatusNotFound,
				},
			},
			wantErrType: autocodesign.ProfilesInconsistentError{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := wrapInProfileError(tt.err)
			require.IsType(t, tt.wantErrType, err)
		})
	}
}
