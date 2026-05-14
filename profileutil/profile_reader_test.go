package profileutil

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/mocks"
	"github.com/fullsailor/pkcs7"
	"github.com/stretchr/testify/require"
)

func TestProfileReader_ProvisioningProfileInfoFromFile(t *testing.T) {
	t.Run("file open error is propagated", func(t *testing.T) {
		fileManager := mocks.NewFileManager(t)
		fileManager.On("Open", "/path/to/profile.mobileprovision").Return(nil, errors.New("file not found"))

		reader := NewProfileReader(log.NewLogger(), fileManager, mocks.NewPathModifier(t), mocks.NewPathProvider(t))
		_, err := reader.ProvisioningProfileInfoFromFile("/path/to/profile.mobileprovision")

		require.Error(t, err)
	})

	t.Run("parses iOS profile from file", func(t *testing.T) {
		f := newPKCS7TempFile(t, iosDevelopmentProfileContent)
		fileManager := mocks.NewFileManager(t)
		fileManager.On("Open", f.Name()).Return(f, nil)

		reader := NewProfileReader(log.NewLogger(), fileManager, mocks.NewPathModifier(t), mocks.NewPathProvider(t))
		got, err := reader.ProvisioningProfileInfoFromFile(f.Name())

		require.NoError(t, err)
		require.Equal(t, ProfileTypeIos, got.Type)
		require.Equal(t, "4b617a5f-e31e-4edc-9460-718a5abacd05", got.UUID)
	})
}

func TestProfileReader_InstalledProvisioningProfileInfos(t *testing.T) {
	const (
		modernTilde = "~/Library/Developer/Xcode/UserData/Provisioning Profiles"
		legacyTilde = "~/Library/MobileDevice/Provisioning Profiles"
		modernAbs   = "/Users/user/Library/Developer/Xcode/UserData/Provisioning Profiles"
		legacyAbs   = "/Users/user/Library/MobileDevice/Provisioning Profiles"
	)

	t.Run("list profiles error is propagated", func(t *testing.T) {
		pathModifier := mocks.NewPathModifier(t)
		pathModifier.On("AbsPath", modernTilde).Return("", errors.New("access denied"))

		reader := NewProfileReader(log.NewLogger(), mocks.NewFileManager(t), pathModifier, mocks.NewPathProvider(t))
		_, err := reader.InstalledProvisioningProfileInfos(ProfileTypeIos)

		require.Error(t, err)
	})

	t.Run("file open error is propagated", func(t *testing.T) {
		const profilePath = "/Users/user/Library/Developer/Xcode/UserData/Provisioning Profiles/uuid.mobileprovision"
		pathModifier := mocks.NewPathModifier(t)
		pathModifier.On("AbsPath", modernTilde).Return(modernAbs, nil)
		pathModifier.On("AbsPath", legacyTilde).Return(legacyAbs, nil)
		pathModifier.On("EscapeGlobPath", modernAbs).Return(modernAbs)
		pathModifier.On("EscapeGlobPath", legacyAbs).Return(legacyAbs)
		pathProvider := mocks.NewPathProvider(t)
		pathProvider.On("Glob", filepath.Join(modernAbs, "*"+IOSExtension)).Return([]string{profilePath}, nil)
		pathProvider.On("Glob", filepath.Join(legacyAbs, "*"+IOSExtension)).Return([]string{}, nil)
		fileManager := mocks.NewFileManager(t)
		fileManager.On("Open", profilePath).Return(nil, errors.New("permission denied"))

		reader := NewProfileReader(log.NewLogger(), fileManager, pathModifier, pathProvider)
		_, err := reader.InstalledProvisioningProfileInfos(ProfileTypeIos)

		require.Error(t, err)
	})

	t.Run("returns parsed profiles", func(t *testing.T) {
		f := newPKCS7TempFile(t, iosDevelopmentProfileContent)
		pathModifier := mocks.NewPathModifier(t)
		pathModifier.On("AbsPath", modernTilde).Return(modernAbs, nil)
		pathModifier.On("AbsPath", legacyTilde).Return(legacyAbs, nil)
		pathModifier.On("EscapeGlobPath", modernAbs).Return(modernAbs)
		pathModifier.On("EscapeGlobPath", legacyAbs).Return(legacyAbs)
		pathProvider := mocks.NewPathProvider(t)
		pathProvider.On("Glob", filepath.Join(modernAbs, "*"+IOSExtension)).Return([]string{f.Name()}, nil)
		pathProvider.On("Glob", filepath.Join(legacyAbs, "*"+IOSExtension)).Return([]string{}, nil)
		fileManager := mocks.NewFileManager(t)
		fileManager.On("Open", f.Name()).Return(f, nil)

		reader := NewProfileReader(log.NewLogger(), fileManager, pathModifier, pathProvider)
		got, err := reader.InstalledProvisioningProfileInfos(ProfileTypeIos)

		require.NoError(t, err)
		require.Len(t, got, 1)
		require.Equal(t, ProfileTypeIos, got[0].Type)
		require.Equal(t, "4b617a5f-e31e-4edc-9460-718a5abacd05", got[0].UUID)
	})
}

