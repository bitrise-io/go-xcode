package autocodesign

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-xcode/v2/autocodesign/devportalclient/appstoreconnect"
	"github.com/bitrise-io/go-xcode/v2/certificateutil"
	"github.com/bitrise-io/go-xcode/v2/timeutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type devportalArgs struct {
	certs    map[appstoreconnect.CertificateType][]Certificate
	devices  []appstoreconnect.Device
	profiles map[appstoreconnect.ProfileType][]Profile
	appIDs   []appstoreconnect.BundleID
}

// newMockDevportalClient is a default mock implementing listing of trivial assets
// To be mocked in tests:
// - RegisterDevice
// - DeleteProfile
// - CreateProfile
// - CheckBundleIDEntitlements
// - SyncBundleID
// - CreateBundleID
func newMockDevportalClient(m devportalArgs) *MockDevPortalClient {
	mockDevportalClient := new(MockDevPortalClient)
	mockDevportalClient.On("QueryCertificateBySerial", mock.Anything).Return(
		func(serial big.Int) Certificate {
			for _, certList := range m.certs {
				for _, cert := range certList {
					if serial.Cmp(cert.CertificateInfo.Certificate.SerialNumber) == 0 {
						return cert
					}
				}
			}

			return Certificate{}
		},
		func(serial big.Int) error {
			for _, certList := range m.certs {
				for _, cert := range certList {
					if serial.Cmp(cert.CertificateInfo.Certificate.SerialNumber) == 0 {
						return nil
					}
				}
			}

			return fmt.Errorf("certificate with serial %s not found", serial.String())
		},
	)
	mockDevportalClient.On("QueryAllIOSCertificates").Return(func() map[appstoreconnect.CertificateType][]Certificate {
		return m.certs
	}, nil)
	mockDevportalClient.On("ListDevices", "", appstoreconnect.IOSDevice).Return(func(udid string, platform appstoreconnect.DevicePlatform) []appstoreconnect.Device {
		return m.devices
	}, nil)
	mockDevportalClient.On("FindProfile", mock.Anything, mock.Anything).Return(func(name string, profileType appstoreconnect.ProfileType) Profile {
		profiles, ok := m.profiles[profileType]
		if !ok {
			panic(fmt.Sprintf("invalid type: %T", profileType))
		}

		for _, profile := range profiles {
			if profile.Attributes().Name == name {
				return profile
			}
		}

		return nil
	}, nil)
	mockDevportalClient.On("FindBundleID", mock.Anything).Return(func(bundleIDIdentifier string) *appstoreconnect.BundleID {
		for _, appID := range m.appIDs {
			if appID.Attributes.Identifier == bundleIDIdentifier {
				return &appID
			}
		}

		return nil
	}, nil)

	return mockDevportalClient
}

func newMockCertClient(certs map[appstoreconnect.CertificateType][]Certificate) DevPortalClient {
	return newMockDevportalClient(devportalArgs{
		certs: certs,
	})
}

