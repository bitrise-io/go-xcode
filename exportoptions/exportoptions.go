package exportoptions

import (
	"fmt"

	"github.com/bitrise-io/go-plist"
	"github.com/bitrise-io/go-utils/fileutil"
)

// ExportOptions ...
type ExportOptions interface {
	Hash() map[string]interface{}
	String() (string, error)
	WriteToFile(pth string) error
}

// WritePlistToFile ...
func WritePlistToFile(options map[string]interface{}, pth string) error {
	plistBytes, err := plist.MarshalIndent(options, plist.XMLFormat, "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal export options model, error: %s", err)
	}
	if err := fileutil.WriteBytesToFile(pth, plistBytes); err != nil {
		return fmt.Errorf("failed to write export options, error: %s", err)
	}

	return nil
}
