package xcjson

// NewJSON5Parser returns a Parser covering the whole JSON5 dialect: comments,
// trailing commas, unquoted object keys, single-quoted strings, hexadecimal
// numbers, Infinity and NaN. That is everything Xcode itself accepts on read,
// so a project file someone has edited by hand parses even though Xcode's own
// writer uses only part of the dialect.
func NewJSON5Parser() Parser {
	return json5Parser{}
}

// json5Parser reads with the hand-written scanner and recursive-descent parser
// in this package.
type json5Parser struct{}

// Parse implements Parser.
func (json5Parser) Parse(src []byte) (*Document, error) {
	parser := parser{lex: lexer{src: src}}

	root, err := parser.parseDocument()
	if err != nil {
		return nil, err
	}

	return &Document{Source: src, Root: root}, nil
}
