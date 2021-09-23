package autocodesign

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-xcode/autocodesign/devportalclient/appstoreconnect"
	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/bitrise-io/go-xcode/devportalservice"
)

type MockCertificateSource struct {
	certs map[appstoreconnect.CertificateType][]Certificate
}

func (m *MockCertificateSource) QueryCertificateBySerial(serial *big.Int) (Certificate, error) {
	for _, certList := range m.certs {
		for _, cert := range certList {
			if serial.Cmp(cert.Certificate.Certificate.SerialNumber) == 0 {
				return cert, nil
			}
		}
	}

	return Certificate{}, fmt.Errorf("certificate with serial %s not found", serial.String())
}

func (m *MockCertificateSource) QueryAllIOSCertificates() (map[appstoreconnect.CertificateType][]Certificate, error) {
	return m.certs, nil
}

func (m *MockCertificateSource) ListDevices(udid string, platform appstoreconnect.DevicePlatform) ([]appstoreconnect.Device, error) {
	return nil, nil
}
func (m *MockCertificateSource) RegisterDevice(testDevice devportalservice.TestDevice) (*appstoreconnect.Device, error) {
	return nil, nil
}

func (m *MockCertificateSource) FindProfile(name string, profileType appstoreconnect.ProfileType) (Profile, error) {
	return nil, nil
}
func (m *MockCertificateSource) DeleteProfile(id string) error {
	return nil
}
func (m *MockCertificateSource) CreateProfile(name string, profileType appstoreconnect.ProfileType, bundleID appstoreconnect.BundleID, certificateIDs []string, deviceIDs []string) (Profile, error) {
	return nil, nil
}

func (m *MockCertificateSource) FindBundleID(bundleIDIdentifier string) (*appstoreconnect.BundleID, error) {
	return nil, nil
}
func (m *MockCertificateSource) CheckBundleIDEntitlements(bundleID appstoreconnect.BundleID, projectEntitlements Entitlement) error {
	return nil
}
func (m *MockCertificateSource) SyncBundleID(bundleID appstoreconnect.BundleID, entitlements Entitlement) error {
	return nil
}
func (m *MockCertificateSource) CreateBundleID(bundleIDIdentifier string) (*appstoreconnect.BundleID, error) {
	return nil, nil
}

func NewMockCertificateSource(certs map[appstoreconnect.CertificateType][]Certificate) DevPortalClient {
	return &MockCertificateSource{
		certs: certs,
	}
}

