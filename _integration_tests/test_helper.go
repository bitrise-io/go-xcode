package _integration_tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bitrise-io/go-utils/command/git"
	"github.com/stretchr/testify/require"
)

const sampleArtifactsRepoURL = "https://github.com/bitrise-io/sample-artifacts.git"

var reposToDir map[string]map[string]string

func GetSampleArtifactsRepository(t *testing.T) string {
	return GetRepository(t, sampleArtifactsRepoURL, "master")
}

func GetRepository(t *testing.T, url, branch string) string {
	if repoDir := getRepoDir(url, branch); repoDir != "" {
		return repoDir
	}

	tmpDir := createDirForRepo(t, url, branch)
	gitCommand, err := git.New(tmpDir)
	require.NoError(t, err)

	out, err := gitCommand.Clone(url, "--depth=1", "--branch", branch).RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)

	saveRepoDir(tmpDir, url, branch)

	return tmpDir
}

func getRepoDir(url, branch string) string {
	if reposToDir == nil {
		return ""
	}

	branchToDir, ok := reposToDir[url]
	if !ok {
		return ""
	}

	dir, ok := branchToDir[branch]
	if !ok {
		return ""
	}
	return dir
}

func saveRepoDir(dir, url, branch string) {
	if reposToDir == nil {
		reposToDir = map[string]map[string]string{}
	}

	branchToDir, ok := reposToDir[url]
	if !ok {
		branchToDir = map[string]string{}
	}

	branchToDir[branch] = dir
	reposToDir[url] = branchToDir
}

func createDirForRepo(t *testing.T, repo, branch string) string {
	tmpDir, err := os.MkdirTemp("", "go-xcode")
	require.NoError(t, err)

	repoRootDir := strings.TrimSuffix(filepath.Base(repo), filepath.Ext(repo))
	pth := filepath.Join(tmpDir, repoRootDir, branch)
	err = os.MkdirAll(pth, os.ModePerm)
	require.NoError(t, err)

	return pth
}