func Test_getValidCertificates(t *testing.T) {
	log.SetEnableDebugLog(true)

	const (
		teamID   = "MYTEAMID"
		teamName = "BITFALL FEJLESZTO KORLATOLT FELELOSSEGU TARSASAG"
	)
	notBefore := time.Now()
	expiry := notBefore.AddDate(1, 0, 0)

	cert, privateKey, err := certificateutil.GenerateTestCertificate(int64(1), teamID, teamName, "Apple Development: test", notBefore, expiry)
	require.NoError(t, err, "init: failed to generate certificate: %s", err)
	devCert := certificateutil.NewCertificateInfo(*cert, privateKey)
	t.Logf("Test certificate generated. %s", devCert)

	cert, privateKey, err = certificateutil.GenerateTestCertificate(int64(2), teamID, teamName, "iPhone Developer: test2", notBefore, expiry)
	require.NoError(t, err, "init: failed to generate certificate: %s", err)
	devCert2 := certificateutil.NewCertificateInfo(*cert, privateKey)
	t.Logf("Test certificate generated. %s", devCert2)

	distCert, privateKey, err := certificateutil.GenerateTestCertificate(int64(10), teamID, teamName, "Apple Distribution: test", notBefore, expiry)
	require.NoError(t, err, "init: failed to generate certificate: %s", err)
	distributionCert := certificateutil.NewCertificateInfo(*distCert, privateKey)
	t.Logf("Test certificate generated. %s", distributionCert)

	type args struct {
		typeToLocalCerts         LocalCertificates
		client                   DevPortalClient
		requiredCertificateTypes map[appstoreconnect.CertificateType]bool
	}
	tests := []struct {
		name    string
		args    args
		want    map[appstoreconnect.CertificateType][]Certificate
		wantErr bool
	}{
		{
			name: "dev local; no API; dev required",
			args: args{
				typeToLocalCerts: LocalCertificates{
					appstoreconnect.IOSDevelopment: {devCert},
				},
				client:                   newMockCertClient(map[appstoreconnect.CertificateType][]Certificate{}),
				requiredCertificateTypes: map[appstoreconnect.CertificateType]bool{appstoreconnect.IOSDevelopment: true, appstoreconnect.IOSDistribution: false},
			},
			want:    map[appstoreconnect.CertificateType][]Certificate{},
			wantErr: true,
		},
		{
			name: "2 dev local with same name; 1 dev API; dev required",
			args: args{
				typeToLocalCerts: LocalCertificates{
					appstoreconnect.IOSDevelopment: {
						devCert,
						devCert,
						devCert2,
					},
				},
				client: newMockCertClient(map[appstoreconnect.CertificateType][]Certificate{
					appstoreconnect.IOSDevelopment: {{
						CertificateInfo: devCert,
						ID:              "devcert",
					}},
				}),
				requiredCertificateTypes: map[appstoreconnect.CertificateType]bool{appstoreconnect.IOSDevelopment: true, appstoreconnect.IOSDistribution: false},
			},
			want: map[appstoreconnect.CertificateType][]Certificate{
				appstoreconnect.IOSDevelopment: {
					{
						CertificateInfo: devCert,
						ID:              "devcert",
					},
					{
						CertificateInfo: devCert,
						ID:              "devcert",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "no local; no API; dev+dist required",
			args: args{
				typeToLocalCerts:         LocalCertificates{},
				client:                   newMockCertClient(map[appstoreconnect.CertificateType][]Certificate{}),
				requiredCertificateTypes: map[appstoreconnect.CertificateType]bool{appstoreconnect.IOSDevelopment: true, appstoreconnect.IOSDistribution: true},
			},
			want:    map[appstoreconnect.CertificateType][]Certificate{},
			wantErr: true,
		},
		{
			name: "dev local; none API; dev+dist required",
			args: args{
				typeToLocalCerts: LocalCertificates{
					appstoreconnect.IOSDevelopment: {devCert},
				},
				client:                   newMockCertClient(map[appstoreconnect.CertificateType][]Certificate{}),
				requiredCertificateTypes: map[appstoreconnect.CertificateType]bool{appstoreconnect.IOSDevelopment: true, appstoreconnect.IOSDistribution: true},
			},
			want:    map[appstoreconnect.CertificateType][]Certificate{},
			wantErr: true,
		},
		{
			name: "dev local; dev API; dev required",
			args: args{
				typeToLocalCerts: LocalCertificates{
					appstoreconnect.IOSDevelopment: {devCert},
				},
				client: newMockCertClient(map[appstoreconnect.CertificateType][]Certificate{
					appstoreconnect.IOSDevelopment: {{
						CertificateInfo: devCert,
						ID:              "apicertid",
					}},
				}),
				requiredCertificateTypes: map[appstoreconnect.CertificateType]bool{appstoreconnect.IOSDevelopment: true, appstoreconnect.IOSDistribution: false},
			},
			want: map[appstoreconnect.CertificateType][]Certificate{
				appstoreconnect.IOSDevelopment: {{
					CertificateInfo: devCert,
					ID:              "apicertid",
				}},
			},
			wantErr: false,
		},
		{
			name: "2 dev local; 1 dev API; dev required",
			args: args{
				typeToLocalCerts: LocalCertificates{
					appstoreconnect.IOSDevelopment: {
						devCert,
						devCert2,
					},
				},
				client: newMockCertClient(map[appstoreconnect.CertificateType][]Certificate{
					appstoreconnect.IOSDevelopment: {{
						CertificateInfo: devCert,
						ID:              "dev1",
					}},
				}),
				requiredCertificateTypes: map[appstoreconnect.CertificateType]bool{appstoreconnect.IOSDevelopment: true, appstoreconnect.IOSDistribution: false},
			},
			want: map[appstoreconnect.CertificateType][]Certificate{
				appstoreconnect.IOSDevelopment: {{
					CertificateInfo: devCert,
					ID:              "dev1",
				}},
			},
			wantErr: false,
		},
		{
			name: "dev local; dev+dist API; both required",
			args: args{
				typeToLocalCerts: LocalCertificates{
					appstoreconnect.IOSDevelopment: {devCert},
				},
				client: newMockCertClient(map[appstoreconnect.CertificateType][]Certificate{
					appstoreconnect.IOSDevelopment: {
						{
							CertificateInfo: devCert,
							ID:              "apicertid_dev",
						},
						{
							CertificateInfo: distributionCert,
							ID:              "apicertid_dist",
						},
					},
				}),
				requiredCertificateTypes: map[appstoreconnect.CertificateType]bool{appstoreconnect.IOSDevelopment: true, appstoreconnect.IOSDistribution: true},
			},
			want:    map[appstoreconnect.CertificateType][]Certificate{},
			wantErr: true,
		},
		{
			name: "dev+dist local; dist API; dev+dist required",
			args: args{
				typeToLocalCerts: LocalCertificates{
					appstoreconnect.IOSDevelopment:  {devCert},
					appstoreconnect.IOSDistribution: {distributionCert},
				},
				client: newMockCertClient(map[appstoreconnect.CertificateType][]Certificate{
					appstoreconnect.IOSDevelopment: {{
						CertificateInfo: devCert,
						ID:              "dev",
					}},
				}),
				requiredCertificateTypes: map[appstoreconnect.CertificateType]bool{
					appstoreconnect.IOSDevelopment:  true,
					appstoreconnect.IOSDistribution: true,
				},
			},
			want:    map[appstoreconnect.CertificateType][]Certificate{},
			wantErr: true,
		},
		{
			name: "dev+dist local; dev+dist API; dev+dist required",
			args: args{
				typeToLocalCerts: LocalCertificates{
					appstoreconnect.IOSDevelopment:  {devCert},
					appstoreconnect.IOSDistribution: {distributionCert},
				},
				client: newMockCertClient(map[appstoreconnect.CertificateType][]Certificate{
					appstoreconnect.IOSDevelopment: {
						{
							CertificateInfo: devCert,
							ID:              "dev",
						},
					},
					appstoreconnect.IOSDistribution: {
						{
							CertificateInfo: distributionCert,
							ID:              "dist",
						},
					},
				}),
				requiredCertificateTypes: map[appstoreconnect.CertificateType]bool{appstoreconnect.IOSDevelopment: true, appstoreconnect.IOSDistribution: true},
			},
			want: map[appstoreconnect.CertificateType][]Certificate{
				appstoreconnect.IOSDevelopment: {{
					CertificateInfo: devCert,
					ID:              "dev",
				}},
				appstoreconnect.IOSDistribution: {{
					CertificateInfo: distributionCert,
					ID:              "dist",
				}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getValidCertificates(tt.args.typeToLocalCerts, tt.args.client, tt.args.requiredCertificateTypes, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetValidCertificates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for certType, wantCerts := range tt.want {
				if !reflect.DeepEqual(wantCerts, got[certType]) {
					t.Errorf("GetValidCertificates()[%s] = %v, want %v", certType, got, tt.want)
				}
			}
		})
	}
}

func TestGetValidLocalCertificates(t *testing.T) {
	log.SetEnableDebugLog(true)

	const (
		teamID   = "MYTEAMID"
		teamName = "BITFALL FEJLESZTO KORLATOLT FELELOSSEGU TARSASAG"
	)
	notBefore := time.Now()
	expiry := notBefore.AddDate(1, 0, 0)

	cert, privateKey, err := certificateutil.GenerateTestCertificate(int64(1), teamID, teamName, "Apple Development: test", notBefore, expiry)
	require.NoError(t, err, "init: failed to generate certificate: %s", err)
	devCert := certificateutil.NewCertificateInfo(*cert, privateKey)
	t.Logf("Test certificate generated. %s", devCert)

	cert, privateKey, err = certificateutil.GenerateTestCertificate(int64(2), teamID, teamName, "iPhone Developer: test2", notBefore, expiry)
	require.NoError(t, err, "init: failed to generate certificate: %s", err)
	devCert2 := certificateutil.NewCertificateInfo(*cert, privateKey)
	t.Logf("Test certificate generated. %s", devCert2)

	distCert, privateKey, err := certificateutil.GenerateTestCertificate(int64(10), teamID, teamName, "Apple Distribution: test", notBefore, expiry)
	require.NoError(t, err, "init: failed to generate certificate: %s", err)
	distributionCert := certificateutil.NewCertificateInfo(*distCert, privateKey)
	t.Logf("Test certificate generated. %s", distributionCert)

	tests := []struct {
		name         string
		certificates []certificateutil.CertificateInfo
		want         LocalCertificates
		wantErr      bool
	}{
		{
			name: "Duplicate certificate (same name)",
			certificates: []certificateutil.CertificateInfo{
				devCert,
				devCert,
				devCert2,
			},
			want: LocalCertificates{
				appstoreconnect.IOSDevelopment: {
					devCert,
					devCert2,
				},
				appstoreconnect.IOSDistribution: nil,
			},
		},
		{
			name: "dev + dist cert",
			certificates: []certificateutil.CertificateInfo{
				distributionCert,
				devCert,
			},
			want: LocalCertificates{
				appstoreconnect.IOSDevelopment: {
					devCert,
				},
				appstoreconnect.IOSDistribution: {
					distributionCert,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetValidLocalCertificates(tt.certificates, timeutil.NewDefaultTimeProvider())

			require.NoError(t, err)
			for _, certType := range []appstoreconnect.CertificateType{appstoreconnect.IOSDevelopment, appstoreconnect.IOSDistribution} {
				require.ElementsMatch(t, tt.want[certType], got[certType])
			}
		})
	}
}
