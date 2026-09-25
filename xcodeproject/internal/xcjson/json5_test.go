package xcjson

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseValues(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want any
	}{
		{name: "null", src: `null`, want: nil},
		{name: "true", src: `true`, want: true},
		{name: "false", src: `false`, want: false},
		{name: "string", src: `"value"`, want: "value"},
		{name: "empty object", src: `{}`, want: map[string]any{}},
		{name: "empty array", src: `[]`, want: []any{}},
		{name: "array", src: `[1, "two", false, null]`, want: []any{1.0, "two", false, nil}},
		{name: "object", src: `{"a": 1, "b": [true]}`, want: map[string]any{"a": 1.0, "b": []any{true}}},
		{name: "nested", src: `{"a": {"b": {"c": []}}}`, want: map[string]any{"a": map[string]any{"b": map[string]any{"c": []any{}}}}},
		{name: "leading byte order mark", src: "\uFEFF{}", want: map[string]any{}},
		{name: "surrounding whitespace", src: "\t\r\n {} \n", want: map[string]any{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			doc, err := parse([]byte(test.src))
			require.NoError(t, err)
			require.Equal(t, test.want, plain(doc.Root))
		})
	}
}

func TestParseNumbers(t *testing.T) {
	tests := []struct {
		src  string
		want float64
	}{
		{src: `0`, want: 0},
		{src: `-0`, want: math.Copysign(0, -1)},
		{src: `123`, want: 123},
		{src: `-123`, want: -123},
		{src: `+123`, want: 123},
		{src: `1.5`, want: 1.5},
		{src: `.5`, want: 0.5},
		{src: `-.5`, want: -0.5},
		{src: `5.`, want: 5},
		{src: `1e3`, want: 1000},
		{src: `1E+3`, want: 1000},
		{src: `1.5e-2`, want: 0.015},
		{src: `0x0A8C`, want: 2700},
		{src: `0XFF`, want: 255},
		{src: `-0xff`, want: -255},
		{src: `Infinity`, want: math.Inf(1)},
		{src: `+Infinity`, want: math.Inf(1)},
		{src: `-Infinity`, want: math.Inf(-1)},
	}

	for _, test := range tests {
		t.Run(test.src, func(t *testing.T) {
			doc, err := parse([]byte(test.src))
			require.NoError(t, err)
			require.Equal(t, KindNumber, doc.Root.Kind)
			require.Equal(t, test.want, doc.Root.Number)
		})
	}
}

func TestParseNotANumber(t *testing.T) {
	for _, src := range []string{`NaN`, `+NaN`, `-NaN`} {
		t.Run(src, func(t *testing.T) {
			doc, err := parse([]byte(src))
			require.NoError(t, err)
			require.Equal(t, KindNumber, doc.Root.Kind)
			require.True(t, math.IsNaN(doc.Root.Number))
		})
	}
}

func TestParseStrings(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{name: "double quoted", src: `"value"`, want: "value"},
		{name: "single quoted", src: `'value'`, want: "value"},
		{name: "empty", src: `""`, want: ""},
		{name: "opposite quote is literal", src: `"it's"`, want: "it's"},
		{name: "escaped quote", src: `"say \"hi\""`, want: `say "hi"`},
		{name: "control escapes", src: `"a\b\f\n\r\t\vb"`, want: "a\b\f\n\r\t\vb"},
		{name: "null escape", src: `"a\0b"`, want: "a\x00b"},
		{name: "backslash", src: `"C:\\Users"`, want: `C:\Users`},
		{name: "hex escape", src: `"\x41\xe9"`, want: "A\u00e9"},
		{name: "unicode escape", src: `"\u00e9"`, want: "é"},
		{name: "surrogate pair", src: `"\ud83d\ude00"`, want: "😀"},
		{name: "lone surrogate", src: `"\ud83d"`, want: "\uFFFD"},
		{name: "identity escape", src: `"\q\/"`, want: "q/"},
		{name: "line continuation", src: "\"one \\\ntwo\"", want: "one two"},
		{name: "crlf line continuation", src: "\"one \\\r\ntwo\"", want: "one two"},
		{name: "paragraph separator is literal", src: "\"a\u2029b\"", want: "a\u2029b"},
		{name: "starts with an escape", src: `"\tvalue"`, want: "\tvalue"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			doc, err := parse([]byte(test.src))
			require.NoError(t, err)
			require.Equal(t, KindString, doc.Root.Kind)
			require.Equal(t, test.want, doc.Root.String)
		})
	}
}

