package xcscheme

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
)

// BuildableReference ...
type BuildableReference struct {
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
	BuildForTesting    string `xml:"buildForTesting,attr"`
	BuildForArchiving  string `xml:"buildForArchiving,attr"`
	BuildableReference BuildableReference
}

// BuildAction ...
type BuildAction struct {
	BuildActionEntries []BuildActionEntry `xml:"BuildActionEntries>BuildActionEntry"`
}

// TestableReference ...
type TestableReference struct {
	Skipped            string `xml:"skipped,attr"`
	BuildableReference BuildableReference
}

// TestAction ...
type TestAction struct {
	Testables          []TestableReference `xml:"Testables>TestableReference"`
	BuildConfiguration string              `xml:"buildConfiguration,attr"`
}

// ArchiveAction ...
type ArchiveAction struct {
	BuildConfiguration string `xml:"buildConfiguration,attr"`
}

// Scheme ...
type Scheme struct {
	BuildAction   BuildAction
	ArchiveAction ArchiveAction
	TestAction    TestAction

	Name string `xml:"-"`
	Path string `xml:"-"`
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
