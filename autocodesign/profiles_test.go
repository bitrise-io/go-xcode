package autocodesign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	models "github.com/bitrise-io/go-xcode/autocodesign/codesignmodels"
	"github.com/bitrise-io/go-xcode/autocodesign/devportalclient/appstoreconnect"
	"github.com/bitrise-io/go-xcode/autocodesign/devportalclient/appstoreconnectclient"
	"github.com/bitrise-io/go-xcode/xcodeproject/serialized"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEnsureProfile_ExpiredProfile(t *testing.T) {
	// Arrange
	mockClient := &MockClient{}

	mockClient.
		On("GetProfiles", mock.AnythingOfType("*http.Request")).
		Return(newResponse(t, http.StatusOK, map[string]interface{}{}), nil)

	mockClient.
		On("PostProfilesFailed", mock.AnythingOfType("*http.Request")).
		Return(newResponse(t, http.StatusConflict,
			map[string]interface{}{
				"errors": []interface{}{map[string]interface{}{"detail": "ENTITY_ERROR: There is a problem with the request entity: Multiple profiles found with the name 'Bitrise iOS development - (io.bitrise.testapp)'.  Please remove the duplicate profiles and try again."}},
			}), nil)

	mockClient.
		On("GetBundleIDCapabilities", mock.AnythingOfType("*http.Request")).
		Return(newResponse(t, http.StatusOK, map[string]interface{}{}), nil)

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

	client := appstoreconnect.NewClient(mockClient, "keyID", "issueID", []byte("privateKey"))
	devportalClient := appstoreconnectclient.NewAPIDevportalClient(client)
	manager := profileManager{
		client: devportalClient.ProfileClient,
		// cache io.bitrise.testapp bundle ID, so that no need to mock bundle ID GET requests
		bundleIDByBundleIDIdentifer: map[string]*appstoreconnect.BundleID{"io.bitrise.testapp": {
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
		}},
		containersByBundleID: nil}

	// Act
	profile, err := manager.ensureProfile(
		appstoreconnect.IOSAppDevelopment,
		"io.bitrise.testapp",
		serialized.Object(map[string]interface{}{}),
		[]string{},
		[]string{},
		0,
	)

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
		Body:       ioutil.NopCloser(nil),
	}

	if body != nil {
		var buff bytes.Buffer
		require.NoError(t, json.NewEncoder(&buff).Encode(body))
		resp.Body = ioutil.NopCloser(&buff)
		resp.ContentLength = int64(buff.Len())
	}

	return &resp
}

func Test_createWildcardBundleID(t *testing.T) {
	tests := []struct {
		name     string
		bundleID string
		want     string
		wantErr  bool
	}{
		{
			name:     "Invalid bundle id: empty",
			bundleID: "",
			want:     "",
			wantErr:  true,
		},
		{
			name:     "Invalid bundle id: does not contain *",
			bundleID: "my_app",
			want:     "",
			wantErr:  true,
		},
		{
			name:     "2 component bundle id",
			bundleID: "com.my_app",
			want:     "com.*",
			wantErr:  false,
		},
		{
			name:     "multi component bundle id",
			bundleID: "com.bitrise.my_app.uitest",
			want:     "com.bitrise.my_app.*",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createWildcardBundleID(tt.bundleID)
			if (err != nil) != tt.wantErr {
				t.Errorf("createWildcardBundleID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("createWildcardBundleID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_profileName(t *testing.T) {
	tests := []struct {
		profileType appstoreconnect.ProfileType
		bundleID    string
		want        string
		wantErr     bool
	}{
		{
			profileType: appstoreconnect.IOSAppDevelopment,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise iOS development - (io.bitrise.app)",
			wantErr:     false,
		},
		{
			profileType: appstoreconnect.IOSAppStore,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise iOS app-store - (io.bitrise.app)",
			wantErr:     false,
		},
		{
			profileType: appstoreconnect.IOSAppAdHoc,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise iOS ad-hoc - (io.bitrise.app)",
			wantErr:     false,
		},
		{
			profileType: appstoreconnect.IOSAppInHouse,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise iOS enterprise - (io.bitrise.app)",
			wantErr:     false,
		},

		{
			profileType: appstoreconnect.TvOSAppDevelopment,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise tvOS development - (io.bitrise.app)",
			wantErr:     false,
		},
		{
			profileType: appstoreconnect.TvOSAppStore,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise tvOS app-store - (io.bitrise.app)",
			wantErr:     false,
		},
		{
			profileType: appstoreconnect.TvOSAppAdHoc,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise tvOS ad-hoc - (io.bitrise.app)",
			wantErr:     false,
		},
		{
			profileType: appstoreconnect.TvOSAppInHouse,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise tvOS enterprise - (io.bitrise.app)",
			wantErr:     false,
		},
		{
			profileType: appstoreconnect.ProfileType("unknown"),
			bundleID:    "io.bitrise.app",
			want:        "",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(string(tt.profileType), func(t *testing.T) {
			got, err := profileName(tt.profileType, tt.bundleID)
			if (err != nil) != tt.wantErr {
				t.Errorf("profileName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("profileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findMissingContainers(t *testing.T) {
	tests := []struct {
		name        string
		projectEnts serialized.Object
		profileEnts serialized.Object
		want        []string
		wantErr     bool
	}{
		{
			name: "equal without container",
			projectEnts: serialized.Object(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": []interface{}{},
			}),
			profileEnts: serialized.Object(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": []interface{}{},
			}),

			want:    nil,
			wantErr: false,
		},
		{
			name: "equal with container",
			projectEnts: serialized.Object(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": []interface{}{"container1"},
			}),
			profileEnts: serialized.Object(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": []interface{}{"container1"},
			}),

			want:    nil,
			wantErr: false,
		},
		{
			name: "profile has more containers than project",
			projectEnts: serialized.Object(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": []interface{}{},
			}),
			profileEnts: serialized.Object(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": []interface{}{"container1"},
			}),

			want:    nil,
			wantErr: false,
		},
		{
			name: "project has more containers than profile",
			projectEnts: serialized.Object(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": []interface{}{"container1"},
			}),
			profileEnts: serialized.Object(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": []interface{}{},
			}),

			want:    []string{"container1"},
			wantErr: false,
		},
		{
			name: "project has containers but profile doesn't",
			projectEnts: serialized.Object(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": []interface{}{"container1"},
			}),
			profileEnts: serialized.Object(map[string]interface{}{
				"otherentitlement": "",
			}),

			want:    []string{"container1"},
			wantErr: false,
		},
		{
			name: "error check",
			projectEnts: serialized.Object(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": "break",
			}),

			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findMissingContainers(tt.projectEnts, tt.profileEnts)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, got, tt.want)
		})
	}
}

