package metaparser

import (
	"time"

	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/exportoptions"
)

// ArtifactMetadata ...
type ArtifactMetadata struct {
	AppInfo          Info          `json:"app_info"`
	FileSizeBytes    int64         `json:"file_size_bytes"`
	ProvisioningInfo ProvisionInfo `json:"provisioning_info,omitempty"`
	Scheme           string        `json:"scheme,omitempty"`
}

// Info ...
type Info struct {
	AppTitle          string   `json:"app_title"`
	BundleID          string   `json:"bundle_id"`
	Version           string   `json:"version"`
	BuildNumber       string   `json:"build_number"`
	MinOSVersion      string   `json:"min_OS_version"`
	DeviceFamilyList  []uint64 `json:"device_family_list"`
	RawPackageContent string   `json:"-"`
}

// ProvisionInfo ...
type ProvisionInfo struct {
	CreationDate         time.Time            `json:"creation_date"`
	ExpireDate           time.Time            `json:"expire_date"`
	DeviceUDIDList       []string             `json:"device_UDID_list"`
	TeamName             string               `json:"team_name"`
	ProfileName          string               `json:"profile_name"`
	ProvisionsAllDevices bool                 `json:"provisions_all_devices"`
	IPAExportMethod      exportoptions.Method `json:"ipa_export_method"`
}

// Parser ...
type Parser struct {
	logger      log.Logger
	fileManager fileutil.FileManager
}

// New ...
func New(logger log.Logger, fileManager fileutil.FileManager) *Parser {
	return &Parser{
		logger:      logger,
		fileManager: fileManager,
	}
}
