package xcscheme

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-xcode/xcodeproject/xcodeproj"
)

const (
	yes         = "YES"
	no          = "NO"
	buildableID = "primary"
)

// BuildableReference ...
type BuildableReference struct {
	BuildableIdentifier string `xml:"BuildableIdentifier,attr"`
	BlueprintIdentifier string `xml:"BlueprintIdentifier,attr"`
	BuildableName       string `xml:"BuildableName,attr"`
	BlueprintName       string `xml:"BlueprintName,attr"`
	ReferencedContainer string `xml:"ReferencedContainer,attr"`
}

// IsAppReference ...
func (r BuildableReference) IsAppReference() bool {
	return filepath.Ext(r.BuildableName) == ".app"
}

// ReferencedContainerAbsPath ...
func (r BuildableReference) ReferencedContainerAbsPath(schemeContainerDir string) (string, error) {
	s := strings.Split(r.ReferencedContainer, ":")
	if len(s) != 2 {
		return "", fmt.Errorf("unknown referenced container (%s)", r.ReferencedContainer)
	}

	base := s[1]
	absPth := filepath.Join(schemeContainerDir, base)

	return pathutil.AbsPath(absPth)
}

// BuildActionEntry ...
type BuildActionEntry struct {
	BuildForTesting   string `xml:"buildForTesting,attr"`
	BuildForRunning   string `xml:"buildForRunning,attr"`
	BuildForProfiling string `xml:"buildForProfiling,attr"`
	BuildForArchiving string `xml:"buildForArchiving,attr"`
	BuildForAnalyzing string `xml:"buildForAnalyzing,attr"`

	BuildableReference BuildableReference
}

// BuildAction ...
type BuildAction struct {
	ParallelizeBuildables     string             `xml:"parallelizeBuildables,attr"`
	BuildImplicitDependencies string             `xml:"buildImplicitDependencies,attr"`
	BuildActionEntries        []BuildActionEntry `xml:"BuildActionEntries>BuildActionEntry"`
}

func newBuildAction(target xcodeproj.Target, projectName string) BuildAction {
	return BuildAction{
		ParallelizeBuildables:     yes,
		BuildImplicitDependencies: yes,
		BuildActionEntries: []BuildActionEntry{
			{
				BuildForTesting:   yes,
				BuildForRunning:   yes,
				BuildForProfiling: yes,
				BuildForArchiving: yes,
				BuildForAnalyzing: yes,
				BuildableReference: BuildableReference{
					BuildableIdentifier: buildableID,
					BlueprintIdentifier: target.ID,
					BuildableName:       path.Base(target.ProductReference.Path),
					BlueprintName:       target.Name,
					ReferencedContainer: fmt.Sprintf("container:%s", projectName),
				},
			},
		},
	}
}

// TestableReference ...
type TestableReference struct {
	Skipped            string `xml:"skipped,attr"`
	BuildableReference BuildableReference
}

// MacroExpansion ...
type MacroExpansion struct {
	BuildableReference BuildableReference
}

// AdditionalOptions ...
type AdditionalOptions struct {
}

// TestAction ...
type TestAction struct {
	BuildConfiguration           string `xml:"buildConfiguration,attr"`
	SelectedDebuggerIdentifier   string `xml:"selectedDebuggerIdentifier,attr"`
	SelectedLauncherIdentifier   string `xml:"selectedLauncherIdentifier,attr"`
	ShouldUseLaunchSchemeArgsEnv string `xml:"shouldUseLaunchSchemeArgsEnv,attr"`

	Testables         []TestableReference `xml:"Testables>TestableReference"`
	MacroExpansion    MacroExpansion
	AdditionalOptions AdditionalOptions
}

// ArchiveAction ...
type ArchiveAction struct {
	BuildConfiguration       string `xml:"buildConfiguration,attr"`
	RevealArchiveInOrganizer string `xml:"revealArchiveInOrganizer,attr"`
}

// Scheme ...
type Scheme struct {
	LastUpgradeVersion string `xml:"LastUpgradeVersion,attr"`
	Version            string `xml:"version,attr"`

	BuildAction   BuildAction
	TestAction    TestAction
	ArchiveAction ArchiveAction

	Name     string `xml:"-"`
	Path     string `xml:"-"`
	IsShared bool   `xml:"-"`
}

// Open ...
func Open(pth string) (Scheme, error) {
	b, err := fileutil.ReadBytesFromFile(pth)
	if err != nil {
		return Scheme{}, err
	}

	var scheme Scheme
	if err := xml.Unmarshal(b, &scheme); err != nil {
		return Scheme{}, fmt.Errorf("failed to unmarshal scheme file: %s, error: %s", pth, err)
	}

	scheme.Name = strings.TrimSuffix(filepath.Base(pth), filepath.Ext(pth))
	scheme.Path = pth

	return scheme, nil
}

func NewScheme(buildTarget xcodeproj.Target, testTarget []xcodeproj.Target, projectname string) Scheme {
	return Scheme{
		LastUpgradeVersion: "1240",
		Version:            "1.3",
		BuildAction:        newBuildAction(buildTarget, projectname),
	}
}

type XMLToken int

const (
	Invalid XMLToken = iota
	XMLStart
	XMLEnd
	XMLAttribute
)

// Write ...
func (s Scheme) Write(pth string) error {
	contents, err := xml.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to marshal Scheme: %v", err)
	}

	contentsNewline := strings.ReplaceAll(string(contents), "><", ">\n<")
	// Place XML Attributes on separate lines
	contentsNewline = strings.ReplaceAll(contentsNewline, " ", "\n")

	var contentsIndented string

	indent := 0
	for _, line := range strings.Split(contentsNewline, "\n") {
		currentLine := XMLAttribute
		if strings.HasPrefix(line, "</") {
			currentLine = XMLEnd
		} else if strings.HasPrefix(line, "<") {
			currentLine = XMLStart
		}

		if currentLine == XMLAttribute {
			line = strings.Replace(line, "=", " = ", 1)
		}

		if currentLine == XMLEnd && indent != 0 {
			indent--
		}

		contentsIndented += strings.Repeat("   ", indent)
		contentsIndented += line + "\n"

		if currentLine == XMLStart {
			indent++
		}
	}

	if err := ioutil.WriteFile(pth, []byte(xml.Header+contentsIndented), 0600); err != nil {
		return fmt.Errorf("failed to write Scheme file (%s): %v", pth, err)
	}

	return nil
}

// AppBuildActionEntry ...
func (s Scheme) AppBuildActionEntry() (BuildActionEntry, bool) {
	var entry BuildActionEntry
	for _, e := range s.BuildAction.BuildActionEntries {
		if e.BuildForArchiving != "YES" {
			continue
		}
		if !e.BuildableReference.IsAppReference() {
			continue
		}
		entry = e
		break
	}

	return entry, (entry.BuildableReference.BlueprintIdentifier != "")
}
