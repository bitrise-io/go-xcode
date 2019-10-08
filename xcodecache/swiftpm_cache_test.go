package cache

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/command/git"
	"github.com/bitrise-io/go-utils/log"
)

// fileContentHash returns file's md5 content hash.
func fileContentHash(pth string) (string, error) {
	f, err := os.Open(pth)
	if err != nil {
		return "", err
	}

	defer func() {
		if err := f.Close(); err != nil {
			log.Errorf("Failed to close file (%s), error: %+v", pth, err)
		}
	}()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func getXcodebuildCmd(xcodeProjectPath string) *command.Model {
	buildCmd := command.New("xcodebuild", "build",
		"-project", xcodeProjectPath,
		"-scheme", "sample swiftpm",
		"-configuration", "Debug",
		"-destination", "platform=iOS Simulator,name=iPhone 8,OS=latest",
		`CODE_SIGNING_ALLOWED="NO"`)
	return buildCmd.SetStdout(os.Stdout).SetStderr(os.Stderr)
}

func TestCollectSwiftPackages(t *testing.T) {
	xcodeProjDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("setup: failed to create temp dir, error: %s", err)
	}
	defer func() {
		if err := os.RemoveAll(xcodeProjDir); err != nil {
			log.Warnf("failed to remove temp dir, error: %s", err)
		}
	}()

	cacheDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("setup: failed to create temp dir, error: %s", err)
	}
	defer func() {
		if err := os.RemoveAll(cacheDir); err != nil {
			log.Warnf("failed to remove temp dir, error: %s", err)
		}
	}()

	gitCommand, err := git.New(xcodeProjDir)
	if err != nil {
		t.Fatalf("setup: failed to create git project, error: %s", err)
	}
	if err := gitCommand.Clone("https://github.com/bitrise-io/sample-apps-ios-swiftpm").Run(); err != nil {
		t.Fatalf("setup: failed to clone sample project repo, error: %s", err)
	}

	xcodeProjPath := path.Join(xcodeProjDir, "sample-swiftpm.xcodeproj")
	packagesPath, err := SwiftPackagesPath(xcodeProjPath)
	if err != nil {
		t.Fatalf("failed to get Swift packages path, err: %s", err)
	}

	if err := os.RemoveAll(packagesPath); err != nil {
		t.Fatalf("setup: failed to remove cache dir, err: %s", err)
	}

	// Build xcode project for the first time with no swift packages cache.
	cleanStartTime := time.Now()
	exitCode, err := getXcodebuildCmd(xcodeProjPath).RunAndReturnExitCode()
	if err != nil {
		t.Fatalf("failed to run xcodebuild command, error: %s", err)
	}
	if exitCode != 0 {
		t.Fatalf("xcodebuild exited with exit code: %d", exitCode)
	}
	cleanBuildTime := time.Since(cleanStartTime)

	// Check that swift packages path exists.
	if _, err := os.Stat(packagesPath); err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("swift packages directory path does not exist, error: %s", err)
		}
		t.Fatalf("failed to get file info, error: %s", err)
	}

	// Removing manifest.db, as it changes after a rebuild, and it would invalidate cache  if included.
	if err := os.Remove(path.Join(packagesPath, "manifest.db")); err != nil {
		t.Fatalf("failed to remove file, error: %s", err)
	}

	// Simulating the cache-push step by saving packages cache content.
	if err := command.CopyDir(packagesPath, cacheDir, true); err != nil {
		t.Fatalf("failed to sync directory, error: %s", err)
	}

	// Remove DerivedData
	projectDerivedData, err := xcodeProjectDerivedDataPath(xcodeProjPath)
	if err != nil {
		t.Fatalf("failed to get project DerivedData path, error: %s", err)
	}
	if err := os.RemoveAll(projectDerivedData); err != nil {
		t.Fatalf("setup: failed to remove project DerivedData dir, err: %s", err)
	}

	// Simulate the cache-pull step by restoring cached folder
	if err := os.MkdirAll(packagesPath, 0770); err != nil {
		t.Fatalf("failed to create directory, error: %s", err)
	}
	if err := command.CopyDir(cacheDir, packagesPath, true); err != nil {
		t.Fatalf("failed to sync directory, error: %s", err)
	}

	// Build xcode project for the second time with swift packages cached and compare build times.
	cachedStartTime := time.Now()
	exitCode, err = getXcodebuildCmd(xcodeProjPath).RunAndReturnExitCode()
	if err != nil {
		t.Fatalf("failed to run xcodebuild command, error: %s", err)
	}
	if exitCode != 0 {
		t.Fatalf("xcodebuild exited with exit code: %d", exitCode)
	}
	cachedBuildTime := time.Since(cachedStartTime)

	t.Logf("Clean cache: %s Build with cache: %s", cleanBuildTime, cachedBuildTime)
	if cachedBuildTime > cleanBuildTime*7/10 {
		t.Errorf("cached build is not much shorter than clean build")
	}

	// Compare swift packages content to the cached one, check that no files changed on the second build.
	// This ensures that no new cache is created on every build, even if the project did not change.
	if err := os.Remove(path.Join(packagesPath, "manifest.db")); err != nil {
		t.Fatalf("failed to remove file, error: %s", err)
	}
	if err := filepath.Walk(packagesPath, func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fileInfo.IsDir() {
			return nil
		}

		relCachedFilePath, err := filepath.Rel(packagesPath, path)
		if err != nil {
			t.Fatalf("failed to get relative path, error: %s", err)
		}
		cachedFilePath := filepath.Join(cacheDir, relCachedFilePath)

		cachedFileInfo, err := os.Stat(cachedFilePath)
		if err != nil {
			if os.IsNotExist(err) {
				t.Fatalf("Cache content changed: file does not exist in cache, error: %s", err)
			}
			t.Fatalf("failed to get file info, error: %s", err)
		}

		if fileInfo.Size() != cachedFileInfo.Size() {
			t.Fatalf("Cache content changed: sizes do not match for file %s: %d != %d", relCachedFilePath, fileInfo.Size(), cachedFileInfo.Size())
		}

		fileHash, err := fileContentHash(path)
		if err != nil {
			t.Fatalf("failed to get file hash, error: %s", err)
		}
		cachedFileHash, err := fileContentHash(cachedFilePath)
		if err != nil {
			t.Fatalf("failed to get file hash, error: %s", err)
		}
		if fileHash != cachedFileHash {
			t.Fatalf("Cache content changed: different file content for files: %s %s", path, cachedFilePath)
		}
		return nil
	}); err != nil {
		t.Fatalf("%s", err)
	}
}
