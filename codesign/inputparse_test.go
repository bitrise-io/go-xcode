package codesign

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-xcode/devportalservice"

	"github.com/bitrise-io/go-xcode/v2/autocodesign/certdownloader"
	"github.com/stretchr/testify/require"
)

func TestParseCertificates(t *testing.T) {
	tests := []struct {
		name    string
		input   Input
		want    []certdownloader.CertificateAndPassphrase
		wantErr bool
	}{
		{
			name: "One certificate and passphrase",
			input: Input{
				CertificateURLList:        "https://example.com/storage/development.p12 ",
				CertificatePassphraseList: "password",
				KeychainPath:              t.TempDir(),
				KeychainPassword:          "keychainpassword",
			},
			want: []certdownloader.CertificateAndPassphrase{{
				URL:        "https://example.com/storage/development.p12",
				Passphrase: "password",
			}},
		},
		{
			name: "Multiple certificates and passphrases",
			input: Input{
				CertificateURLList:        "https://example.com/storage/development.p12|https://example.com/storage/distribution.p12",
				CertificatePassphraseList: "password|password2",
				KeychainPath:              t.TempDir(),
				KeychainPassword:          "keychainpassword",
			},
			want: []certdownloader.CertificateAndPassphrase{
				{
					URL:        "https://example.com/storage/development.p12",
					Passphrase: "password",
				},
				{
					URL:        "https://example.com/storage/distribution.p12",
					Passphrase: "password2",
				},
			},
		},
		{
			name: "Empty certificate and passphrase value in lists",
			input: Input{
				CertificateURLList:        "https://example.com/storage/development.p12|",
				CertificatePassphraseList: "password|",
				KeychainPath:              t.TempDir(),
				KeychainPassword:          "keychainpassword",
			},
			wantErr: true, // Because empty passphrase is a valid value for a no-passphrase-cert, so the list has 2 items
		},
		{
			name: "No certificate nor passphrase",
			input: Input{
				CertificateURLList:        "",
				CertificatePassphraseList: "",
				KeychainPath:              t.TempDir(),
				KeychainPassword:          "keychainpassword",
			},
			wantErr: true,
		},
		{
			name: "Mismatch in certificate and passphrase count",
			input: Input{
				CertificateURLList:        "https://example.com/storage/development.p12|https://example.com/storage/distribution.p12",
				CertificatePassphraseList: "password",
				KeychainPath:              t.TempDir(),
				KeychainPassword:          "keychainpassword",
			},
			wantErr: true,
		},
		{
			name: "One certificate without passphrase",
			input: Input{
				CertificateURLList:        "https://example.com/storage/development.p12",
				CertificatePassphraseList: "",
				KeychainPath:              t.TempDir(),
				KeychainPassword:          "keychainpassword",
			},
			want: []certdownloader.CertificateAndPassphrase{{
				URL:        "https://example.com/storage/development.p12",
				Passphrase: "",
			}},
		},
		{
			name: "Multiple certificates without passphrases",
			input: Input{
				CertificateURLList:        "https://example.com/storage/development.p12|https://example.com/storage/distribution.p12|https://example.com/storage/adhoc.p12",
				CertificatePassphraseList: "||",
				KeychainPath:              t.TempDir(),
				KeychainPassword:          "keychainpassword",
			},
			want: []certdownloader.CertificateAndPassphrase{
				{
					URL:        "https://example.com/storage/development.p12",
					Passphrase: "",
				},
				{
					URL:        "https://example.com/storage/distribution.p12",
					Passphrase: "",
				},
				{
					URL:        "https://example.com/storage/adhoc.p12",
					Passphrase: "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCertificates(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateAndExpandProfilePaths(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "file.mobileprovision"), []byte{}, 0600)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(dir, "file2.mobileprovision"), []byte{}, 0600)
	require.NoError(t, err)

	tests := []struct {
		name         string
		profilesList string
		want         []string
		wantErr      bool
	}{
		{
			name:         "Single profile",
			profilesList: "https://file",
			want:         []string{"https://file"},
		},
		{
			name:         "Multiple profiles pipe separated",
			profilesList: "file://file1| https://file2 ",
			want:         []string{"file://file1", "https://file2"},
		},
		{
			name:         "Multiple profiles newline separated",
			profilesList: "file://file1\nfile://file2\n",
			want:         []string{"file://file1", "file://file2"},
		},
		{
			name:         "Multiple profiles newline at the end",
			profilesList: "https://file1|https://file2|https://file3\n",
			want:         []string{"https://file1", "https://file2", "https://file3"},
		},
		{
			name:         "Multiple profiles mixed (not supported)",
			profilesList: "file://file1\nfile://file2 | file://file3",
			want:         []string{},
			wantErr:      true,
		},
		{
			name:         "Directory",
			profilesList: dir,
			want: []string{
				fmt.Sprintf("file://%s", filepath.Join("", dir, "file.mobileprovision")),
				fmt.Sprintf("file://%s", filepath.Join("", dir, "file2.mobileprovision")),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateAndExpandProfilePaths(tt.profilesList)

			if !tt.wantErr {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_ParseConnectionOverrideConfig(t *testing.T) {
	// Given
	path := filepath.Join(t.TempDir(), "private_key.p8")
	fileContent := "this is a private key"
	err := ioutil.WriteFile(path, []byte(fileContent), 0666)
	if err != nil {
		t.Errorf(err.Error())
	}

	keyID := " ABC123   "
	keyIssuerID := "   ABC456 "

	// When
	connection, err := ParseConnectionOverrideConfig(stepconf.Secret(path), keyID, keyIssuerID)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Then
	expected := devportalservice.APIKeyConnection{
		KeyID:      "ABC123",
		IssuerID:   "ABC456",
		PrivateKey: fileContent,
	}
	require.Equal(t, expected, *connection)
}
