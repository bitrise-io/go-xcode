package codesignasset_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/autocodesign"
	"github.com/bitrise-io/go-xcode/v2/autocodesign/codesignasset"
	"github.com/bitrise-io/go-xcode/v2/autocodesign/devportalclient/appstoreconnect"
	"github.com/bitrise-io/go-xcode/v2/autocodesign/keychain"
	"github.com/bitrise-io/go-xcode/v2/mocks"
)

func TestWriter_InstallProfile(t *testing.T) {
	logger := log.NewLogger()
	keychain := keychain.Keychain{}
	homeDir, _ := os.UserHomeDir()
	legacyProfilePath := filepath.Join(homeDir, "Library", "MobileDevice", "Provisioning Profiles")
	modernProfilePath := filepath.Join(homeDir, "Library", "Developer", "Xcode", "UserData", "Provisioning Profiles")

	tests := []struct {
		name              string
		xcodeMajorVersion int64
		profilePlatform   appstoreconnect.BundleIDPlatform
		profileUUID       string
		wantProfileDir    string
		wantErr           bool
	}{
		{
			name:              "Xcode 15",
			xcodeMajorVersion: 15,
			profilePlatform:   appstoreconnect.IOS,
			profileUUID:       "test-uuid1",
			wantProfileDir:    legacyProfilePath,
		},
		{
			name:              "Xcode 26",
			xcodeMajorVersion: 26,
			profilePlatform:   appstoreconnect.IOS,
			profileUUID:       "test-uuid2",
			wantProfileDir:    modernProfilePath,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := &autocodesign.MockProfile{}
			profile.On("Attributes").Return(func() appstoreconnect.ProfileAttributes {
				return appstoreconnect.ProfileAttributes{
					Platform:       tt.profilePlatform,
					UUID:           tt.profileUUID,
					ProfileContent: []byte("test-content"),
				}
			})

			pathChecker := mocks.NewPathChecker(t)
			pathChecker.On("IsDirExists", tt.wantProfileDir).Return(true, nil).Once()

			fileManager := mocks.NewFileManager(t)
			expectedProfilePath := filepath.Join(tt.wantProfileDir, fmt.Sprintf("%s.mobileprovision", tt.profileUUID))
			fileManager.On("Write", expectedProfilePath, "test-content", os.FileMode(0600)).Return(nil).Once()

			w := codesignasset.NewWriter(logger, keychain, pathChecker, fileManager, tt.xcodeMajorVersion)
			gotErr := w.InstallProfile(profile)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("InstallProfile() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("InstallProfile() succeeded unexpectedly")
			}
		})
	}
}
