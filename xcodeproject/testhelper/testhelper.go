package testhelper

import (
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

var clonedRepos = map[string]string{}

// GitCloneIntoTmpDir ...
func GitCloneIntoTmpDir(t *testing.T, repo string) string {
	if tmpDir, ok := clonedRepos[repo]; ok {
		return tmpDir
	}

	tmpDir, err := pathutil.NormalizedOSTempDirPath("__xcode-proj__")
	require.NoError(t, err)

	cmd := command.New("git", "clone", repo, tmpDir)
	require.NoError(t, cmd.Run())

	clonedRepos[repo] = tmpDir

	return tmpDir
}

// GitCloneBranchIntoTmpDir clones a specific branch from a git repository.
func GitCloneBranchIntoTmpDir(t *testing.T, repo string, branch string) string {
	tmpDir, err := pathutil.NormalizedOSTempDirPath("__xcode-proj__")
	require.NoError(t, err)

	GitCloneBranch(t, repo, branch, tmpDir)

	return tmpDir
}

// GitCloneBranch clones a branch from a git repository into an existing directory.
func GitCloneBranch(t *testing.T, repo string, branch string, dir string) {
	cmd := command.New("git", "clone", "--depth", "1", "--branch", branch, repo, dir)
	require.NoError(t, cmd.Run())
}

// CreateTmpFile ...
func CreateTmpFile(t *testing.T, name, content string) string {
	tmpDir, err := pathutil.NormalizedOSTempDirPath("__xcode-proj__")
	require.NoError(t, err)

	pth := filepath.Join(tmpDir, name)
	require.NoError(t, fileutil.WriteStringToFile(pth, content))
	return pth
}
