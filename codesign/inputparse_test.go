package codesign

import (
	"reflect"
	"testing"

	"github.com/bitrise-io/go-xcode/v2/autocodesign/certdownloader"
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