func TestGetValidCertificates(t *testing.T) {
	log.SetEnableDebugLog(true)

	const teamID = "MYTEAMID"
	// Could be "Apple Development: test"
	const commonNameIOSDevelopment = "iPhone Developer: test"
	// Could be "Apple Distribution: test"
	const commonNameIOSDistribution = "iPhone Distribution: test"
	const teamName = "BITFALL FEJLESZTO KORLATOLT FELELOSSEGU TARSASAG"
	expiry := time.Now().AddDate(1, 0, 0)

	cert, privateKey, err := certificateutil.GenerateTestCertificate(int64(1), teamID, teamName, commonNameIOSDevelopment, expiry)
	if err != nil {
		t.Fatalf("init: failed to generate certificate: %s", err)
	}
	devCert := certificateutil.NewCertificateInfo(*cert, privateKey)
	t.Logf("Test certificate generated. %s", devCert)

	cert, privateKey, err = certificateutil.GenerateTestCertificate(int64(2), teamID, teamName, "iPhone Developer: test2", expiry)
	if err != nil {
		t.Errorf("init: failed to generate certificate: %s", err)
	}
	devCert2 := certificateutil.NewCertificateInfo(*cert, privateKey)
	t.Logf("Test certificate generated. %s", devCert)

	distCert, privateKey, err := certificateutil.GenerateTestCertificate(int64(10), teamID, teamName, commonNameIOSDistribution, expiry)
	if err != nil {
		t.Errorf("init: failed to generate certificate: %s", err)
	}
	distributionCert := certificateutil.NewCertificateInfo(*distCert, privateKey)
	t.Logf("Test certificate generated. %s", distributionCert)

	type args struct {
		localCertificates        []certificateutil.CertificateInfoModel
		client                   DevPortalClient
		requiredCertificateTypes map[appstoreconnect.CertificateType]bool
		teamID                   string
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
				localCertificates: []certificateutil.CertificateInfoModel{
					devCert,
				},
				client:                   NewMockCertificateSource(map[appstoreconnect.CertificateType][]Certificate{}),
				requiredCertificateTypes: map[appstoreconnect.CertificateType]bool{appstoreconnect.IOSDevelopment: true, appstoreconnect.IOSDistribution: false},
				teamID:                   "",
			},
			want:    map[appstoreconnect.CertificateType][]Certificate{},
			wantErr: true,
		},
		{
			name: "2 dev local with same name; 1 dev API; dev required",
			args: args{
				localCertificates: []certificateutil.CertificateInfoModel{
					devCert,
					devCert,
					devCert2,
				},
				client: NewMockCertificateSource(map[appstoreconnect.CertificateType][]Certificate{
					appstoreconnect.IOSDevelopment: {{
						Certificate: devCert,
						ID:          "devcert",
					}},
				}),
				requiredCertificateTypes: map[appstoreconnect.CertificateType]bool{appstoreconnect.IOSDevelopment: true, appstoreconnect.IOSDistribution: false},
				teamID:                   "",
			},
			want: map[appstoreconnect.CertificateType][]Certificate{
				appstoreconnect.IOSDevelopment: {{
					Certificate: devCert,
					ID:          "devcert",
				}},
			},
			wantErr: false,
		},
		{
			name: "no local; no API; dev+dist required",
			args: args{
				localCertificates:        []certificateutil.CertificateInfoModel{},
				client:                   NewMockCertificateSource(map[appstoreconnect.CertificateType][]Certificate{}),
				requiredCertificateTypes: map[appstoreconnect.CertificateType]bool{appstoreconnect.IOSDevelopment: true, appstoreconnect.IOSDistribution: true},
				teamID:                   "",
			},
			want:    map[appstoreconnect.CertificateType][]Certificate{},
			wantErr: true,
		},
		{
			name: "dev local; none API; dev+dist required",
			args: args{
				localCertificates: []certificateutil.CertificateInfoModel{
					devCert,
				},
				client:                   NewMockCertificateSource(map[appstoreconnect.CertificateType][]Certificate{}),
				requiredCertificateTypes: map[appstoreconnect.CertificateType]bool{appstoreconnect.IOSDevelopment: true, appstoreconnect.IOSDistribution: true},
				teamID:                   "",
			},
			want:    map[appstoreconnect.CertificateType][]Certificate{},
			wantErr: true,
		},
		{
			name: "dev local; dev API; dev required",
			args: args{
				localCertificates: []certificateutil.CertificateInfoModel{
					devCert,
				},
				client: NewMockCertificateSource(map[appstoreconnect.CertificateType][]Certificate{
					appstoreconnect.IOSDevelopment: {{
						Certificate: devCert,
						ID:          "apicertid",
					}},
				}),
				requiredCertificateTypes: map[appstoreconnect.CertificateType]bool{appstoreconnect.IOSDevelopment: true, appstoreconnect.IOSDistribution: false},
				teamID:                   "",
			},
			want: map[appstoreconnect.CertificateType][]Certificate{
				appstoreconnect.IOSDevelopment: {{
					Certificate: devCert,
					ID:          "apicertid",
				}},
			},
			wantErr: false,
		},
		{
			name: "2 dev local; 1 dev API; dev required",
			args: args{
				localCertificates: []certificateutil.CertificateInfoModel{
					devCert,
					devCert2,
				},
				client: NewMockCertificateSource(map[appstoreconnect.CertificateType][]Certificate{
					appstoreconnect.IOSDevelopment: {{
						Certificate: devCert,
						ID:          "dev1",
					}},
				}),
				requiredCertificateTypes: map[appstoreconnect.CertificateType]bool{appstoreconnect.IOSDevelopment: true, appstoreconnect.IOSDistribution: false},
				teamID:                   "",
			},
			want: map[appstoreconnect.CertificateType][]Certificate{
				appstoreconnect.IOSDevelopment: {{
					Certificate: devCert,
					ID:          "dev1",
				}},
			},
			wantErr: false,
		},
		{
			name: "dev local; dev+dist API; both required",
			args: args{
				localCertificates: []certificateutil.CertificateInfoModel{
					devCert,
				},
				client: NewMockCertificateSource(map[appstoreconnect.CertificateType][]Certificate{
					appstoreconnect.IOSDevelopment: {
						{
							Certificate: devCert,
							ID:          "apicertid_dev",
						},
						{
							Certificate: distributionCert,
							ID:          "apicertid_dist",
						},
					},
				}),
				requiredCertificateTypes: map[appstoreconnect.CertificateType]bool{appstoreconnect.IOSDevelopment: true, appstoreconnect.IOSDistribution: true},
				teamID:                   "",
			},
			want:    map[appstoreconnect.CertificateType][]Certificate{},
			wantErr: true,
		},
		{
			name: "dev+dist local; dist API; dev+dist required",
			args: args{
				localCertificates: []certificateutil.CertificateInfoModel{
					devCert,
					distributionCert,
				},
				client: NewMockCertificateSource(map[appstoreconnect.CertificateType][]Certificate{
					appstoreconnect.IOSDevelopment: {{
						Certificate: devCert,
						ID:          "dev",
					}},
				}),
				requiredCertificateTypes: map[appstoreconnect.CertificateType]bool{
					appstoreconnect.IOSDevelopment:  true,
					appstoreconnect.IOSDistribution: true,
				},
				teamID: "",
			},
			want:    map[appstoreconnect.CertificateType][]Certificate{},
			wantErr: true,
		},
		{
			name: "dev+dist local; dev+dist API; dev+dist required",
			args: args{
				localCertificates: []certificateutil.CertificateInfoModel{
					devCert,
					distributionCert,
				},
				client: NewMockCertificateSource(map[appstoreconnect.CertificateType][]Certificate{
					appstoreconnect.IOSDevelopment: {
						{
							Certificate: devCert,
							ID:          "dev",
						},
					},
					appstoreconnect.IOSDistribution: {
						{
							Certificate: distributionCert,
							ID:          "dist",
						},
					},
				}),
				requiredCertificateTypes: map[appstoreconnect.CertificateType]bool{appstoreconnect.IOSDevelopment: true, appstoreconnect.IOSDistribution: true},
				teamID:                   "",
			},
			want: map[appstoreconnect.CertificateType][]Certificate{
				appstoreconnect.IOSDevelopment: {{
					Certificate: devCert,
					ID:          "dev",
				}},
				appstoreconnect.IOSDistribution: {{
					Certificate: distributionCert,
					ID:          "dist",
				}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getValidCertificates(tt.args.localCertificates, tt.args.client, tt.args.requiredCertificateTypes, tt.args.teamID, true)
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
