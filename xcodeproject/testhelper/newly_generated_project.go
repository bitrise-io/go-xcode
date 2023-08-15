package testhelper

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

// NewlyGeneratedXcodeProjectPath ...
func NewlyGeneratedXcodeProjectPath(t *testing.T) string {
	testdataDir := ensureTmpTestdataDir(t)
	newlyGeneratedXcodeProjectDir := filepath.Join(testdataDir, "newly_generated_xcode_project")
	_, err := os.Stat(newlyGeneratedXcodeProjectDir)
	exist := !errors.Is(err, os.ErrNotExist)
	if exist {
		cmd := command.New("git", "clean", "-f", "-x", "-d")
		cmd.SetDir(newlyGeneratedXcodeProjectDir)
		require.NoError(t, cmd.Run())
	} else {
		repo := "https://github.com/godrei/ios-sample.git"
		branch := "main"
		GitCloneBranch(t, repo, branch, newlyGeneratedXcodeProjectDir)
	}
	return filepath.Join(newlyGeneratedXcodeProjectDir, "ios-sample.xcodeproj")
}

// NewlyGeneratedXcodeWorkspacePath ...
func NewlyGeneratedXcodeWorkspacePath(t *testing.T) string {
	testdataDir := ensureTmpTestdataDir(t)
	newlyGeneratedXcodeWorkspaceDir := filepath.Join(testdataDir, "newly_generated_xcode_workspace")
	_, err := os.Stat(newlyGeneratedXcodeWorkspaceDir)
	exist := !errors.Is(err, os.ErrNotExist)
	if exist {
		cmd := command.New("git", "clean", "-f", "-x", "-d")
		cmd.SetDir(newlyGeneratedXcodeWorkspaceDir)
		require.NoError(t, cmd.Run())
	} else {
		repo := "https://github.com/godrei/ios-sample.git"
		branch := "workspace"
		GitCloneBranch(t, repo, branch, newlyGeneratedXcodeWorkspaceDir)
	}
	return filepath.Join(newlyGeneratedXcodeWorkspaceDir, "ios-sample.xcworkspace")
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