type MockProfile struct {
	attributes appstoreconnect.ProfileAttributes
}

func (m MockProfile) ID() string {
	return ""
}

func (m MockProfile) Attributes() appstoreconnect.ProfileAttributes {
	return m.attributes
}

func (m MockProfile) CertificateIDs() (map[string]bool, error) {
	return nil, nil
}

func (m MockProfile) DeviceIDs() (map[string]bool, error) {
	return nil, nil
}

func (m MockProfile) BundleID() (appstoreconnect.BundleID, error) {
	return appstoreconnect.BundleID{}, nil
}

func Test_IsProfileExpired(t *testing.T) {
	tests := []struct {
		prof                models.Profile
		minProfileDaysValid int
		name                string
		want                bool
	}{
		{
			name:                "no days set - profile expiry date after current time",
			minProfileDaysValid: 0,
			prof:                MockProfile{attributes: appstoreconnect.ProfileAttributes{ExpirationDate: appstoreconnect.Time(time.Now().Add(5 * time.Hour))}},
			want:                false,
		},
		{
			name:                "no days set - profile expiry date before current time",
			minProfileDaysValid: 0,
			prof:                MockProfile{attributes: appstoreconnect.ProfileAttributes{ExpirationDate: appstoreconnect.Time(time.Now().Add(-5 * time.Hour))}},
			want:                true,
		},
		{
			name:                "days set - profile expiry date after current time + days set",
			minProfileDaysValid: 2,
			prof:                MockProfile{attributes: appstoreconnect.ProfileAttributes{ExpirationDate: appstoreconnect.Time(time.Now().Add(5 * 24 * time.Hour))}},
			want:                false,
		},
		{
			name:                "days set - profile expiry date before current time + days set",
			minProfileDaysValid: 2,
			prof:                MockProfile{attributes: appstoreconnect.ProfileAttributes{ExpirationDate: appstoreconnect.Time(time.Now().Add(1 * 24 * time.Hour))}},
			want:                true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isProfileExpired(tt.prof, tt.minProfileDaysValid); got != tt.want {
				t.Errorf("checkProfileExpiry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanGenerateProfileWithEntitlements(t *testing.T) {
	tests := []struct {
		name                   string
		entitlementsByBundleID map[string]serialized.Object
		wantOk                 bool
		wantEntitlement        string
		wantBundleID           string
	}{
		{
			name: "no entitlements",
			entitlementsByBundleID: map[string]serialized.Object{
				"com.bundleid": map[string]interface{}{},
			},
			wantOk:          true,
			wantEntitlement: "",
			wantBundleID:    "",
		},
		{
			name: "contains unsupported entitlement",
			entitlementsByBundleID: map[string]serialized.Object{
				"com.bundleid": map[string]interface{}{
					"com.entitlement-ignored":            true,
					"com.apple.developer.contacts.notes": true,
				},
			},
			wantOk:          false,
			wantEntitlement: "com.apple.developer.contacts.notes",
			wantBundleID:    "com.bundleid",
		},
		{
			name: "contains unsupported entitlement, multiple bundle IDs",
			entitlementsByBundleID: map[string]serialized.Object{
				"com.bundleid": map[string]interface{}{
					"aps-environment": true,
				},
				"com.bundleid2": map[string]interface{}{
					"com.entitlement-ignored":            true,
					"com.apple.developer.contacts.notes": true,
				},
			},
			wantOk:          false,
			wantEntitlement: "com.apple.developer.contacts.notes",
			wantBundleID:    "com.bundleid2",
		},
		{
			name: "all entitlements supported",
			entitlementsByBundleID: map[string]serialized.Object{
				"com.bundleid": map[string]interface{}{
					"aps-environment": true,
				},
			},
			wantOk:          true,
			wantEntitlement: "",
			wantBundleID:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOk, gotEntilement, gotBundleID := CanGenerateProfileWithEntitlements(tt.entitlementsByBundleID)
			if gotOk != tt.wantOk {
				t.Errorf("CanGenerateProfileWithEntitlements() got = %v, want %v", gotOk, tt.wantOk)
			}
			if gotEntilement != tt.wantEntitlement {
				t.Errorf("CanGenerateProfileWithEntitlements() got1 = %v, want %v", gotEntilement, tt.wantEntitlement)
			}
			if gotBundleID != tt.wantBundleID {
				t.Errorf("CanGenerateProfileWithEntitlements() got2 = %v, want %v", gotBundleID, tt.wantBundleID)
			}
		})
	}
}
