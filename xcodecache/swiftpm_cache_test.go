package cache

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/bitrise-io/go-utils/v2/env"

	v1command "github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/command/git"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/v2/command"
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

func getXcodebuildCmd(xcodeProjectPath string) command.Command {
	f := command.NewFactory(env.NewRepository())
	c := f.Create("xcodebuild", []string{
		"-project", xcodeProjectPath,
		"-scheme", "sample-swiftpm2",
		"-resolvePackageDependencies",
	}, &command.Opts{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	return c
}

func TestCollectSwiftPackages(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test as -short flag is set.")
	}

	tempDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatalf("setup: failed to create temp dir, error: %s", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			log.Warnf("failed to remove temp dir, error: %s", err)
		}
	}()

	actualProjectDir := path.Join(tempDir, "project")
	if err := os.Mkdir(actualProjectDir, os.ModePerm); err != nil {
		t.Fatalf("setup: failed to create cache dir: %s", err)
	}

	checkoutDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatalf("setup: failed to create temp dir, error: %s", err)
	}
	defer func() {
		if err := os.RemoveAll(checkoutDir); err != nil {
			log.Warnf("failed to remove temp dir, error: %s", err)
		}
	}()

	cacheTempDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatalf("setup: failed to create temp dir, error: %s", err)
	}
	defer func() {
		if err := os.RemoveAll(cacheTempDir); err != nil {
			log.Warnf("failed to remove temp dir, error: %s", err)
		}
	}()

	cacheDir := path.Join(cacheTempDir, "cache")
	if err := os.Mkdir(cacheDir, os.ModePerm); err != nil {
		t.Fatalf("setup: failed to create cache dir: %s", err)
	}

	gitCommand, err := git.New(checkoutDir)
	if err != nil {
		t.Fatalf("setup: failed to create git project, error: %s", err)
	}
	if err := gitCommand.Clone("https://github.com/bitrise-io/sample-apps-ios-swiftpm").Run(); err != nil {
		t.Fatalf("setup: failed to clone sample project repo, error: %s", err)
	}

	// Build xcode project for the first time with no swift packages cache.
	initialProject := path.Join(checkoutDir, "sample-swiftpm4")
	cleanTime, _ := resolveProject(t, actualProjectDir, cacheDir, initialProject)

	// Build xcode project for the second time with swift packages cached and compare build times.
	cachedTime, packagesPath := resolveProject(t, actualProjectDir, cacheDir, initialProject)

	t.Logf("Clean cache: %s Build with cache: %s", cleanTime, cachedTime)
	if cleanTime*7/10 < cachedTime {
		t.Fatalf("cached build is not much shorter than clean build")
	}

	// Compare swift packages content to the cached one, check that no files changed on the second build.
	// This ensures that no new cache is created on every build, even if the project did not change.
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

	// Change swift packages
	changeTime, _ := resolveProject(t, actualProjectDir, cacheDir, path.Join(checkoutDir, "sample-swiftpm7"))
	t.Logf("Change time; %s", changeTime)
}

func resolveProject(t *testing.T, projectPath, cacheDir, xcodeProjSourceDir string) (resolveTime time.Duration, packagesPath string) {
	// Remove and copy new Xcode project directory
	if err := os.RemoveAll(projectPath); err != nil {
		t.Fatalf("failed to remove temp dir, error: %s", err)
	}
	if err := os.Mkdir(projectPath, os.ModePerm); err != nil {
		t.Fatalf("failed to create project dir: %s", err)
	}
	if err := v1command.CopyDir(projectPath, xcodeProjSourceDir, true); err != nil {
		t.Fatalf("setup: failed to copy sample project: %s", err)
	}

	xcodeProjPath := path.Join(xcodeProjSourceDir, "sample-swiftpm2.xcodeproj")
	packagesPath, err := NewSwiftPackageCache().SwiftPackagesPath(xcodeProjPath)
	if err != nil {
		t.Fatalf("failed to get Swift packages path, err: %s", err)
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
	if err := v1command.CopyDir(cacheDir, packagesPath, true); err != nil {
		t.Fatalf("failed to sync directory, error: %s", err)
	}

	// Resolve Xcode packages
	resolveStart := time.Now()
	log.Donef("$ %s", getXcodebuildCmd(xcodeProjPath).PrintableCommandArgs())
	exitCode, err := getXcodebuildCmd(xcodeProjPath).RunAndReturnExitCode()
	if err != nil {
		t.Fatalf("failed to run xcodebuild command, error: %s", err)
	}
	if exitCode != 0 {
		t.Fatalf("xcodebuild exited with exit code: %d", exitCode)
	}
	resolveTime = time.Since(resolveStart)

	t.Logf("Resolution time: %s, source %s", resolveTime, xcodeProjSourceDir)

	// Check that swift packages path exists.
	if _, err := os.Stat(packagesPath); err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("swift packages directory path does not exist, error: %s", err)
		}
		t.Fatalf("failed to get file info, error: %s", err)
	}

	// Simulating the cache-push step by saving packages cache content.
	// Removing manifest.db, as it changes after a rebuild, and it would invalidate cache  if included.
	if err := os.Remove(path.Join(packagesPath, "manifest.db")); err != nil {
		t.Logf("failed to remove file, error: %s", err)
	}
	if err := os.RemoveAll(cacheDir); err != nil {
		t.Fatalf("failed to remove cache directory: %s", err)
	}
	if err := os.Mkdir(cacheDir, os.ModePerm); err != nil {
		t.Fatalf("failed to create cache dir: %s", err)
	}
	if err := v1command.CopyDir(packagesPath, cacheDir, true); err != nil {
		t.Fatalf("failed to sync directory, error: %s", err)
	}

	return resolveTime, packagesPath
}
