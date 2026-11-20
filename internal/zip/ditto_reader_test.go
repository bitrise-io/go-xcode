package zip

import (
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/stretchr/testify/require"
)

func TestDittoReader_ReadFile_rejectsFileOverMaxSize(t *testing.T) {
	if !IsDittoReaderAvailable() {
		t.Skip("ditto is not available")
	}

	reader := NewDittoReader(createArchive(t, testEntryName, MaxFileSize+1), log.NewLogger())
	defer func() { require.NoError(t, reader.Close()) }()

	_, err := reader.ReadFile("Payload/*.app/Info.plist")

	require.ErrorIs(t, err, ErrFileTooLarge)
}
