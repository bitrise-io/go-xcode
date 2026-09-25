package xcjson

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// FuzzParse guards the two properties the rest of the package is built on: a
// hand-written scanner must not panic on arbitrary bytes, and a node's span
// must bound a complete value, because that is what lets a writer replace one
// value by splicing over its bytes.
func FuzzParse(f *testing.F) {
	for _, name := range []string{"SimpleApp.xcproj", "hand-edited.json5"} {
		src, err := os.ReadFile(filepath.Join("testdata", name))
		require.NoError(f, err)
		f.Add(src)
	}
	for _, seed := range []string{
		``, `{}`, `[]`, `null`, `0x1p`, `.5`, `5.`, `-Infinity`, `"😀"`,
		`{a: 'b', /* c */ "d": [1,], }`, "// comment\n{}", `"\`, `{"a"`, `[[[[`,
	} {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, src []byte) {
		doc, err := parse(src)
		if err != nil {
			var syntaxErr *SyntaxError
			require.ErrorAs(t, err, &syntaxErr)
			require.GreaterOrEqual(t, syntaxErr.Offset, 0)
			require.LessOrEqual(t, syntaxErr.Offset, len(src))
			return
		}

		walk(doc.Root, func(node *Node) {
			require.GreaterOrEqual(t, node.Start, 0)
			require.GreaterOrEqual(t, node.End, node.Start)
			require.LessOrEqual(t, node.End, len(src))

			_, err := parse(doc.Raw(node))
			require.NoError(t, err, "a node's span is not a document on its own")
		})
	})
}
