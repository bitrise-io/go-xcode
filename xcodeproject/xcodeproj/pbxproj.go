package xcodeproj

import (
	"fmt"
	"reflect"
	"sort"

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
		return fmt.Errorf("failed to set build setting %s: project has no objects", key)
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

type objectChange struct {
	start, end int
	content    []byte
}

// perObjectModify splices re-serialised changed objects into the original bytes.
// The original and annotated trees are decoded from originalContents here rather than kept from
// parse time, so an opened project holds no extra copies of the tree; Save is called rarely.
func (p *pbxprojFile) perObjectModify() ([]byte, error) {
	var annotated serialized.Object
	if _, err := plist.UnmarshalWithCustomAnnotation(p.originalContents, &annotated); err != nil {
		return nil, fmt.Errorf("failed to decode original project: %w", err)
	}

	original, _, err := decodePBXProj(p.originalContents)
	if err != nil {
		return nil, err
	}

	objectsModified, ok := p.rawProj.Object("objects")
	if !ok {
		return nil, fmt.Errorf("failed to parse project: no objects")
	}
	objectsOriginal, ok := original.Object("objects")
	if !ok {
		return nil, fmt.Errorf("failed to parse original project: no objects")
	}
	objectsAnnotated, ok := annotated.Object("objects")
	if !ok {
		return nil, fmt.Errorf("failed to parse annotated project: no objects")
	}

	var changes []objectChange
	for id := range objectsModified {
		objectModified, ok := objectsModified.Object(id)
		if !ok {
			return nil, fmt.Errorf("object %s is not a dictionary", id)
		}

		objectOriginal, ok := objectsOriginal.Object(id)
		if !ok {
			return nil, fmt.Errorf("new object added, not in original project: %s", id)
		}

		if reflect.DeepEqual(objectOriginal, objectModified) {
			continue
		}

		objectAnnotated, ok := objectsAnnotated.Object(id)
		if !ok {
			return nil, fmt.Errorf("new object added, not in original annotated project: %s", id)
		}

		position, ok := objectAnnotated.Object(customAnnotationKey)
		if !ok {
			return nil, fmt.Errorf("no raw object position available for %s", id)
		}
		start, ok := position.Int64(startKey)
		if !ok {
			return nil, fmt.Errorf("no raw object start position available for %s", id)
		}
		end, ok := position.Int64(endKey)
		if !ok {
			return nil, fmt.Errorf("no raw object end position available for %s", id)
		}

		content, err := plist.MarshalIndent(objectModified, p.format, "\t")
		if err != nil {
			return nil, fmt.Errorf("could not marshal object %s: %w", id, err)
		}

		changes = append(changes, objectChange{start: int(start), end: int(end), content: content})
	}

	if len(changes) == 0 {
		return p.originalContents, nil
	}

	sort.Slice(changes, func(i, j int) bool {
		if changes[i].start == changes[j].start {
			return changes[i].end < changes[j].end
		}
		return changes[i].start < changes[j].start
	})

	var result []byte
	previousEnd := 0
	for i, change := range changes {
		if i < len(changes)-1 && change.end >= changes[i+1].start {
			return nil, fmt.Errorf("overlapping changes: %d, %d", change.end, changes[i+1].start)
		}

		result = append(result, p.originalContents[previousEnd:change.start]...)
		result = append(result, change.content...)
		previousEnd = change.end
	}

	if previousEnd <= len(p.originalContents)-1 {
		result = append(result, p.originalContents[previousEnd:]...)
	}

	return result, nil
}
