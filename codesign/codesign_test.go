package codesign

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/bitrise-io/go-xcode/exportoptions"
	"github.com/bitrise-io/go-xcode/profileutil"
	"github.com/bitrise-io/go-xcode/v2/autocodesign"
	"github.com/bitrise-io/go-xcode/v2/autocodesign/localcodesignasset"
	localcodesignassetMocks "github.com/bitrise-io/go-xcode/v2/autocodesign/localcodesignasset/mocks"
	"github.com/bitrise-io/go-xcode/v2/codesign/mocks"
	"github.com/bitrise-io/go-xcode/v2/devportalservice"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_manager_selectCodeSigningStrategy(t *testing.T) {
	tests := []struct {
		name                   string
		project                DetailsProvider
		credentials            devportalservice.Credentials
		XcodeMajorVersion      int
		minDaysProfileValidity int
		want                   codeSigningStrategy
		wantErr                bool
	}{
		{
			name: "Apple ID",
			credentials: devportalservice.Credentials{
				AppleID: &devportalservice.AppleID{},
			},
			project: newMockProject(false, nil),
			want:    codeSigningBitriseAppleID,
		},
		{
			name: "API Key, Xcode 12",
			credentials: devportalservice.Credentials{
				APIKey: &devportalservice.APIKeyConnection{},
			},
			XcodeMajorVersion: 12,
			project:           newMockProject(false, nil),
			want:              codeSigningBitriseAPIKey,
		},
		{
			name: "API Key, Xcode 13, Manual signing",
			credentials: devportalservice.Credentials{
				APIKey: &devportalservice.APIKeyConnection{},
			},
			XcodeMajorVersion: 13,
			project:           newMockProject(false, nil),
			want:              codeSigningBitriseAPIKey,
		},
		{
			name: "API Key, Xcode 13, Xcode managed signing, custom features",
			credentials: devportalservice.Credentials{
				APIKey: &devportalservice.APIKeyConnection{},
			},
			XcodeMajorVersion: 13,
			project:           newMockProject(true, nil),
			want:              codeSigningXcode,
		},
		{
			name: "API Key, Xcode 13, Xcode managed signing, no custom features",
			credentials: devportalservice.Credentials{
				APIKey: &devportalservice.APIKeyConnection{},
			},
			XcodeMajorVersion:      13,
			minDaysProfileValidity: 5,
			project:                newMockProject(true, nil),
			want:                   codeSigningBitriseAPIKey,
		},
		{
			name: "API Key, Xcode 13, can not determine if project automtic",
			credentials: devportalservice.Credentials{
				APIKey: &devportalservice.APIKeyConnection{},
			},
			XcodeMajorVersion: 13,
			project:           newMockProject(true, errors.New("")),
			want:              codeSigningBitriseAPIKey,
			wantErr:           true,
		},
		{
			name: "Enterprise API Key",
			credentials: devportalservice.Credentials{
				APIKey: &devportalservice.APIKeyConnection{
					EnterpriseAccount: true,
				},
			},
			XcodeMajorVersion: 16,
			project:           newMockProject(true, nil),
			want:              codeSigningBitriseAPIKey,
			wantErr:           false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manager{
				detailsProvider: tt.project,
				opts: Opts{
					XcodeMajorVersion:          tt.XcodeMajorVersion,
					ShouldConsiderXcodeSigning: true,
					MinDaysProfileValidity:     tt.minDaysProfileValidity,
				},
			}

			got, _, err := m.selectCodeSigningStrategy(tt.credentials)
			if (err != nil) != tt.wantErr {
				t.Errorf("manager.selectCodeSigningStrategy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func newMockProject(isAutoSign bool, mockErr error) DetailsProvider {
	mockProjectHelper := new(mocks.DetailsProvider)
	mockProjectHelper.On("IsSigningManagedAutomatically", mock.Anything).Return(isAutoSign, mockErr)

	return mockProjectHelper
}

func TestManager_checkXcodeManagedCertificates(t *testing.T) {
	devCert := generateCert(t, "Apple Development: test")
	distCert := generateCert(t, "Apple Distribution: test")

	tests := []struct {
		name               string
		distributionMethod autocodesign.DistributionType
		certificates       []certificateutil.CertificateInfoModel
		wantErr            bool
	}{
		{
			name:               "no certs uploaded, development",
			distributionMethod: autocodesign.Development,
			certificates:       []certificateutil.CertificateInfoModel{},
			wantErr:            true,
		},
		{
			name:               "development, no matching cert",
			distributionMethod: autocodesign.Development,
			certificates: []certificateutil.CertificateInfoModel{
				distCert,
			},
			wantErr: true,
		},
		{
			name:               "no certs uploaded, distribution",
			distributionMethod: autocodesign.AppStore,
			certificates:       []certificateutil.CertificateInfoModel{},
		},
		{
			name:               "1 certs uploaded, development",
			distributionMethod: autocodesign.Development,
			certificates: []certificateutil.CertificateInfoModel{
				devCert,
			},
		},
		{
			name:               "1 certs uploaded, distribution",
			distributionMethod: autocodesign.AdHoc,
			certificates: []certificateutil.CertificateInfoModel{
				distCert,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manager{
				opts: Opts{
					ExportMethod: tt.distributionMethod,
				},
				logger: log.NewLogger(),
			}

			if err := m.validateCertificatesForXcodeManagedSigning(tt.certificates); (err != nil) != tt.wantErr {
				t.Errorf("Manager.downloadAndInstallCertificates() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func generateCert(t *testing.T, commonName string) certificateutil.CertificateInfoModel {
	const (
		teamID   = "MYTEAMID"
		teamName = "BITFALL FEJLESZTO KORLATOLT FELELOSSEGU TARSASAG"
	)
	expiry := time.Now().AddDate(1, 0, 0)

	cert, privateKey, err := certificateutil.GenerateTestCertificate(int64(1), teamID, teamName, commonName, expiry)
	if err != nil {
		t.Fatalf("init: failed to generate certificate: %s", err)
	}

	return certificateutil.NewCertificateInfo(*cert, privateKey)
}

func TestSelectConnectionCredentials(t *testing.T) {
	testAPIKeyConnection := devportalservice.APIKeyConnection{
		KeyID:      "TestKeyID",
		IssuerID:   "TestIssuerID",
		PrivateKey: "test private key contents",
	}
	testAppleIDConnection := devportalservice.AppleIDConnection{
		AppleID:             "test@bitrise.io",
		Password:            "testpw",
		AppSpecificPassword: "testapppw",
		SessionExpiryDate:   nil,
		SessionCookies:      nil,
	}

	localKeyPath := filepath.Join(t.TempDir(), "key.p8")
	err := os.WriteFile(localKeyPath, []byte("private key contents"), 0700)
	if err != nil {
		t.Fatal(err.Error())
	}
	testInputs := ConnectionOverrideInputs{
		APIKeyPath:     stepconf.Secret(localKeyPath),
		APIKeyID:       "TestKeyIDFromInput",
		APIKeyIssuerID: "TestKeyIssuerIDFromInput",
	}
	testNoInputs := ConnectionOverrideInputs{}

	tests := []struct {
		name              string
		authType          AuthType
		bitriseConnection *devportalservice.AppleDeveloperConnection
		inputs            ConnectionOverrideInputs
		want              devportalservice.Credentials
		wantErr           bool
	}{
		{
			name:              "API key auth with nil Bitrise connection",
			authType:          APIKeyAuth,
			bitriseConnection: nil,
			inputs:            testNoInputs,
			wantErr:           true,
		},
		{
			name:     "API key auth type with valid Bitrise connection",
			authType: APIKeyAuth,
			bitriseConnection: &devportalservice.AppleDeveloperConnection{
				AppleIDConnection:     nil,
				APIKeyConnection:      &testAPIKeyConnection,
				TestDevices:           []devportalservice.TestDevice{},
				DuplicatedTestDevices: []devportalservice.TestDevice{},
			},
			inputs: testNoInputs,
			want: devportalservice.Credentials{
				AppleID: nil,
				APIKey:  &testAPIKeyConnection,
			},
		},
		{
			name:     "API key auth type without valid Bitrise connection",
			authType: APIKeyAuth,
			bitriseConnection: &devportalservice.AppleDeveloperConnection{
				AppleIDConnection:     nil,
				APIKeyConnection:      nil,
				TestDevices:           nil,
				DuplicatedTestDevices: nil,
			},
			inputs:  testNoInputs,
			wantErr: true,
		},
		{
			name:     "API key auth type without valid Bitrise connection but input overrides",
			authType: APIKeyAuth,
			bitriseConnection: &devportalservice.AppleDeveloperConnection{
				AppleIDConnection:     nil,
				APIKeyConnection:      nil,
				TestDevices:           nil,
				DuplicatedTestDevices: nil,
			},
			inputs: testInputs,
			want: devportalservice.Credentials{
				AppleID: nil,
				APIKey: &devportalservice.APIKeyConnection{
					KeyID:      "TestKeyIDFromInput",
					IssuerID:   "TestKeyIssuerIDFromInput",
					PrivateKey: "private key contents",
				},
			},
		},
		{
			name:     "API key auth type with valid Bitrise connection and input overrides",
			authType: APIKeyAuth,
			bitriseConnection: &devportalservice.AppleDeveloperConnection{
				AppleIDConnection:     nil,
				APIKeyConnection:      &testAPIKeyConnection,
				TestDevices:           []devportalservice.TestDevice{},
				DuplicatedTestDevices: []devportalservice.TestDevice{},
			},
			inputs: testInputs,
			want: devportalservice.Credentials{
				AppleID: nil,
				APIKey: &devportalservice.APIKeyConnection{
					KeyID:      "TestKeyIDFromInput",
					IssuerID:   "TestKeyIssuerIDFromInput",
					PrivateKey: "private key contents",
				},
			},
		},
		{
			name:              "Apple ID auth type with nil Bitrise connection",
			authType:          AppleIDAuth,
			bitriseConnection: nil,
			inputs:            testNoInputs,
			wantErr:           true,
		},
		{
			name:     "Apple ID auth type without valid Bitrise connection and input overrides for API key params",
			authType: AppleIDAuth,
			bitriseConnection: &devportalservice.AppleDeveloperConnection{
				AppleIDConnection:     nil,
				APIKeyConnection:      nil,
				TestDevices:           nil,
				DuplicatedTestDevices: nil,
			},
			inputs:  testInputs,
			wantErr: true,
		},
		{
			name:     "Apple ID auth type with valid Bitrise connection and input overrides for API key params",
			authType: AppleIDAuth,
			bitriseConnection: &devportalservice.AppleDeveloperConnection{
				AppleIDConnection:     &testAppleIDConnection,
				APIKeyConnection:      nil,
				TestDevices:           nil,
				DuplicatedTestDevices: nil,
			},
			inputs: testInputs,
			want: devportalservice.Credentials{
				AppleID: &devportalservice.AppleID{
					Username:            "test@bitrise.io",
					Password:            "testpw",
					Session:             "",
					AppSpecificPassword: "testapppw",
				},
				APIKey: nil,
			},
		},
		{
			name:     "Apple ID auth type with valid Bitrise connection",
			authType: AppleIDAuth,
			bitriseConnection: &devportalservice.AppleDeveloperConnection{
				AppleIDConnection:     &testAppleIDConnection,
				APIKeyConnection:      &testAPIKeyConnection,
				TestDevices:           nil,
				DuplicatedTestDevices: nil,
			},
			inputs: testNoInputs,
			want: devportalservice.Credentials{
				AppleID: &devportalservice.AppleID{
					Username:            "test@bitrise.io",
					Password:            "testpw",
					Session:             "",
					AppSpecificPassword: "testapppw",
				},
				APIKey: nil,
			},
		},
		{
			name:     "Apple ID auth type without valid Bitrise connection",
			authType: AppleIDAuth,
			bitriseConnection: &devportalservice.AppleDeveloperConnection{
				AppleIDConnection:     nil,
				APIKeyConnection:      nil,
				TestDevices:           nil,
				DuplicatedTestDevices: nil,
			},
			inputs:  testNoInputs,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SelectConnectionCredentials(tt.authType, tt.bitriseConnection, tt.inputs, log.NewLogger())
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectConnectionCredentials() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SelectConnectionCredentials() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManager_createCodeSignAssetMap(t *testing.T) {
	bundleID := "com.example.app"
	teamID := "team_1"
	certificate := certificateutil.CertificateInfoModel{
		Serial: "serial_1",
		TeamID: teamID,
	}
	profile := profileutil.ProvisioningProfileInfoModel{
		DeveloperCertificates: []certificateutil.CertificateInfoModel{certificate},
		BundleID:              bundleID,
		TeamID:                teamID,
		ExportType:            exportoptions.MethodDevelopment,
	}
	localProfile := localcodesignasset.NewProfile(profile, nil)

	tests := []struct {
		name         string
		appLayout    autocodesign.AppLayout
		certificates []certificateutil.CertificateInfoModel
		profiles     []profileutil.ProvisioningProfileInfoModel

		opts Opts

		want    map[autocodesign.DistributionType]autocodesign.AppCodesignAssets
		wantErr string
	}{
		{
			name: "Creates codesign assets map with single development profile",
			appLayout: autocodesign.AppLayout{
				EntitlementsByArchivableTargetBundleID: map[string]autocodesign.Entitlements{
					bundleID: nil,
				},
			},
			certificates: []certificateutil.CertificateInfoModel{certificate},
			profiles:     []profileutil.ProvisioningProfileInfoModel{profile},
			opts: Opts{
				ExportMethod: autocodesign.Development,
				TeamID:       teamID,
			},
			want: map[autocodesign.DistributionType]autocodesign.AppCodesignAssets{
				autocodesign.Development: {
					ArchivableTargetProfilesByBundleID: map[string]autocodesign.Profile{
						bundleID: localProfile,
					},
					UITestTargetProfilesByBundleID: map[string]autocodesign.Profile(nil),
					Certificate:                    certificate,
				},
			},
		},
		{
			name: "Throws an error when no code signing assets are found for the provided Developer Team",
			appLayout: autocodesign.AppLayout{
				EntitlementsByArchivableTargetBundleID: map[string]autocodesign.Entitlements{
					bundleID: {},
				},
			},
			certificates: []certificateutil.CertificateInfoModel{certificate},
			profiles:     []profileutil.ProvisioningProfileInfoModel{profile},
			opts: Opts{
				ExportMethod: autocodesign.Development,
				TeamID:       "team_2",
			},
			wantErr: "failed to determine codesign group for development distribution: no signing assets found",
		},
		{
			name: "Project entitlements are not filtering the profiles",
			appLayout: autocodesign.AppLayout{
				EntitlementsByArchivableTargetBundleID: map[string]autocodesign.Entitlements{
					bundleID: {"key1": "value1"},
				},
			},
			certificates: []certificateutil.CertificateInfoModel{certificate},
			profiles:     []profileutil.ProvisioningProfileInfoModel{profile},
			opts: Opts{
				ExportMethod: autocodesign.Development,
				TeamID:       teamID,
			},
			want: map[autocodesign.DistributionType]autocodesign.AppCodesignAssets{
				autocodesign.Development: {
					ArchivableTargetProfilesByBundleID: map[string]autocodesign.Profile{
						bundleID: localProfile,
					},
					UITestTargetProfilesByBundleID: map[string]autocodesign.Profile(nil),
					Certificate:                    certificate,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.NewLogger()
			profileConverter := new(localcodesignassetMocks.ProvisioningProfileConverter)
			profileConverter.On("ProfileInfoToProfile", profile).Return(localProfile, nil)

			m := &Manager{
				opts:             tt.opts,
				profileConverter: profileConverter,
				logger:           logger,
			}

			got, err := m.createCodeSignAssetMap(tt.appLayout, tt.certificates, tt.profiles)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}
