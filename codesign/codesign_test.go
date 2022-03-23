package codesign

import (
	"errors"
	"testing"
	"time"

	"github.com/bitrise-io/go-xcode/v2/autocodesign/devportalclient/appstoreconnect"
	"github.com/bitrise-io/go-xcode/v2/codesign/mocks"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/appleauth"
	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/bitrise-io/go-xcode/devportalservice"
	"github.com/bitrise-io/go-xcode/v2/autocodesign"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_manager_selectCodeSigningStrategy(t *testing.T) {
	tests := []struct {
		name                   string
		project                DetailsProvider
		credentials            appleauth.Credentials
		XcodeMajorVersion      int
		minDaysProfileValidity int
		want                   codeSigningStrategy
		wantErr                bool
	}{
		{
			name: "Apple ID",
			credentials: appleauth.Credentials{
				AppleID: &appleauth.AppleID{},
			},
			project: newMockProject(false, nil),
			want:    codeSigningBitriseAppleID,
		},
		{
			name: "API Key, Xcode 12",
			credentials: appleauth.Credentials{
				APIKey: &devportalservice.APIKeyConnection{},
			},
			XcodeMajorVersion: 12,
			project:           newMockProject(false, nil),
			want:              codeSigningBitriseAPIKey,
		},
		{
			name: "API Key, Xcode 13, Manual signing",
			credentials: appleauth.Credentials{
				APIKey: &devportalservice.APIKeyConnection{},
			},
			XcodeMajorVersion: 13,
			project:           newMockProject(false, nil),
			want:              codeSigningBitriseAPIKey,
		},
		{
			name: "API Key, Xcode 13, Xcode managed signing, custom features",
			credentials: appleauth.Credentials{
				APIKey: &devportalservice.APIKeyConnection{},
			},
			XcodeMajorVersion: 13,
			project:           newMockProject(true, nil),
			want:              codeSigningXcode,
		},
		{
			name: "API Key, Xcode 13, Xcode managed signing, no custom features",
			credentials: appleauth.Credentials{
				APIKey: &devportalservice.APIKeyConnection{},
			},
			XcodeMajorVersion:      13,
			minDaysProfileValidity: 5,
			project:                newMockProject(true, nil),
			want:                   codeSigningBitriseAPIKey,
		},
		{
			name: "API Key, Xcode 13, can not determine if project automtic",
			credentials: appleauth.Credentials{
				APIKey: &devportalservice.APIKeyConnection{},
			},
			XcodeMajorVersion: 13,
			project:           newMockProject(true, errors.New("")),
			want:              codeSigningBitriseAPIKey,
			wantErr:           true,
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
		typeToLocalCerts   localCertificates
		wantErr            bool
	}{
		{
			name:               "no certs uploaded, development",
			distributionMethod: autocodesign.Development,
			typeToLocalCerts:   localCertificates{},
			wantErr:            true,
		},
		{
			name:               "development, no matching cert",
			distributionMethod: autocodesign.Development,
			typeToLocalCerts: localCertificates{
				appstoreconnect.IOSDistribution: {distCert},
			},
			wantErr: true,
		},
		{
			name:               "no certs uploaded, distribution",
			distributionMethod: autocodesign.AppStore,
			typeToLocalCerts:   localCertificates{},
		},
		{
			name:               "1 certs uploaded, development",
			distributionMethod: autocodesign.Development,
			typeToLocalCerts: localCertificates{
				appstoreconnect.IOSDevelopment: {devCert},
			},
		},
		{
			name:               "1 certs uploaded, distribution",
			distributionMethod: autocodesign.AdHoc,
			typeToLocalCerts: localCertificates{
				appstoreconnect.IOSDistribution: {distCert},
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

			if err := m.checkXcodeManagedCertificates(tt.typeToLocalCerts); (err != nil) != tt.wantErr {
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

func newCertDownloaderMock(certs []certificateutil.CertificateInfoModel) autocodesign.CertificateProvider {
	mockDownloader := new(autocodesign.MockCertificateProvider)
	mockDownloader.On("GetCertificates").Return(certs, nil)

	return mockDownloader
}

func newMockAssetWriter(mockErr error) autocodesign.AssetWriter {
	mockWriter := new(autocodesign.MockAssetWriter)
	mockWriter.On("InstallCertificate", mock.Anything).Return(mockErr)

	return mockWriter
}
