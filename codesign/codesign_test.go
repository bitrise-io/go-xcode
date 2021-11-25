package codesign

import (
	"errors"
	"testing"
	"time"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-xcode/appleauth"
	"github.com/bitrise-io/go-xcode/autocodesign"
	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/bitrise-io/go-xcode/codesign/mocks"
	"github.com/bitrise-io/go-xcode/devportalservice"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_manager_selectCodeSigningStrategy(t *testing.T) {
	tests := []struct {
		name              string
		project           Project
		credentials       appleauth.Credentials
		XcodeMajorVersion int
		want              codeSigningStrategy
		wantErr           bool
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
			name: "API Key, Xcode 13, Xcode managed signing",
			credentials: appleauth.Credentials{
				APIKey: &devportalservice.APIKeyConnection{},
			},
			XcodeMajorVersion: 13,
			project:           newMockProject(true, nil),
			want:              codeSigningXcode,
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
				project: tt.project,
				opts: Opts{
					XcodeMajorVersion:          tt.XcodeMajorVersion,
					ShouldConsiderXcodeSigning: true,
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

func newMockProject(isAutoSign bool, mockErr error) Project {
	mockProjectHelper := new(mocks.Project)
	mockProjectHelper.On("IsSigningManagedAutomatically", mock.Anything).Return(isAutoSign, mockErr)

	return mockProjectHelper
}

func TestManager_downloadAndInstallCertificates(t *testing.T) {
	devCert := generateCert(t, "Apple Development: test")
	distCert := generateCert(t, "Apple Distribution: test")

	tests := []struct {
		name               string
		distributionMethod autocodesign.DistributionType
		certDownloader     autocodesign.CertificateProvider
		assetWriter        autocodesign.AssetWriter
		wantErr            bool
	}{
		{
			name:               "no certs uploaded, development",
			distributionMethod: autocodesign.Development,
			certDownloader:     newCertDownloaderMock([]certificateutil.CertificateInfoModel{}),
			wantErr:            true,
		},
		{
			name:               "development, no matching cert",
			distributionMethod: autocodesign.Development,
			certDownloader:     newCertDownloaderMock([]certificateutil.CertificateInfoModel{distCert}),
			wantErr:            true,
		},
		{
			name:               "no certs uploaded, distribution",
			distributionMethod: autocodesign.AppStore,
			certDownloader:     newCertDownloaderMock([]certificateutil.CertificateInfoModel{}),
		},
		{
			name:               "1 certs uploaded, development",
			distributionMethod: autocodesign.Development,
			certDownloader:     newCertDownloaderMock([]certificateutil.CertificateInfoModel{devCert}),
		},
		{
			name:               "1 certs uploaded, distribution",
			distributionMethod: autocodesign.AdHoc,
			certDownloader:     newCertDownloaderMock([]certificateutil.CertificateInfoModel{distCert}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manager{
				opts: Opts{
					ExportMethod: tt.distributionMethod,
				},
				certDownloader: tt.certDownloader,
				assetWriter:    newMockAssetWriter(nil),
				logger:         log.NewLogger(),
			}

			if err := m.downloadAndInstallCertificates(); (err != nil) != tt.wantErr {
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