func TestProfileReader_ListProfiles(t *testing.T) {
	const (
		modernTilde = "~/Library/Developer/Xcode/UserData/Provisioning Profiles"
		legacyTilde = "~/Library/MobileDevice/Provisioning Profiles"
		modernAbs   = "/Users/user/Library/Developer/Xcode/UserData/Provisioning Profiles"
		legacyAbs   = "/Users/user/Library/MobileDevice/Provisioning Profiles"
	)

	setupPathMocks := func(pathModifier *mocks.PathModifier, pathProvider *mocks.PathProvider, uuid, ext string, modernResults, legacyResults []string) {
		pathModifier.On("AbsPath", modernTilde).Return(modernAbs, nil)
		pathModifier.On("AbsPath", legacyTilde).Return(legacyAbs, nil)
		pathModifier.On("EscapeGlobPath", modernAbs).Return(modernAbs)
		pathModifier.On("EscapeGlobPath", legacyAbs).Return(legacyAbs)
		pathProvider.On("Glob", filepath.Join(modernAbs, uuid+ext)).Return(modernResults, nil)
		pathProvider.On("Glob", filepath.Join(legacyAbs, uuid+ext)).Return(legacyResults, nil)
	}

	t.Run("iOS type uses .mobileprovision extension", func(t *testing.T) {
		uuid := "abc123"
		pathModifier := mocks.NewPathModifier(t)
		pathProvider := mocks.NewPathProvider(t)
		modernResult := []string{filepath.Join(modernAbs, uuid+IOSExtension)}
		setupPathMocks(pathModifier, pathProvider, uuid, IOSExtension, modernResult, []string{})

		reader := newTestProfileReader(t, pathModifier, pathProvider)
		got, err := reader.ListProfiles(ProfileTypeIos, uuid)

		require.NoError(t, err)
		require.Equal(t, modernResult, got)
	})

	t.Run("macOS type uses .provisionprofile extension", func(t *testing.T) {
		uuid := "abc123"
		pathModifier := mocks.NewPathModifier(t)
		pathProvider := mocks.NewPathProvider(t)
		modernResult := []string{filepath.Join(modernAbs, uuid+MacExtension)}
		setupPathMocks(pathModifier, pathProvider, uuid, MacExtension, modernResult, []string{})

		reader := newTestProfileReader(t, pathModifier, pathProvider)
		got, err := reader.ListProfiles(ProfileTypeMacOs, uuid)

		require.NoError(t, err)
		require.Equal(t, modernResult, got)
	})

	t.Run("results from both dirs are concatenated", func(t *testing.T) {
		uuid := "*"
		pathModifier := mocks.NewPathModifier(t)
		pathProvider := mocks.NewPathProvider(t)
		modernResults := []string{filepath.Join(modernAbs, "profile1.mobileprovision")}
		legacyResults := []string{filepath.Join(legacyAbs, "profile2.mobileprovision")}
		setupPathMocks(pathModifier, pathProvider, uuid, IOSExtension, modernResults, legacyResults)

		reader := newTestProfileReader(t, pathModifier, pathProvider)
		got, err := reader.ListProfiles(ProfileTypeIos, uuid)

		require.NoError(t, err)
		require.Equal(t, append(modernResults, legacyResults...), got)
	})

	t.Run("AbsPath error is propagated", func(t *testing.T) {
		pathModifier := mocks.NewPathModifier(t)
		pathModifier.On("AbsPath", modernTilde).Return("", errors.New("access denied"))

		reader := newTestProfileReader(t, pathModifier, mocks.NewPathProvider(t))
		_, err := reader.ListProfiles(ProfileTypeIos, "uuid")

		require.Error(t, err)
	})

	t.Run("Glob error is propagated", func(t *testing.T) {
		pathModifier := mocks.NewPathModifier(t)
		pathModifier.On("AbsPath", modernTilde).Return(modernAbs, nil)
		pathModifier.On("AbsPath", legacyTilde).Return(legacyAbs, nil)
		pathModifier.On("EscapeGlobPath", modernAbs).Return(modernAbs)
		pathProvider := mocks.NewPathProvider(t)
		pathProvider.On("Glob", filepath.Join(modernAbs, "*"+IOSExtension)).Return(nil, errors.New("glob failed"))

		reader := newTestProfileReader(t, pathModifier, pathProvider)
		_, err := reader.ListProfiles(ProfileTypeIos, "*")

		require.Error(t, err)
	})
}

