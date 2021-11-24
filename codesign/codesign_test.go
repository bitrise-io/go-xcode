package codesign

import (
	"errors"
	"testing"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-xcode/appleauth"
	"github.com/bitrise-io/go-xcode/autocodesign"
	"github.com/bitrise-io/go-xcode/autocodesign/codesignasset"
	"github.com/bitrise-io/go-xcode/autocodesign/keychain"
	autoMocks "github.com/bitrise-io/go-xcode/autocodesign/mocks"
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
	// const teamID = "MYTEAMID"
	// // Could be "Apple Development: test"
	// const commonNameIOSDevelopment = "iPhone Developer: test"
	// // Could be "Apple Distribution: test"
	// const commonNameIOSDistribution = "iPhone Distribution: test"
	// const teamName = "BITFALL FEJLESZTO KORLATOLT FELELOSSEGU TARSASAG"
	// expiry := time.Now().AddDate(1, 0, 0)

	// cert, privateKey, err := certificateutil.GenerateTestCertificate(int64(1), teamID, teamName, commonNameIOSDevelopment, expiry)
	// if err != nil {
	// 	t.Fatalf("init: failed to generate certificate: %s", err)
	// }
	// devCert := certificateutil.NewCertificateInfo(*cert, privateKey)

	tests := []struct {
		name               string
		distributionMethod autocodesign.DistributionType
		certDownloader     autocodesign.CertificateProvider
		keychain           keychain.Keychain
		assetWriter        codesignasset.Writer
		wantErr            bool
	}{
		{
			name:               "no certs uploaded, development",
			distributionMethod: autocodesign.Development,
			certDownloader:     newCertDownloaderMock([]certificateutil.CertificateInfoModel{}),
			wantErr:            true,
		},
		{
			name:               "no certs uploaded, distribution",
			distributionMethod: autocodesign.AppStore,
			certDownloader:     newCertDownloaderMock([]certificateutil.CertificateInfoModel{}),
		},
		// {
		// 	name:               "1 certs uploaded, development",
		// 	distributionMethod: autocodesign.Development,
		// 	certDownloader:     newCertDownloaderMock([]certificateutil.CertificateInfoModel{devCert}),
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manager{
				opts: Opts{
					ExportMethod: tt.distributionMethod,
				},
				certDownloader: tt.certDownloader,
				keychain:       tt.keychain,
				assetWriter:    tt.assetWriter,
				logger:         log.NewLogger(),
			}

			if err := m.downloadAndInstallCertificates(); (err != nil) != tt.wantErr {
				t.Errorf("Manager.downloadAndInstallCertificates() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func newCertDownloaderMock(certs []certificateutil.CertificateInfoModel) autocodesign.CertificateProvider {
	mockDownloader := new(autoMocks.CertificateProvider)
	mockDownloader.On("GetCertificates").Return(certs, nil)

	return mockDownloader
}