func TestParseJSON5Dialect(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want any
	}{
		{name: "unquoted keys", src: `{id: 1, $ref: 2, _x: 3, a1: 4}`, want: map[string]any{"id": 1.0, "$ref": 2.0, "_x": 3.0, "a1": 4.0}},
		{name: "unquoted key spelling a literal", src: `{null: 1, true: 2, NaN: 3}`, want: map[string]any{"null": 1.0, "true": 2.0, "NaN": 3.0}},
		{name: "escaped key", src: `{\u0061b: 1}`, want: map[string]any{"ab": 1.0}},
		{name: "single quoted key", src: `{'a-b': 1}`, want: map[string]any{"a-b": 1.0}},
		{name: "trailing comma in object", src: `{"a": 1,}`, want: map[string]any{"a": 1.0}},
		{name: "trailing comma in array", src: `[1,]`, want: []any{1.0}},
		{name: "line comment", src: "// lead\n{\"a\": 1} // trail", want: map[string]any{"a": 1.0}},
		{name: "block comment", src: `/* lead */ {"a": /* mid */ 1} /* trail */`, want: map[string]any{"a": 1.0}},
		{name: "comment between members", src: "{\n\"a\": 1, // why\n\"b\": 2,\n}", want: map[string]any{"a": 1.0, "b": 2.0}},
		{name: "unterminated line comment at end", src: `{"a": 1} // trail`, want: map[string]any{"a": 1.0}},
		{name: "block comment spanning lines", src: "{\n/* one\n   two */\n\"a\": 1}", want: map[string]any{"a": 1.0}},
		{name: "non-ascii whitespace", src: "{\u00a0\"a\":\u2028 1\u2029}", want: map[string]any{"a": 1.0}},
		{name: "non-ascii unquoted key", src: `{café: 1, ᾩ2: 2}`, want: map[string]any{"café": 1.0, "ᾩ2": 2.0}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			doc, err := parse([]byte(test.src))
			require.NoError(t, err)
			require.Equal(t, test.want, plain(doc.Root))
		})
	}
}

func TestParseSpansBoundValuesExactly(t *testing.T) {
	src := []byte(`{ "a": [ 1, "two" ], /* c */ "b": { "c": true }, }`)

	doc, err := parse(src)
	require.NoError(t, err)

	require.Equal(t, `{ "a": [ 1, "two" ], /* c */ "b": { "c": true }, }`, string(doc.Raw(doc.Root)))
	require.Equal(t, `[ 1, "two" ]`, string(doc.Raw(doc.Root.Lookup("a"))))
	require.Equal(t, `1`, string(doc.Raw(doc.Root.Lookup("a").Elements[0])))
	require.Equal(t, `"two"`, string(doc.Raw(doc.Root.Lookup("a").Elements[1])))
	require.Equal(t, `{ "c": true }`, string(doc.Raw(doc.Root.Lookup("b"))))
	require.Equal(t, `true`, string(doc.Raw(doc.Root.Lookup("b").Lookup("c"))))

	key := doc.Root.Members[1]
	require.Equal(t, `"b"`, string(src[key.KeyStart:key.KeyEnd]))
}

func TestParseSpansRoundTrip(t *testing.T) {
	for _, name := range []string{"SimpleApp.xcproj", "hand-edited.json5"} {
		t.Run(name, func(t *testing.T) {
			doc := parseFixture(t, name)
			walk(doc.Root, func(node *Node) {
				reparsed, err := parse(doc.Raw(node))
				require.NoError(t, err, "%q is not a document on its own", doc.Raw(node))
				require.Equal(t, plain(node), plain(reparsed.Root))
			})
		})
	}
}