func TestProfileReader_ProvisioningProfilesDirPath(t *testing.T) {
	const (
		modernTilde = "~/Library/Developer/Xcode/UserData/Provisioning Profiles"
		legacyTilde = "~/Library/MobileDevice/Provisioning Profiles"
		modernAbs   = "/Users/user/Library/Developer/Xcode/UserData/Provisioning Profiles"
		legacyAbs   = "/Users/user/Library/MobileDevice/Provisioning Profiles"
	)

	tests := []struct {
		name               string
		xcodeMajorVersion  int64
		expectedAbsPathArg string
		returnPath         string
	}{
		{
			name:               "xcode 0 (unknown) uses modern path",
			xcodeMajorVersion:  0,
			expectedAbsPathArg: modernTilde,
			returnPath:         modernAbs,
		},
		{
			name:               "xcode 16 uses modern path",
			xcodeMajorVersion:  16,
			expectedAbsPathArg: modernTilde,
			returnPath:         modernAbs,
		},
		{
			name:               "xcode 17 uses modern path",
			xcodeMajorVersion:  17,
			expectedAbsPathArg: modernTilde,
			returnPath:         modernAbs,
		},
		{
			name:               "xcode 15 uses legacy path",
			xcodeMajorVersion:  15,
			expectedAbsPathArg: legacyTilde,
			returnPath:         legacyAbs,
		},
		{
			name:               "xcode 1 uses legacy path",
			xcodeMajorVersion:  1,
			expectedAbsPathArg: legacyTilde,
			returnPath:         legacyAbs,
		},
		{
			name:               "modern path resolves tilde to $HOME",
			xcodeMajorVersion:  16,
			expectedAbsPathArg: modernTilde,
			returnPath:         filepath.Join(os.Getenv("HOME"), "Library/Developer/Xcode/UserData/Provisioning Profiles"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pathModifier := mocks.NewPathModifier(t)
			pathModifier.On("AbsPath", tt.expectedAbsPathArg).Return(tt.returnPath, nil)

			reader := newTestProfileReader(t, pathModifier, mocks.NewPathProvider(t))
			got, err := reader.ProvisioningProfilesDirPath(tt.xcodeMajorVersion)

			require.NoError(t, err)
			require.Equal(t, tt.returnPath, got)
		})
	}

	t.Run("AbsPath error is propagated", func(t *testing.T) {
		pathModifier := mocks.NewPathModifier(t)
		pathModifier.On("AbsPath", modernTilde).Return("", errors.New("access denied"))

		reader := newTestProfileReader(t, pathModifier, mocks.NewPathProvider(t))
		_, err := reader.ProvisioningProfilesDirPath(16)

		require.Error(t, err)
	})
}

func newTestProfileReader(t *testing.T, pathModifier *mocks.PathModifier, pathProvider *mocks.PathProvider) ProfileReader {
	return NewProfileReader(log.NewLogger(), mocks.NewFileManager(t), pathModifier, pathProvider)
}

// newPKCS7TempFile writes the given plist content into a PKCS7 envelope, saves it to a
// temp file, and returns the open file rewound to the beginning.
func newPKCS7TempFile(t *testing.T, content string) *os.File {
	t.Helper()
	sd, err := pkcs7.NewSignedData([]byte(content))
	require.NoError(t, err)
	pkcs7Bytes, err := sd.Finish()
	require.NoError(t, err)

	f, err := os.CreateTemp(t.TempDir(), "*.mobileprovision")
	require.NoError(t, err)
	_, err = f.Write(pkcs7Bytes)
	require.NoError(t, err)
	_, err = f.Seek(0, 0)
	require.NoError(t, err)
	return f
}
