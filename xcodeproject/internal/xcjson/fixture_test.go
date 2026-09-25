package xcjson

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseProjectFile(t *testing.T) {
	doc := parseFixture(t, "SimpleApp.xcproj")

	require.Equal(t, "Release", doc.Root.Lookup("default-configuration").String)
	require.Equal(t, 2700.0, doc.Root.Lookup("last-upgrade").Number)

	targets := doc.Root.Lookup("targets")
	require.Len(t, targets.Elements, 2)

	app := targets.Elements[0]
	require.Equal(t, "SimpleApp", app.Lookup("name").String)
	require.Equal(t, "manual", app.Lookup("legacy-provisioning-style").String)

	settings := app.Lookup("build-settings")
	require.Equal(t, "Apple Development", settings.Lookup("CODE_SIGN_IDENTITY").String)
	require.Equal(t, "Apple Distribution", settings.Lookup("CODE_SIGN_IDENTITY[config=Release][sdk=iphoneos*]").String)
	require.True(t, targets.Elements[1].Lookup("build-settings").Lookup("GENERATE_INFOPLIST_FILE").Bool)

	script := app.Lookup("build-phases").Elements[3].Lookup("script")
	require.Equal(t, []string{"set -euo pipefail", `echo "building $CONFIGURATION"`}, stringElements(script))
}

func TestParseHandEditedProjectFile(t *testing.T) {
	doc := parseFixture(t, "hand-edited.json5")

	require.Equal(t, "2A9F3C1D4B8E5A6700000001", doc.Root.Lookup("id").String)
	require.Equal(t, "Bitrise", doc.Root.Lookup("organization").String)
	require.Equal(t, 2700.0, doc.Root.Lookup("last-upgrade").Number)
	require.Equal(t, 150.0, doc.Root.Lookup("timeout").Number)
	require.Equal(t, 0.25, doc.Root.Lookup("slack").Number)
	require.Equal(t, 12.0, doc.Root.Lookup("budget").Number)
	require.True(t, math.IsInf(doc.Root.Lookup("unbounded").Number, 1))
	require.Equal(t, KindNull, doc.Root.Lookup("disabled").Kind)

	settings := doc.Root.Lookup("targets").Elements[0].Lookup("build-settings")
	require.Equal(t, "-D FIRST -D SECOND", settings.Lookup("OTHER_SWIFT_FLAGS").String)
}

// TestSpliceLeavesTheRestOfTheFileAlone exercises the reason spans exist: a
// writer rewrites one value in place and every byte it does not understand,
// comments and formatting included, comes through unchanged.
func TestSpliceLeavesTheRestOfTheFileAlone(t *testing.T) {
	doc := parseFixture(t, "SimpleApp.xcproj")

	version := doc.Root.Lookup("targets").Elements[0].Lookup("build-settings").Lookup("SWIFT_VERSION")
	spliced := string(doc.Source[:version.Start]) + `"6.1"` + string(doc.Source[version.End:])

	require.Equal(t, strings.Replace(string(doc.Source), `"SWIFT_VERSION": "6.0"`, `"SWIFT_VERSION": "6.1"`, 1), spliced)

	reparsed, err := parse([]byte(spliced))
	require.NoError(t, err)
	require.Equal(t, "6.1", reparsed.Root.Lookup("targets").Elements[0].Lookup("build-settings").Lookup("SWIFT_VERSION").String)
}

func parseFixture(t *testing.T, name string) *Document {
	t.Helper()

	src, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err)

	doc, err := parse(src)
	require.NoError(t, err)

	return doc
}

func stringElements(node *Node) []string {
	values := make([]string, 0, len(node.Elements))
	for _, element := range node.Elements {
		values = append(values, element.String)
	}
	return values
}
