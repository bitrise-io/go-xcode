package xcodeproj

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-xcode/xcodeproject/testhelper"
	"github.com/stretchr/testify/require"
)

func Test_GivenNewlyGeneratedXcodeProject_WhenListingSchemes_ThenReturnsTheDefaultSchemes(t *testing.T) {
	xcodeProjectPath := newlyGeneratedXcodeProjectPath(t)
	proj, err := Open(xcodeProjectPath)
	require.NoError(t, err)

	schemes, err := proj.Schemes()
	require.NoError(t, err)

	expectedSchemeNames := []string{"ios-sample"}
	require.Equal(t, len(expectedSchemeNames), len(schemes))
	for _, expectedSchemeName := range expectedSchemeNames {
		schemeFound := false
		for _, scheme := range schemes {
			if scheme.Name == expectedSchemeName {
				schemeFound = true
				break
			}
		}
		require.True(t, schemeFound)
	}
}

func ensureTmpTestdataDir(t *testing.T) string {
	_, callerFilename, _, ok := runtime.Caller(0)
	require.True(t, ok)
	callerDir := filepath.Dir(callerFilename)
	callerPackageDir := filepath.Dir(callerDir)
	packageTmpTestdataDir := filepath.Join(callerPackageDir, "_testdata")
	if _, err := os.Stat(packageTmpTestdataDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(packageTmpTestdataDir, os.ModePerm)
		require.NoError(t, err)
	}
	return packageTmpTestdataDir
}

func newlyGeneratedXcodeProjectPath(t *testing.T) string {
	testdataDir := ensureTmpTestdataDir(t)
	newlyGeneratedXcodeProjectDir := filepath.Join(testdataDir, "newly_generated_xcode_project")
	_, err := os.Stat(newlyGeneratedXcodeProjectDir)
	newlyGeneratedXcodeProjectDirExist := !errors.Is(err, os.ErrNotExist)
	if newlyGeneratedXcodeProjectDirExist {
		cmd := command.New("git", "clean", "-f", "-x", "-d")
		cmd.SetDir(newlyGeneratedXcodeProjectDir)
		require.NoError(t, cmd.Run())
	} else {
		repo := "https://github.com/godrei/ios-sample.git"
		branch := "main"
		testhelper.GitCloneBranch(t, repo, branch, newlyGeneratedXcodeProjectDir)
	}
	return filepath.Join(newlyGeneratedXcodeProjectDir, "ios-sample.xcodeproj")
}
