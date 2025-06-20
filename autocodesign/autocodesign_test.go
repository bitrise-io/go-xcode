package autocodesign

import (
	"errors"
	"testing"
	"time"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-xcode/v2/autocodesign/devportalclient/appstoreconnect"
	devportaltime "github.com/bitrise-io/go-xcode/v2/autocodesign/devportalclient/time"
	"github.com/bitrise-io/go-xcode/v2/certificateutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newDefaultMockAssetWriter() AssetWriter {
	mockAssetWriter := new(MockAssetWriter)
	mockAssetWriter.On("Write", mock.Anything).Return(nil)
	mockAssetWriter.On("InstallCertificate", mock.Anything).Return(nil)

	return mockAssetWriter
}

func newMockLocalCodeSignAssetManager(assets *AppCodesignAssets, missingAppLayout *AppLayout) LocalCodeSignAssetManager {
	mockLocalCodeSignAssetManager := new(MockLocalCodeSignAssetManager)
	mockLocalCodeSignAssetManager.On("FindCodesignAssets", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(assets, missingAppLayout, nil)

	return mockLocalCodeSignAssetManager
}

type profileArgs struct {
	attributes   appstoreconnect.ProfileAttributes
	id           string
	appID        appstoreconnect.BundleID
	devices      []string
	certificates []string
	entitlements Entitlements
}

func newMockProfile(m profileArgs) Profile {
	profile := new(MockProfile)
	profile.On("Attributes").Return(func() appstoreconnect.ProfileAttributes {
		return m.attributes
	})
	profile.On("ID").Return(func() string {
		return m.id
	})
	profile.On("BundleID").Return(func() appstoreconnect.BundleID {
		return m.appID
	}, nil)
	profile.On("DeviceUDIDs").Return(func() []string {
		return m.devices
	}, nil)
	profile.On("CertificateIDs").Return(func() []string {
		return m.certificates
	}, nil)
	profile.On("Entitlements").Return(func() Entitlements {
		return m.entitlements
	}, nil)

	return profile
}

func newCertificate(t *testing.T, teamID, teamName, commonName string, expiry time.Time) certificateutil.CertificateInfo {
	cert, privateKey, err := certificateutil.GenerateTestCertificate(int64(1), teamID, teamName, commonName, expiry)
	if err != nil {
		t.Fatalf("init: failed to generate certificate: %s", err)
	}
	return certificateutil.NewCertificateInfo(*cert, privateKey)
}

func Test_codesignAssetManager_EnsureCodesignAssets(t *testing.T) {
	log.SetEnableDebugLog(true)

	const teamID = "MYTEAMID"
	const commonNameIOSDevelopment = "Apple Development: test"
	const teamName = "BITFALL FEJLESZTO KORLATOLT FELELOSSEGU TARSASAG"
	expiry := time.Now().AddDate(1, 0, 0)
	devCert := newCertificate(t, teamID, teamName, commonNameIOSDevelopment, expiry)

	t.Logf("Test certificate generated. %s", devCert)

	devProfile := newMockProfile(profileArgs{
		attributes: appstoreconnect.ProfileAttributes{
			Name:           profileName(appstoreconnect.IOSAppDevelopment, "io.test"),
			ProfileState:   appstoreconnect.Active,
			ExpirationDate: devportaltime.Time(expiry),
		},
		certificates: []string{"dev1"},
	})

	checkOnlyDevportalProfile := newMockDevportalClient(devportalArgs{
		certs: map[appstoreconnect.CertificateType][]Certificate{
			appstoreconnect.IOSDevelopment: {{
				CertificateInfo: devCert,
				ID:              "dev1",
			}},
		},
		profiles: map[appstoreconnect.ProfileType][]Profile{
			appstoreconnect.IOSAppDevelopment: {
				devProfile,
			},
		},
		appIDs: []appstoreconnect.BundleID{{
			Attributes: appstoreconnect.BundleIDAttributes{
				Identifier: "io.test",
				Name:       "test-app",
			},
		}},
	})
	checkOnlyDevportalProfile.On("CheckBundleIDEntitlements", mock.Anything, mock.Anything).Return(nil)

	createdAppID := &appstoreconnect.BundleID{
		ID: "app1",
		Attributes: appstoreconnect.BundleIDAttributes{
			Identifier: "io.test",
			Name:       "Bitrise io test",
		},
	}

	devportalWithNoAppID := newMockDevportalClient(devportalArgs{
		certs: map[appstoreconnect.CertificateType][]Certificate{
			appstoreconnect.IOSDevelopment: {{
				CertificateInfo: devCert,
				ID:              "dev1",
			}},
		},
		profiles: map[appstoreconnect.ProfileType][]Profile{
			appstoreconnect.IOSAppDevelopment: {},
		},
		appIDs: []appstoreconnect.BundleID{},
	})
	devportalWithNoAppID.On("CreateBundleID", "io.test", "Bitrise io test").Return(createdAppID, nil).
		On("SyncBundleID", *createdAppID, mock.Anything).Return(nil).
		On("CreateProfile", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(newMockProfile(profileArgs{}), nil)

	appIDAndProfileFoundAppLayout := AppLayout{
		Platform: IOS,
		EntitlementsByArchivableTargetBundleID: map[string]Entitlements{
			"io.test": {},
		},
	}

	appIDAndProfileFoundAppLayout2 := AppLayout{
		Platform: IOS,
		EntitlementsByArchivableTargetBundleID: map[string]Entitlements{
			"io.test":             {},
			"io.test.development": {},
		},
	}

	localCodeSignAsset := AppCodesignAssets{
		ArchivableTargetProfilesByBundleID: map[string]Profile{
			"io.test.development": devProfile,
		},
		UITestTargetProfilesByBundleID: map[string]Profile{},
		Certificate:                    devCert,
	}

	icloudContainerAppLayout := AppLayout{
		Platform: IOS,
		EntitlementsByArchivableTargetBundleID: map[string]Entitlements{
			"io.test": map[string]interface{}{
				"com.apple.developer.icloud-services": []interface{}{
					"CloudDocuments",
				},
				"com.apple.developer.icloud-container-identifiers": []interface{}{
					"iCloud.test.container.id",
				},
			},
		},
	}

	type fields struct {
		devPortalClient           DevPortalClient
		assetWriter               AssetWriter
		localCodeSignAssetManager LocalCodeSignAssetManager
	}
	tests := []struct {
		name      string
		fields    fields
		appLayout AppLayout
		opts      CodesignAssetsOpts
		want      map[DistributionType]AppCodesignAssets
		wantErr   *DetailedError
	}{
		{
			name: "no valid certs found",
			fields: fields{
				devPortalClient: newMockDevportalClient(devportalArgs{}),
			},
			opts: CodesignAssetsOpts{
				DistributionType:        Development,
				TypeToLocalCertificates: LocalCertificates{},
			},
			wantErr: &DetailedError{},
		},
		{
			name: "App ID and Profile found, valid",
			fields: fields{
				devPortalClient:           checkOnlyDevportalProfile,
				assetWriter:               newDefaultMockAssetWriter(),
				localCodeSignAssetManager: newMockLocalCodeSignAssetManager(nil, &appIDAndProfileFoundAppLayout),
			},
			appLayout: appIDAndProfileFoundAppLayout,
			opts: CodesignAssetsOpts{
				DistributionType: Development,
				TypeToLocalCertificates: LocalCertificates{
					appstoreconnect.IOSDevelopment: {devCert},
				},
			},
			want: map[DistributionType]AppCodesignAssets{
				Development: {
					ArchivableTargetProfilesByBundleID: map[string]Profile{
						"io.test": devProfile,
					},
					UITestTargetProfilesByBundleID: map[string]Profile{},
					Certificate:                    devCert,
				},
			},
			wantErr: nil,
		},
		{
			name: "Codesign assets are merged",
			fields: fields{
				devPortalClient:           checkOnlyDevportalProfile,
				assetWriter:               newDefaultMockAssetWriter(),
				localCodeSignAssetManager: newMockLocalCodeSignAssetManager(&localCodeSignAsset, &appIDAndProfileFoundAppLayout),
			},
			appLayout: appIDAndProfileFoundAppLayout2,
			opts: CodesignAssetsOpts{
				DistributionType: Development,
				TypeToLocalCertificates: LocalCertificates{
					appstoreconnect.IOSDevelopment: {devCert},
				},
			},
			want: map[DistributionType]AppCodesignAssets{
				Development: {
					ArchivableTargetProfilesByBundleID: map[string]Profile{
						"io.test":             devProfile,
						"io.test.development": devProfile,
					},
					UITestTargetProfilesByBundleID: map[string]Profile{},
					Certificate:                    devCert,
				},
			},
			wantErr: nil,
		},
		{
			name: "can not create iCloud containers",
			fields: fields{
				devPortalClient:           devportalWithNoAppID,
				localCodeSignAssetManager: newMockLocalCodeSignAssetManager(nil, &icloudContainerAppLayout),
			},
			appLayout: icloudContainerAppLayout,
			opts: CodesignAssetsOpts{
				DistributionType: Development,
				TypeToLocalCertificates: LocalCertificates{
					appstoreconnect.IOSDevelopment: {devCert},
				},
			},
			wantErr: &DetailedError{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := codesignAssetManager{
				devPortalClient:           tt.fields.devPortalClient,
				assetWriter:               tt.fields.assetWriter,
				localCodeSignAssetManager: tt.fields.localCodeSignAssetManager,
			}

			got, err := m.EnsureCodesignAssets(tt.appLayout, tt.opts)

			if ((tt.wantErr == nil) && (err != nil)) ||
				(tt.wantErr != nil && !errors.As(err, &tt.wantErr)) {
				t.Errorf("codesignAssetManager.EnsureCodesignAssets() got type = %T want type = %T got error: %s", err, tt.wantErr, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_GivenNoValidAppID_WhenEnsureAppClipProfile_ThenItFails(t *testing.T) {
	// Given
	const teamID = "MY_TEAM_ID"
	expiry := time.Now().AddDate(1, 0, 0)
	devCert := newCertificate(t, teamID, "MY_TEAM", "Apple Development: test", expiry)

	client := newClientWithoutAppIDAndProfile(devCert)
	assetWriter := newDefaultMockAssetWriter()

	appLayout := AppLayout{
		Platform: IOS,
		EntitlementsByArchivableTargetBundleID: map[string]Entitlements{
			"io.bitrise.appclip": {"com.apple.developer.parent-application-identifiers": []string{"io.bitrise.app"}},
		},
	}

	localCodeSignAssetManager := newMockLocalCodeSignAssetManager(nil, &appLayout)
	manager := NewCodesignAssetManager(client, assetWriter, localCodeSignAssetManager)

	opts := CodesignAssetsOpts{
		DistributionType: Development,
		TypeToLocalCertificates: LocalCertificates{
			appstoreconnect.IOSDevelopment: {devCert},
		},
	}

	// When
	_, err := manager.EnsureCodesignAssets(appLayout, opts)

	// Then
	require.ErrorAs(t, err, &ErrAppClipAppID{})
}

func Test_GivenAppIDWithoutAppleSignIn_WhenEnsureAppClipProfile_ThenItFails(t *testing.T) {
	// Given
	const teamID = "MY_TEAM_ID"
	const appClipBundleID = "io.bitrise.appclip"

	expiry := time.Now().AddDate(1, 0, 0)
	devCert := newCertificate(t, teamID, "MY_TEAM", "Apple Development: test", expiry)

	client := newClientWithAppIDWithoutAppleSignIn(devCert, appClipBundleID)
	assetWriter := newDefaultMockAssetWriter()

	appLayout := AppLayout{
		Platform: IOS,
		EntitlementsByArchivableTargetBundleID: map[string]Entitlements{
			appClipBundleID: {
				"com.apple.developer.parent-application-identifiers": []string{"io.bitrise.app"},
				"com.apple.developer.applesignin":                    []string{"Default"},
			},
		},
	}

	localCodeSignAssetManager := newMockLocalCodeSignAssetManager(nil, &appLayout)
	manager := NewCodesignAssetManager(client, assetWriter, localCodeSignAssetManager)

	opts := CodesignAssetsOpts{
		DistributionType: Development,
		TypeToLocalCertificates: LocalCertificates{
			appstoreconnect.IOSDevelopment: {devCert},
		},
	}

	// When
	_, err := manager.EnsureCodesignAssets(appLayout, opts)

	// Then
	require.ErrorAs(t, err, &ErrAppClipAppIDWithAppleSigning{})
}

func Test_GivenProfileExpired_WhenProfilesInconsistent_ThenItRetries(t *testing.T) {
	// Given
	const teamID = "MY_TEAM_ID"
	expiry := time.Now().AddDate(1, 0, 0)
	devCert := newCertificate(t, teamID, "MY_TEAM", "Apple Development: test", expiry)

	expiredProfile := newMockProfile(profileArgs{
		attributes: appstoreconnect.ProfileAttributes{
			Name:           profileName(appstoreconnect.IOSAppDevelopment, "io.test"),
			ProfileState:   appstoreconnect.Active,
			ExpirationDate: devportaltime.Time(time.Now().AddDate(0, -1, 0)),
		},
		certificates: []string{"dev1"},
	})
	validProfile := newMockProfile(profileArgs{
		attributes: appstoreconnect.ProfileAttributes{
			Name:           profileName(appstoreconnect.IOSAppDevelopment, "io.test"),
			ProfileState:   appstoreconnect.Active,
			ExpirationDate: devportaltime.Time(expiry),
		},
		certificates: []string{"dev1"},
	})

	client := newMockDevportalClient(devportalArgs{
		certs: map[appstoreconnect.CertificateType][]Certificate{
			appstoreconnect.IOSDevelopment: {{
				CertificateInfo: devCert,
				ID:              "dev1",
			}},
		},
		profiles: map[appstoreconnect.ProfileType][]Profile{
			appstoreconnect.IOSAppDevelopment: {expiredProfile},
		},
		appIDs: []appstoreconnect.BundleID{{
			Attributes: appstoreconnect.BundleIDAttributes{
				Identifier: "io.test",
				Name:       "test-app",
			},
		}},
	})
	// FindProfile
	client.On("DeleteProfile", expiredProfile.ID()).Return(nil).Once()
	client.On("CheckBundleIDEntitlements", mock.Anything, mock.Anything).Return(nil).Once()
	client.On("CreateProfile", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, ProfilesInconsistentError{}).Once()
	// FindProfile
	client.On("DeleteProfile", expiredProfile.ID()).Return(nil).Once()
	client.On("CheckBundleIDEntitlements", mock.Anything, mock.Anything).Return(nil).Once()
	client.On("CreateProfile", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(validProfile, nil).Once()

	assetWriter := newDefaultMockAssetWriter()
	appLayout := AppLayout{
		Platform: IOS,
		EntitlementsByArchivableTargetBundleID: map[string]Entitlements{
			"io.test": {},
		},
	}

	localCodeSignAssetManager := newMockLocalCodeSignAssetManager(nil, &appLayout)
	manager := NewCodesignAssetManager(client, assetWriter, localCodeSignAssetManager)

	opts := CodesignAssetsOpts{
		DistributionType: Development,
		TypeToLocalCertificates: LocalCertificates{
			appstoreconnect.IOSDevelopment: {devCert},
		},
	}

	// When
	_, err := manager.EnsureCodesignAssets(appLayout, opts)

	// Then
	require.NoError(t, err)
}

func Test_GivenLocalProfile_WhenCertificateIsMissing_ThenInstalled(t *testing.T) {
	// Given
	const teamID = "MY_TEAM_ID"
	expiry := time.Now().AddDate(1, 0, 0)
	devCert1 := newCertificate(t, teamID, "MY_TEAM", "Apple Development: test 1", expiry)
	devCert2 := newCertificate(t, teamID, "MY_TEAM", "Apple Development: test 2", expiry)

	validProfile := newMockProfile(profileArgs{
		attributes: appstoreconnect.ProfileAttributes{
			Name:           profileName(appstoreconnect.IOSAppDevelopment, "io.test"),
			ProfileState:   appstoreconnect.Active,
			ExpirationDate: devportaltime.Time(expiry),
		},
		certificates: []string{"dev1", "dev2"},
	})

	client := newMockDevportalClient(devportalArgs{
		certs: map[appstoreconnect.CertificateType][]Certificate{
			appstoreconnect.IOSDevelopment: {
				{
					CertificateInfo: devCert1,
					ID:              "dev1",
				},
				{
					CertificateInfo: devCert2,
					ID:              "dev2",
				},
			},
		},
		profiles: map[appstoreconnect.ProfileType][]Profile{
			appstoreconnect.IOSAppDevelopment: {validProfile},
		},
		appIDs: []appstoreconnect.BundleID{{
			Attributes: appstoreconnect.BundleIDAttributes{
				Identifier: "io.test",
				Name:       "test-app",
			},
		}},
	})

	assetWriter := new(MockAssetWriter)
	assetWriter.On("Write", mock.Anything).Return(nil)
	assetWriter.On("InstallCertificate", devCert2).Return(nil).Once()
	appLayout := AppLayout{
		Platform: IOS,
		EntitlementsByArchivableTargetBundleID: map[string]Entitlements{
			"io.test": {},
		},
	}

	localCodeSignAssetManager := newMockLocalCodeSignAssetManager(&AppCodesignAssets{
		ArchivableTargetProfilesByBundleID: map[string]Profile{
			"io.test": validProfile,
		},
		Certificate: devCert2,
	}, nil)
	manager := NewCodesignAssetManager(client, assetWriter, localCodeSignAssetManager)

	opts := CodesignAssetsOpts{
		DistributionType: Development,
		TypeToLocalCertificates: LocalCertificates{
			appstoreconnect.IOSDevelopment: {devCert1},
		},
	}

	wantAssets := map[DistributionType]AppCodesignAssets{
		Development: {
			ArchivableTargetProfilesByBundleID: map[string]Profile{
				"io.test": validProfile,
			},
			Certificate: devCert2,
		},
	}

	// When
	gotAssets, err := manager.EnsureCodesignAssets(appLayout, opts)

	// Then
	require.NoError(t, err)
	require.Equal(t, wantAssets, gotAssets)
}

func newClientWithoutAppIDAndProfile(cert certificateutil.CertificateInfo) *MockDevPortalClient {
	client := newMockDevportalClient(devportalArgs{
		certs: map[appstoreconnect.CertificateType][]Certificate{
			appstoreconnect.IOSDevelopment: {{
				CertificateInfo: cert,
				ID:              "dev1",
			}},
		},
		profiles: map[appstoreconnect.ProfileType][]Profile{
			appstoreconnect.IOSAppDevelopment: {},
		},
		appIDs: []appstoreconnect.BundleID{},
	})

	return client
}

func newClientWithAppIDWithoutAppleSignIn(cert certificateutil.CertificateInfo, bundleID string) *MockDevPortalClient {
	appID := appstoreconnect.BundleID{
		Attributes: appstoreconnect.BundleIDAttributes{
			Identifier: bundleID,
			Name:       "test-app",
		},
	}

	client := newMockDevportalClient(devportalArgs{
		certs: map[appstoreconnect.CertificateType][]Certificate{
			appstoreconnect.IOSDevelopment: {{
				CertificateInfo: cert,
				ID:              "dev1",
			}},
		},
		profiles: map[appstoreconnect.ProfileType][]Profile{
			appstoreconnect.IOSAppDevelopment: {},
		},
		appIDs: []appstoreconnect.BundleID{appID},
	})
	client.On("CheckBundleIDEntitlements", appID, mock.Anything).Return(NonmatchingProfileError{})

	return client
}
