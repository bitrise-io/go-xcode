package exportoptions

import (
	"fmt"

	plist "github.com/DHowett/go-plist"
)

// AppStoreOptionsModel ...
type AppStoreOptionsModel struct {
	TeamID string

	// for app-store exports
	UploadBitcode bool
	UploadSymbols bool
}

// NewAppStoreOptions ...
func NewAppStoreOptions() AppStoreOptionsModel {
	return AppStoreOptionsModel{
		UploadBitcode: UploadBitcodeDefault,
		UploadSymbols: UploadSymbolsDefault,
	}
}

// Hash ...
func (options AppStoreOptionsModel) Hash() map[string]interface{} {
	hash := map[string]interface{}{}
	hash[MethodKey] = MethodAppStore
	if options.TeamID != "" {
		hash[TeamIDKey] = options.TeamID
	}
	if options.UploadBitcode != UploadBitcodeDefault {
		hash[UploadBitcodeKey] = options.UploadBitcode
	}
	if options.UploadSymbols != UploadSymbolsDefault {
		hash[UploadSymbolsKey] = options.UploadSymbols
	}
	return hash
}

// String ...
func (options AppStoreOptionsModel) String() (string, error) {
	hash := options.Hash()
	plistBytes, err := plist.MarshalIndent(hash, plist.XMLFormat, "\t")
	if err != nil {
		return "", fmt.Errorf("failed to marshal export options model, error: %s", err)
	}
	return string(plistBytes), err
}

// WriteToFile ...
func (options AppStoreOptionsModel) WriteToFile(pth string) error {
	return WritePlistToFile(options.Hash(), pth)
}

// WriteToTmpFile ...
func (options AppStoreOptionsModel) WriteToTmpFile() (string, error) {
	return WritePlistToTmpFile(options.Hash())
}
