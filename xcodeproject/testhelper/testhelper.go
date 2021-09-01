package testhelper

import (
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/env"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

var clonedRespos = map[string]string{}

// GitCloneIntoTmpDir ...
func GitCloneIntoTmpDir(t *testing.T, repo string) string {
	if tmpDir, ok := clonedRespos[repo]; ok {
		return tmpDir
	}

	tmpDir, err := pathutil.NormalizedOSTempDirPath("__xcode-proj__")
	require.NoError(t, err)

	f := command.NewFactory(env.NewRepository())
	cmd := f.Create("git", []string{"clone", repo, tmpDir}, nil)

	require.NoError(t, cmd.Run())

	clonedRespos[repo] = tmpDir

	return tmpDir
}

// GitCloneBranchIntoTmpDir clones a specific branch from a git repository
func GitCloneBranchIntoTmpDir(t *testing.T, repo string, branch string) string {
	tmpDir, err := pathutil.NormalizedOSTempDirPath("__xcode-proj__")
	require.NoError(t, err)

	f := command.NewFactory(env.NewRepository())
	cmd := f.Create("git", []string{"clone", "-b", branch, repo, tmpDir}, nil)

	require.NoError(t, cmd.Run())

	return tmpDir
}

// CreateTmpFile ...
func CreateTmpFile(t *testing.T, name, content string) string {
	tmpDir, err := pathutil.NormalizedOSTempDirPath("__xcode-proj__")
	require.NoError(t, err)

	pth := filepath.Join(tmpDir, name)
	require.NoError(t, fileutil.WriteStringToFile(pth, content))
	return pth
}
