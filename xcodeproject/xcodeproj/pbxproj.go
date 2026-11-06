package xcodeproj

import (
	"fmt"

	"github.com/bitrise-io/go-plist"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
)

const (
	customAnnotationKey = plist.CustomAnnotationKey
	startKey            = plist.CustomAnnotationStartKey
	endKey              = plist.CustomAnnotationEndKey
)

// pbxprojFile is project.pbxproj, the property list format.
type pbxprojFile struct {
	logger log.Logger
	// path is the project bundle's path, which project-relative file paths resolve against.
	path string

	// rawProj is the source of truth content writes out; every edit lands here.
	rawProj serialized.Object
	format  int
	// originalContents is what content diffs against to rewrite only the changed objects.
	originalContents []byte

	projectID        string
	targetAttributes serialized.Object
	// buildPhaseIDs maps target IDs to the IDs of their build phases.
	buildPhaseIDs map[string][]string
}

func (f Factory) newPBXProj(content []byte, projectPath string) (*xcodeProject, error) {
	return parsePBXProj(content, f.newXcodeProject(projectPath))
}

func (p *pbxprojFile) setBuildSetting(target Target, configuration BuildConfiguration, key string, value any) error {
	objects, ok := p.rawProj.Object("objects")
	if !ok {
		return fmt.Errorf("failed to parse project: no objects")
	}
	rawConfiguration, ok := objects.Object(configuration.ID)
	if !ok {
		return fmt.Errorf("build configuration %s of target %s not found", configuration.ID, target.Name)
	}
	buildSettings, ok := rawConfiguration.Object("buildSettings")
	if !ok {
		return fmt.Errorf("build configuration %s of target %s has no buildSettings", configuration.ID, target.Name)
	}

	buildSettings[key] = value

	return nil
}

func (p *pbxprojFile) setManualSigning(target Target, developmentTeam string) {
	if attributes, ok := p.targetAttributes.Object(target.ID); ok {
		attributes["ProvisioningStyle"] = "Manual"
		attributes["DevelopmentTeam"] = developmentTeam
		attributes["DevelopmentTeamName"] = ""
	}
}

// content rewrites only the changed objects of project.pbxproj, so the rest of the file stays byte
// for byte. If that is not possible, it marshals the whole file and logs a warning.
func (p *pbxprojFile) content() ([]byte, error) {
	content, err := p.perObjectModify()
	if err == nil {
		return content, nil
	}

	p.logger.Warnf("failed to modify project in-place: %v", err)

	content, err = plist.MarshalIndent(p.rawProj, p.format, "\t")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal .pbxproj: %w", err)
	}

	return content, nil
}
