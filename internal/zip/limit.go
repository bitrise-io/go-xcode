package zip

import (
	"errors"
	"fmt"
	"io"
)

// MaxFileSize is the size of the largest file ReadFile returns, in bytes.
// The files read from an archive (Info.plist, embedded.mobileprovision) are far smaller,
// while a small archive can hold a file of any size.
const MaxFileSize = 10 << 20

// ErrFileTooLarge is returned by ReadFile for a file larger than MaxFileSize.
var ErrFileTooLarge = errors.New("file is too large")

func readAll(r io.Reader, name string) ([]byte, error) {
	b, err := io.ReadAll(io.LimitReader(r, MaxFileSize+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", name, err)
	}
	if len(b) > MaxFileSize {
		return nil, fmt.Errorf("failed to read %s: %w: it exceeds %d bytes", name, ErrFileTooLarge, MaxFileSize)
	}

	return b, nil
}
