package utility

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindFileInAppDir(t *testing.T) {
	t.Run("direct match at appDir root", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "Info.plist")
		require.NoError(t, os.WriteFile(target, []byte(""), 0o600))

		got, err := FindFileInAppDir(dir, "Info.plist")
		require.NoError(t, err)
		require.Equal(t, target, got)
	})

	t.Run("match inside a .app bundle", func(t *testing.T) {
		dir := t.TempDir()
		bundle := filepath.Join(dir, "MyApp.app")
		require.NoError(t, os.MkdirAll(bundle, 0o755))
		target := filepath.Join(bundle, "embedded.mobileprovision")
		require.NoError(t, os.WriteFile(target, []byte(""), 0o600))

		got, err := FindFileInAppDir(dir, "embedded.mobileprovision")
		require.NoError(t, err)
		require.Equal(t, target, got)
	})

	t.Run("match inside a nested directory within .app bundle", func(t *testing.T) {
		dir := t.TempDir()
		nested := filepath.Join(dir, "MyApp.app", "Frameworks", "Foo.framework")
		require.NoError(t, os.MkdirAll(nested, 0o755))
		target := filepath.Join(nested, "Info.plist")
		require.NoError(t, os.WriteFile(target, []byte(""), 0o600))

		got, err := FindFileInAppDir(dir, "Info.plist")
		require.NoError(t, err)
		require.Equal(t, target, got)
	})

	t.Run("not found returns error", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "Other.app"), 0o755))

		_, err := FindFileInAppDir(dir, "missing.file")
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing.file")
	})

	t.Run("direct match takes precedence over .app search", func(t *testing.T) {
		dir := t.TempDir()
		direct := filepath.Join(dir, "shared")
		require.NoError(t, os.WriteFile(direct, []byte("root"), 0o600))

		bundle := filepath.Join(dir, "MyApp.app")
		require.NoError(t, os.MkdirAll(bundle, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(bundle, "shared"), []byte("bundle"), 0o600))

		got, err := FindFileInAppDir(dir, "shared")
		require.NoError(t, err)
		require.Equal(t, direct, got)
	})
}
