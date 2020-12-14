package xcarchive

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/assert"
)

const (
	tempDirName  = "__artifacts__"
	DSYMSDirName = "dSYMs"
)

func TestIsMacOS(t *testing.T) {
	tests := []struct {
		name     string
		archPath string
		want     bool
		wantErr  bool
	}{
		{
			name:     "macOS",
			archPath: filepath.Join(sampleRepoPath(t), "archives/macos.xcarchive"),
			want:     true,
			wantErr:  false,
		},
		{
			name:     "iOS",
			archPath: filepath.Join(sampleRepoPath(t), "archives/ios.xcarchive"),
			want:     false,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsMacOS(tt.archPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsMacOS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsMacOS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_GivenArchiveWithMultipleAppAndFrameworkDSYMs_WhenFindDSYMsCalled_ThenExpectAllDSYMsToBeReturned(t *testing.T) {
	testCases := []struct {
		name                   string
		numberOfAppDSYMs       int
		numberOfFrameworkDSYMs int
	}{
		{
			name:                   "1. Given archive with multiple app and framework dSYMs when FindDSYMs called then expect all dSYMs to be returned",
			numberOfAppDSYMs:       2,
			numberOfFrameworkDSYMs: 2,
		},
		{
			name:                   "2. Given archive with singe app and framework dSYMs when FindDSYMs called then expect both dSYMs to be returned",
			numberOfAppDSYMs:       1,
			numberOfFrameworkDSYMs: 1,
		},
		{
			name:                   "3. Given archive with multiple app dSYMs when FindDSYMs called then expect all app dSYMs to be returned",
			numberOfAppDSYMs:       2,
			numberOfFrameworkDSYMs: 0,
		},
		{
			name:                   "4. Given archive with multiple framework dSYMs when FindDSYMs called then expect all framework dSYMs to be returned",
			numberOfAppDSYMs:       0,
			numberOfFrameworkDSYMs: 2,
		},
		{
			name:                   "5. Given archive with single app dSYM when FindDSYMs called then expect the app dSYM to be returned",
			numberOfAppDSYMs:       1,
			numberOfFrameworkDSYMs: 0,
		},
		{
			name:                   "6. Given archive with single framework dSYM when FindDSYMs called then expect the framework dSYM to be returned",
			numberOfAppDSYMs:       0,
			numberOfFrameworkDSYMs: 1,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			archivePath, err := createArchiveWithAppAndFrameworkDSYMs(
				"archives/ios.dsyms.xcarchive",
				testCase.numberOfAppDSYMs,
				testCase.numberOfFrameworkDSYMs,
			)
			assert.NoError(t, err)

			appDSYMs, frameworkDSYMs, err := findDSYMs(archivePath)
			assert.NoError(t, err)
			assert.Equal(t, testCase.numberOfAppDSYMs, len(appDSYMs))
			assert.Equal(t, testCase.numberOfFrameworkDSYMs, len(frameworkDSYMs))
		})
	}
}

func Test_GivenArchiveWithNoDSYMs_WhenFindDSYMsCalled_ThenExpectAnError(t *testing.T) {
	archivePath, err := createArchiveWithAppAndFrameworkDSYMs("archives/ios.nodsyms.xcarchive", 0, 0)
	assert.NoError(t, err)

	_, _, err = findDSYMs(archivePath)
	assert.Error(t, errNoDsymFound, err)
}

func createArchiveWithAppAndFrameworkDSYMs(archivePath string, numberOfAppDSYMs, numberOfFrameworkDSYMs int) (string, error) {
	archivePath, err := createArchive(archivePath)
	if err != nil {
		return "", err
	}

	err = createAppDSYMs(archivePath, numberOfAppDSYMs)
	if err != nil {
		return "", err
	}

	err = createFrameworkDSYMs(archivePath, numberOfFrameworkDSYMs)
	if err != nil {
		return "", err
	}

	return archivePath, nil
}

func createAppDSYMs(archivePath string, numberOfDSYMs int) error {
	return createDSYMs(archivePath, "app", numberOfDSYMs)
}

func createFrameworkDSYMs(archivePath string, numberOfDSYMs int) error {
	return createDSYMs(archivePath, "framework", numberOfDSYMs)
}

func createDSYMs(archivePath, dSYMType string, numberOfDSYMs int) error {
	for i := 0; i < numberOfDSYMs; i++ {
		err := ioutil.WriteFile(createDSYMFilePath(archivePath, dSYMType, i), nil, 0777)
		if err != nil {
			return err
		}
	}

	return nil
}

func createDSYMFilePath(archivePath, dSYMType string, index int) string {
	return filepath.Join(archivePath, DSYMSDirName, fmt.Sprintf("ios-%d.%s.dSYM", index, dSYMType))
}

func createArchive(archivePath string) (string, error) {
	tempDirPath, err := pathutil.NormalizedOSTempDirPath(tempDirName)
	if err != nil {
		return "", err
	}

	archivePath = filepath.Join(tempDirPath, archivePath)
	err = os.MkdirAll(filepath.Join(archivePath, DSYMSDirName), 0755)
	if err != nil {
		return "", err
	}

	return archivePath, nil
}
