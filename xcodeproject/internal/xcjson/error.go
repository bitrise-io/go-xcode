package xcjson

import (
	"fmt"
	"unicode/utf8"
)

// SyntaxError reports a malformed document together with the place the parser
// gave up. Project files are often hand-edited, so the location is part of the
// contract: callers surface it to the user rather than just saying the file is
// broken.
type SyntaxError struct {
	// Offset is the byte offset of the first byte the parser could not accept.
	Offset int
	// Line is 1-based.
	Line int
	// Column is 1-based and counts runes, not bytes.
	Column int
	// Msg describes what was wrong, without position information.
	Msg string
}

// Error implements the error interface.
func (e *SyntaxError) Error() string {
	return fmt.Sprintf("xcjson: %s at line %d, column %d", e.Msg, e.Line, e.Column)
}

func newSyntaxError(src []byte, offset int, format string, args ...any) *SyntaxError {
	line, column := position(src, offset)
	return &SyntaxError{
		Offset: offset,
		Line:   line,
		Column: column,
		Msg:    fmt.Sprintf(format, args...),
	}
}

func position(src []byte, offset int) (line, column int) {
	offset = min(offset, len(src))

	line, lineStart := 1, 0
	for i := 0; i < offset; {
		r, size := utf8.DecodeRune(src[i:])
		if !isLineTerminator(r) {
			i += size
			continue
		}
		if r == '\r' && i+size < len(src) && src[i+size] == '\n' {
			size++
		}
		i += size
		line++
		lineStart = i
	}

	return line, utf8.RuneCount(src[lineStart:offset]) + 1
}
