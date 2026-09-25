package xcjson

import (
	"math"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

const (
	zeroWidthNonJoiner = '\u200C'
	zeroWidthJoiner    = '\u200D'
	byteOrderMark      = '\uFEFF'

	// unescapeBufferSize is the starting size of the buffer a string or
	// identifier is unescaped into. Names and paths in a project file are
	// short, so this is sized to hold most of them without growing.
	unescapeBufferSize = 64
)

type tokenKind uint8

const (
	tokenEOF tokenKind = iota
	tokenObjectOpen
	tokenObjectClose
	tokenArrayOpen
	tokenArrayClose
	tokenColon
	tokenComma
	tokenString
	tokenNumber
	tokenIdentifier
)

type token struct {
	kind  tokenKind
	start int
	end   int

	text string  // tokenString: the unescaped contents; tokenIdentifier: the name
	num  float64 // tokenNumber
}

type lexer struct {
	src []byte
	pos int
}

// next skips whitespace and comments and returns the token that follows.
func (l *lexer) next() (token, error) {
	if err := l.skipIgnorable(); err != nil {
		return token{}, err
	}

	if l.pos >= len(l.src) {
		return token{kind: tokenEOF, start: l.pos, end: l.pos}, nil
	}

	switch l.src[l.pos] {
	case '{':
		return l.punctuation(tokenObjectOpen), nil
	case '}':
		return l.punctuation(tokenObjectClose), nil
	case '[':
		return l.punctuation(tokenArrayOpen), nil
	case ']':
		return l.punctuation(tokenArrayClose), nil
	case ':':
		return l.punctuation(tokenColon), nil
	case ',':
		return l.punctuation(tokenComma), nil
	case '"', '\'':
		return l.scanString()
	case '+', '-', '.', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return l.scanNumber()
	default:
		return l.scanIdentifier()
	}
}

func (l *lexer) errorf(offset int, format string, args ...any) error {
	return newSyntaxError(l.src, offset, format, args...)
}

func (l *lexer) punctuation(kind tokenKind) token {
	start := l.pos
	l.pos++
	return token{kind: kind, start: start, end: l.pos}
}

func (l *lexer) skipIgnorable() error {
	for l.pos < len(l.src) {
		c := l.src[l.pos]
		switch {
		case c == '/':
			if err := l.skipComment(); err != nil {
				return err
			}
		case c < utf8.RuneSelf:
			if !isASCIISpace(c) {
				return nil
			}
			l.pos++
		default:
			r, size := utf8.DecodeRune(l.src[l.pos:])
			if !isSpace(r) {
				return nil
			}
			l.pos += size
		}
	}

	return nil
}

func (l *lexer) skipComment() error {
	start := l.pos
	if l.pos+1 >= len(l.src) {
		return l.errorf(start, `unexpected "/"`)
	}

	switch l.src[l.pos+1] {
	case '/':
		l.pos += 2
		for l.pos < len(l.src) {
			r, size := utf8.DecodeRune(l.src[l.pos:])
			if isLineTerminator(r) {
				return nil
			}
			l.pos += size
		}
		return nil
	case '*':
		l.pos += 2
		for l.pos+1 < len(l.src) {
			if l.src[l.pos] == '*' && l.src[l.pos+1] == '/' {
				l.pos += 2
				return nil
			}
			l.pos++
		}
		return l.errorf(start, "unterminated block comment")
	default:
		return l.errorf(start, `unexpected "/"`)
	}
}

func (l *lexer) scanString() (token, error) {
	start := l.pos
	quote := l.src[l.pos]
	l.pos++

	// decoded stays nil while the string needs no unescaping, which lets the
	// common case return a slice of the source instead of a copy.
	var decoded []byte
	literal := l.pos

	for {
		if l.pos >= len(l.src) {
			return token{}, l.errorf(start, "unterminated string")
		}

		switch c := l.src[l.pos]; {
		case c == quote:
			text := string(l.src[literal:l.pos])
			if decoded != nil {
				text = string(append(decoded, l.src[literal:l.pos]...))
			}
			l.pos++
			return token{kind: tokenString, start: start, end: l.pos, text: text}, nil

		case c == '\\':
			if decoded == nil {
				decoded = make([]byte, 0, unescapeBufferSize)
			}
			decoded = append(decoded, l.src[literal:l.pos]...)
			l.pos++
			var err error
			if decoded, err = l.scanEscape(decoded); err != nil {
				return token{}, err
			}
			literal = l.pos

		case c == '\n' || c == '\r':
			return token{}, l.errorf(l.pos, "unescaped line break in string")

		case c < utf8.RuneSelf:
			l.pos++

		default:
			_, size := utf8.DecodeRune(l.src[l.pos:])
			l.pos += size
		}
	}
}

// scanEscape consumes the escape sequence whose backslash has already been
// read and appends what it stands for to decoded.
func (l *lexer) scanEscape(decoded []byte) ([]byte, error) {
	at := l.pos - 1
	if l.pos >= len(l.src) {
		return nil, l.errorf(at, "unterminated escape sequence")
	}

	switch c := l.src[l.pos]; c {
	case 'b':
		l.pos++
		return append(decoded, '\b'), nil
	case 'f':
		l.pos++
		return append(decoded, '\f'), nil
	case 'n':
		l.pos++
		return append(decoded, '\n'), nil
	case 'r':
		l.pos++
		return append(decoded, '\r'), nil
	case 't':
		l.pos++
		return append(decoded, '\t'), nil
	case 'v':
		l.pos++
		return append(decoded, '\v'), nil

	case '0':
		l.pos++
		if l.pos < len(l.src) && isDigit(l.src[l.pos]) {
			return nil, l.errorf(at, `"\0" must not be followed by a digit`)
		}
		return append(decoded, 0), nil

	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return nil, l.errorf(at, "invalid escape sequence")

	case 'x':
		l.pos++
		value, err := l.scanHexDigits(2, at)
		if err != nil {
			return nil, err
		}
		return utf8.AppendRune(decoded, rune(value)), nil

	case 'u':
		l.pos++
		return l.scanUnicodeEscape(decoded, at)

	case '\r':
		// A backslash before a line break is a line continuation: it stands
		// for nothing at all. CRLF is one break, not two.
		l.pos++
		if l.pos < len(l.src) && l.src[l.pos] == '\n' {
			l.pos++
		}
		return decoded, nil

	case '\n':
		l.pos++
		return decoded, nil

	default:
		r, size := utf8.DecodeRune(l.src[l.pos:])
		l.pos += size
		if isLineTerminator(r) {
			return decoded, nil
		}
		return utf8.AppendRune(decoded, r), nil
	}
}

func (l *lexer) scanUnicodeEscape(decoded []byte, at int) ([]byte, error) {
	value, err := l.scanHexDigits(4, at)
	if err != nil {
		return nil, err
	}

	r := rune(value)
	if !utf16.IsSurrogate(r) {
		return utf8.AppendRune(decoded, r), nil
	}

	// A code point above the BMP is written as a surrogate pair, so a leading
	// surrogate is only meaningful together with the escape that follows it.
	if l.pos+1 < len(l.src) && l.src[l.pos] == '\\' && l.src[l.pos+1] == 'u' {
		resume := l.pos
		l.pos += 2
		low, err := l.scanHexDigits(4, at)
		if err != nil {
			return nil, err
		}
		if paired := utf16.DecodeRune(r, rune(low)); paired != utf8.RuneError {
			return utf8.AppendRune(decoded, paired), nil
		}
		l.pos = resume
	}

	return utf8.AppendRune(decoded, utf8.RuneError), nil
}

func (l *lexer) scanHexDigits(count, at int) (uint32, error) {
	if l.pos+count > len(l.src) {
		return 0, l.errorf(at, "invalid escape sequence")
	}

	var value uint32
	for range count {
		digit := hexDigit(l.src[l.pos])
		if digit < 0 {
			return 0, l.errorf(at, "invalid escape sequence")
		}
		value = value<<4 | uint32(digit)
		l.pos++
	}

	return value, nil
}

func (l *lexer) scanNumber() (token, error) {
	start := l.pos
	if c := l.src[l.pos]; c == '+' || c == '-' {
		l.pos++
	}

	switch {
	case l.hasPrefix("Infinity"):
		l.pos += len("Infinity")
	case l.hasPrefix("NaN"):
		l.pos += len("NaN")
	case l.hasPrefix("0x"), l.hasPrefix("0X"):
		l.pos += 2
		if l.skipDigitsFunc(isHexDigit) == 0 {
			return token{}, l.errorf(start, "hexadecimal literal has no digits")
		}
	default:
		if err := l.skipDecimal(start); err != nil {
			return token{}, err
		}
	}

	num, err := parseNumber(string(l.src[start:l.pos]))
	if err != nil {
		return token{}, l.errorf(start, "%s", err)
	}

	return token{kind: tokenNumber, start: start, end: l.pos, num: num}, nil
}

func (l *lexer) skipDecimal(start int) error {
	intStart := l.pos
	intDigits := l.skipDigitsFunc(isDigit)
	if intDigits > 1 && l.src[intStart] == '0' {
		return l.errorf(start, "number must not have a leading zero")
	}

	var fracDigits int
	if l.pos < len(l.src) && l.src[l.pos] == '.' {
		l.pos++
		fracDigits = l.skipDigitsFunc(isDigit)
	}
	if intDigits == 0 && fracDigits == 0 {
		return l.errorf(start, "number has no digits")
	}

	if l.pos < len(l.src) && (l.src[l.pos] == 'e' || l.src[l.pos] == 'E') {
		l.pos++
		if l.pos < len(l.src) && (l.src[l.pos] == '+' || l.src[l.pos] == '-') {
			l.pos++
		}
		if l.skipDigitsFunc(isDigit) == 0 {
			return l.errorf(start, "exponent has no digits")
		}
	}

	return nil
}

func (l *lexer) skipDigitsFunc(isDigit func(byte) bool) int {
	count := 0
	for l.pos < len(l.src) && isDigit(l.src[l.pos]) {
		l.pos++
		count++
	}
	return count
}

func (l *lexer) hasPrefix(prefix string) bool {
	return l.pos+len(prefix) <= len(l.src) && string(l.src[l.pos:l.pos+len(prefix)]) == prefix
}

func (l *lexer) scanIdentifier() (token, error) {
	start := l.pos

	var decoded []byte
	literal := start

	for l.pos < len(l.src) {
		if l.src[l.pos] == '\\' {
			if decoded == nil {
				decoded = make([]byte, 0, unescapeBufferSize)
			}
			decoded = append(decoded, l.src[literal:l.pos]...)

			at := l.pos
			l.pos++
			if l.pos >= len(l.src) || l.src[l.pos] != 'u' {
				return token{}, l.errorf(at, "invalid escape sequence")
			}
			l.pos++
			value, err := l.scanHexDigits(4, at)
			if err != nil {
				return token{}, err
			}
			if !isIdentifierRune(rune(value), at == start) {
				return token{}, l.errorf(at, "invalid character in identifier")
			}
			decoded = utf8.AppendRune(decoded, rune(value))
			literal = l.pos
			continue
		}

		r, size := utf8.DecodeRune(l.src[l.pos:])
		if !isIdentifierRune(r, l.pos == start) {
			break
		}
		l.pos += size
	}

	if l.pos == start {
		r, _ := utf8.DecodeRune(l.src[l.pos:])
		return token{}, l.errorf(start, "unexpected character %q", r)
	}

	text := string(l.src[literal:l.pos])
	if decoded != nil {
		text = string(append(decoded, l.src[literal:l.pos]...))
	}

	return token{kind: tokenIdentifier, start: start, end: l.pos, text: text}, nil
}

// parseNumber decodes a number literal the lexer has already checked the
// shape of.
func parseNumber(literal string) (float64, error) {
	negative := literal[0] == '-'
	unsigned := literal
	if negative || literal[0] == '+' {
		unsigned = literal[1:]
	}

	switch {
	case unsigned == "Infinity":
		if negative {
			return math.Inf(-1), nil
		}
		return math.Inf(1), nil

	case unsigned == "NaN":
		return math.NaN(), nil

	case strings.HasPrefix(unsigned, "0x"), strings.HasPrefix(unsigned, "0X"):
		// ParseFloat accepts a hexadecimal mantissa only alongside a binary
		// exponent, so supply the one that leaves the value unchanged.
		value, err := strconv.ParseFloat(unsigned+"p0", 64)
		if err != nil {
			return 0, err
		}
		if negative {
			value = -value
		}
		return value, nil

	default:
		return strconv.ParseFloat(literal, 64)
	}
}

func isASCIISpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\v' || c == '\f'
}

// isSpace reports whether r separates tokens. JSON5 takes its whitespace from
// ECMAScript, which is unicode.IsSpace plus the byte order mark.
func isSpace(r rune) bool {
	return unicode.IsSpace(r) || r == byteOrderMark
}

func isLineTerminator(r rune) bool {
	return r == '\n' || r == '\r' || r == '\u2028' || r == '\u2029'
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isHexDigit(c byte) bool {
	return hexDigit(c) >= 0
}

func hexDigit(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	default:
		return -1
	}
}

// isIdentifierRune reports whether r may appear in an unquoted object key,
// which JSON5 defines as an ECMAScript IdentifierName.
func isIdentifierRune(r rune, first bool) bool {
	if r == '$' || r == '_' || unicode.IsLetter(r) || unicode.Is(unicode.Nl, r) {
		return true
	}
	if first {
		return false
	}
	return r == zeroWidthNonJoiner || r == zeroWidthJoiner ||
		unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Mc, r) ||
		unicode.Is(unicode.Nd, r) || unicode.Is(unicode.Pc, r)
}
