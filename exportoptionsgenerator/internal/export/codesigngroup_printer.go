package export

import (
	"encoding/json"
	"fmt"

	"github.com/bitrise-io/go-utils/v2/log"
)

type CodeSignGroupPrinter struct {
	logger log.Logger
}

// NewCodeSignGroupPrinter ...
func NewCodeSignGroupPrinter(logger log.Logger) *CodeSignGroupPrinter {
	return &CodeSignGroupPrinter{
		logger: logger,
	}
}

// ToDebugString ...
func (printer CodeSignGroupPrinter) ToDebugString(group SelectableCodeSignGroup) string {
	printable := map[string]any{}
	printable["team"] = fmt.Sprintf("%s (%s)", group.Certificate.TeamName, group.Certificate.TeamID)
	printable["certificate"] = fmt.Sprintf("%s (%s)", group.Certificate.CommonName, group.Certificate.Serial)

	bundleIDProfiles := map[string][]string{}
	for bundleID, profileInfos := range group.BundleIDProfilesMap {
		printableProfiles := []string{}
		for _, profileInfo := range profileInfos {
			printableProfiles = append(printableProfiles, fmt.Sprintf("%s (%s)", profileInfo.Name, profileInfo.UUID))
		}
		bundleIDProfiles[bundleID] = printableProfiles
	}
	printable["bundle_id_profiles"] = bundleIDProfiles

	data, err := json.MarshalIndent(printable, "", "\t")
	if err != nil {
		printer.logger.Errorf("Failed to marshal: %v, error: %s", printable, err)
		return ""
	}

	return string(data)
}
