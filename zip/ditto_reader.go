package zip

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/bitrise-io/go-utils/v2/log"

	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/pathutil"
)

type dittoReader struct {
	extractedDir string
	logger       log.Logger
}

// IsDittoReaderAvailable ...
func IsDittoReaderAvailable() bool {
	_, err := exec.LookPath("ditto")
	return err == nil
}

// NewDittoReader ...
func NewDittoReader(archivePath string, logger log.Logger) (ReadCloser, error) {
	factory := command.NewFactory(env.NewRepository())
	tmpDir, err := pathutil.NewPathProvider().CreateTempDir("ditto_reader")
	if err != nil {
		return nil, err
	}

	/*
	   -x            Extract the archives given as source arguments. The format
	                 is assumed to be CPIO, unless -k is given.  Compressed CPIO
	                 is automatically handled.

	   -k            Create or extract from a PKZip archive instead of the
	                 default CPIO.  PKZip archives should be stored in filenames
	                 ending in .zip.
	*/
	cmd := factory.Create("ditto", []string{"-x", "-k", archivePath, tmpDir}, nil)
	if out, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		fmt.Println(out)
		return nil, err
	}

	return dittoReader{
		extractedDir: tmpDir,
		logger:       logger,
	}, nil
}

// ReadFile ...
func (r dittoReader) ReadFile(relPthPattern string) ([]byte, error) {
	absPthPattern := filepath.Join(r.extractedDir, relPthPattern)
	matches, err := filepath.Glob(absPthPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to find file with pattern: %s: %w", absPthPattern, err)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no file found with pattern: %s", absPthPattern)
	}

	sort.Strings(matches)

	pth := matches[0]
	f, err := os.Open(pth)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", pth, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			r.logger.Warnf("Failed to close %s: %s", pth, err)
		}
	}()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", pth, err)
	}

	return b, nil
}

// Close ...
func (r dittoReader) Close() error {
	return os.RemoveAll(r.extractedDir)
}
