package zip

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
)

// DittoReader ...
type DittoReader struct {
	logger       log.Logger
	archivePath  string
	extractedDir string
}

// IsDittoReaderAvailable ...
func IsDittoReaderAvailable() bool {
	_, err := exec.LookPath("ditto")
	return err == nil
}

// NewDittoReader ...
func NewDittoReader(archivePath string, logger log.Logger) *DittoReader {
	return &DittoReader{
		logger:      logger,
		archivePath: archivePath,
	}
}

// ReadFile ...
func (r *DittoReader) ReadFile(relPthPattern string) ([]byte, error) {
	if r.extractedDir == "" {
		if err := r.extractArchive(); err != nil {
			return nil, fmt.Errorf("failed to extract archive: %w", err)
		}
	}

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
func (r *DittoReader) Close() error {
	if r.extractedDir == "" {
		return nil
	}
	return os.RemoveAll(r.extractedDir)
}

func (r *DittoReader) extractArchive() error {
	tmpDir, err := pathutil.NewPathProvider().CreateTempDir("ditto_reader")
	if err != nil {
		return nil
	}

	/*
	   -x            Extract the archives given as source arguments. The format
	                 is assumed to be CPIO, unless -k is given.  Compressed CPIO
	                 is automatically handled.

	   -k            Create or extract from a PKZip archive instead of the
	                 default CPIO.  PKZip archives should be stored in filenames
	                 ending in .zip.
	*/
	factory := command.NewFactory(env.NewRepository())
	cmd := factory.Create("ditto", []string{"-x", "-k", r.archivePath, tmpDir}, nil)
	if out, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		fmt.Println(out)
		return err
	}

	r.extractedDir = tmpDir

	return nil
}
