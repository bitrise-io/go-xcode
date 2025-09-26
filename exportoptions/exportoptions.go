package exportoptions

import (
	"fmt"

	"github.com/bitrise-io/go-plist"
)

type FileWriter interface {
	WriteBytes(pth string, data []byte) error
}

// ExportOptions ...
type ExportOptions interface {
	Hash() map[string]interface{}
	String() (string, error)
	WriteToFile(pth string, fileWriter FileWriter) error
}

func writePlistToFile(options map[string]interface{}, pth string, fileWriter FileWriter) error {
	plistBytes, err := plist.MarshalIndent(options, plist.XMLFormat, "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal export options model, error: %s", err)
	}
	if err := fileWriter.WriteBytes(pth, plistBytes); err != nil {
		return fmt.Errorf("failed to write export options, error: %s", err)
	}

	return nil
}