func TestLookup(t *testing.T) {
	doc, err := parse([]byte(`{"a": 1, "b": 2, "a": 3}`))
	require.NoError(t, err)

	require.Equal(t, 3.0, doc.Root.Lookup("a").Number, "the last of two members with the same key wins")
	require.Equal(t, 2.0, doc.Root.Lookup("b").Number)
	require.Nil(t, doc.Root.Lookup("missing"))
	require.Nil(t, doc.Root.Lookup("a").Lookup("a"), "a number has no members")
	require.Nil(t, (*Node)(nil).Lookup("a"))
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name   string
		src    string
		msg    string
		line   int
		column int
	}{
		{name: "empty document", src: ``, msg: "expected a value, found end of document", line: 1, column: 1},
		{name: "trailing content", src: `{} {}`, msg: "expected end of document", line: 1, column: 4},
		{name: "unterminated object", src: `{"a": 1`, msg: `expected "," or "}"`, line: 1, column: 8},
		{name: "unterminated array", src: `[1`, msg: `expected "," or "]"`, line: 1, column: 3},
		{name: "missing colon", src: `{"a" 1}`, msg: `expected ":"`, line: 1, column: 6},
		{name: "missing value", src: `{"a": }`, msg: "expected a value", line: 1, column: 7},
		{name: "comma only object", src: `{,}`, msg: "expected an object key", line: 1, column: 2},
		{name: "double comma in array", src: `[1,,]`, msg: "expected a value", line: 1, column: 4},
		{name: "numeric key", src: `{1: 2}`, msg: "expected an object key", line: 1, column: 2},
		{name: "bare identifier", src: `{"a": undefined}`, msg: "expected a value", line: 1, column: 7},
		{name: "unterminated string", src: `{"a": "b}`, msg: "unterminated string", line: 1, column: 7},
		{name: "line break in string", src: "\"a\nb\"", msg: "unescaped line break in string", line: 1, column: 3},
		{name: "invalid escape", src: `"\9"`, msg: "invalid escape sequence", line: 1, column: 2},
		{name: "octal-looking escape", src: `"\01"`, msg: `"\0" must not be followed by a digit`, line: 1, column: 2},
		{name: "short unicode escape", src: `"\u00"`, msg: "invalid escape sequence", line: 1, column: 2},
		{name: "unterminated block comment", src: "{\n/* open", msg: "unterminated block comment", line: 2, column: 1},
		{name: "stray slash", src: `{"a": 1} /`, msg: `unexpected "/"`, line: 1, column: 10},
		{name: "leading zero", src: `01`, msg: "leading zero", line: 1, column: 1},
		{name: "lone decimal point", src: `.`, msg: "number has no digits", line: 1, column: 1},
		{name: "empty exponent", src: `1e`, msg: "exponent has no digits", line: 1, column: 1},
		{name: "empty hex literal", src: `0x`, msg: "hexadecimal literal has no digits", line: 1, column: 1},
		{name: "unexpected character", src: `@`, msg: `unexpected character '@'`, line: 1, column: 1},
		{name: "column counts runes", src: "{\n  \"héllo\" 1}", msg: `expected ":"`, line: 2, column: 11},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parse([]byte(test.src))
			require.Error(t, err)

			var syntaxErr *SyntaxError
			require.ErrorAs(t, err, &syntaxErr)
			require.Contains(t, syntaxErr.Msg, test.msg)
			require.Equal(t, test.line, syntaxErr.Line, "line")
			require.Equal(t, test.column, syntaxErr.Column, "column")
			require.Contains(t, err.Error(), fmt.Sprintf("line %d, column %d", test.line, test.column))
		})
	}
}

func TestParseRejectsRunawayNesting(t *testing.T) {
	deepest := strings.Repeat("[", maxDepth) + strings.Repeat("]", maxDepth)
	_, err := parse([]byte(deepest))
	require.NoError(t, err)

	tooDeep := strings.Repeat("[", maxDepth+1) + strings.Repeat("]", maxDepth+1)
	_, err = parse([]byte(tooDeep))
	require.ErrorContains(t, err, "nested more than")
}

// parse runs the parser under test. It is the single place this suite binds to
// a Parser implementation.
func parse(src []byte) (*Document, error) {
	return NewJSON5Parser().Parse(src)
}

func plain(node *Node) any {
	switch node.Kind {
	case KindNull:
		return nil
	case KindBool:
		return node.Bool
	case KindNumber:
		return node.Number
	case KindString:
		return node.String
	case KindArray:
		values := make([]any, 0, len(node.Elements))
		for _, element := range node.Elements {
			values = append(values, plain(element))
		}
		return values
	case KindObject:
		values := make(map[string]any, len(node.Members))
		for _, member := range node.Members {
			values[member.Key] = plain(member.Value)
		}
		return values
	default:
		panic(fmt.Sprintf("unknown kind %d", node.Kind))
	}
}

func walk(node *Node, visit func(*Node)) {
	visit(node)
	for _, element := range node.Elements {
		walk(element, visit)
	}
	for _, member := range node.Members {
		walk(member.Value, visit)
	}
}
