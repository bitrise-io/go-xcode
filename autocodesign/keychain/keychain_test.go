package keychain

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-xcode/certificateutil"
)

func TestCreateKeychain(t *testing.T) {
	dir, err := os.MkdirTemp("", "test-create-keychain")
	if err != nil {
		t.Errorf("setup: create temp dir for keychain: %s", err)
	}
	path := filepath.Join(dir, "testkeychain")
	_, err = createKeychain(path, "randompassword", command.NewFactory(env.NewRepository()))

	if err != nil {
		t.Errorf("error creating keychain: %s", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("keychain not created")
	}
}

func TestKeychain_importCertificate(t *testing.T) {
	const (
		// #nosec: G101  Potential hardcoded credentials (gosec)
		testPassphrase       = `!&$(){}?<>@ ;'\"/_=+-x\nGG}!Tk3/L'f-w){(}?om$DR&AM887)yowl` + "\t\n"
		testKeychainPassword = "password"
	)

	// Create test keychain
	dirTmp, err := os.MkdirTemp("", "test-import-certificate")
	if err != nil {
		t.Fatalf("setup: create temp dir for keychain: %s", err)
	}
	defer func() {
		if err := os.RemoveAll(dirTmp); err != nil {
			t.Fatalf("could not remove temp dir.")
		}
	}()

	keychainPath := filepath.Join(dirTmp, "testkeychain")
	_, err = createKeychain(keychainPath, testKeychainPassword, command.NewFactory(env.NewRepository()))
	if err != nil {
		t.Fatalf("error creating keychain: %s", err)
	}

	if _, err := os.Stat(keychainPath); os.IsNotExist(err) {
		t.Fatalf("keychain not created")
	}

	kchain := Keychain{path: keychainPath, password: testKeychainPassword, factory: command.NewFactory(env.NewRepository())}
	if err := kchain.unlock(); err != nil {
		t.Fatalf("failed to unlock keychain: %s", err)
	}

	// Create test p12 file
	const teamID = "MYTEAMID"
	const commonNameIOSDevelopment = "iPhone Developer: test"
	const teamName = "BITFALL FEJLESZTO KORLATOLT FELELOSSEGU TARSASAG"
	expiry := time.Now().AddDate(1, 0, 0)

	cert, privateKey, err := certificateutil.GenerateTestCertificate(int64(1), teamID, teamName, commonNameIOSDevelopment, expiry)
	if err != nil {
		t.Fatalf("init: failed to generate certificate: %s", err)
	}
	devCert := certificateutil.NewCertificateInfo(*cert, privateKey)
	t.Logf("Test certificate generated. %s", devCert)

	pfxData, err := devCert.EncodeToP12(testPassphrase)
	if err != nil {
		t.Fatalf("Setup: failed to encode test certificate to p12, error: %s", err)
	}

	testcertPath := filepath.Join(dirTmp, "TestCert.p12")
	if err := os.WriteFile(testcertPath, pfxData, 0600); err != nil {
		t.Fatalf("Setup: failed to write test p12 file.")
	}

	type fields struct {
		Path     string
		Password stepconf.Secret
	}
	type args struct {
		path       string
		passphrase stepconf.Secret
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Good password",
			fields: fields{
				Path:     keychainPath,
				Password: testKeychainPassword,
			},
			args: args{
				path:       testcertPath,
				passphrase: testPassphrase,
			},
			wantErr: false,
		},
		{
			name: "Incorrect password",
			fields: fields{
				Path:     keychainPath,
				Password: testKeychainPassword,
			},
			args: args{
				path:       testcertPath,
				passphrase: "Incorrect password",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := Keychain{
				path:     tt.fields.Path,
				password: tt.fields.Password,
				factory:  command.NewFactory(env.NewRepository()),
			}
			err := k.importCertificate(tt.args.path, tt.args.passphrase)
			if (err != nil) != tt.wantErr {
				t.Errorf("Keychain.importCertificate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				t.Logf("Keychain.importCertificate() error = %v", err)
			}
		})
	}
}
