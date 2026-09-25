package xcjson

import (
	"math"
	"strconv"
)

// maxDepth bounds how deeply containers may nest. Go grows a goroutine stack
// until it hits a limit it cannot recover from, so an untrusted file has to be
// stopped before the recursion is.
const maxDepth = 256

type parser struct {
	lex   lexer
	tok   token
	depth int
}

func (p *parser) parseDocument() (*Node, error) {
	if err := p.advance(); err != nil {
		return nil, err
	}

	root, err := p.parseValue()
	if err != nil {
		return nil, err
	}

	if p.tok.kind != tokenEOF {
		return nil, p.unexpected("end of document")
	}

	return root, nil
}

func (p *parser) parseValue() (*Node, error) {
	tok := p.tok

	switch tok.kind {
	case tokenString:
		return p.scalar(tok, &Node{Kind: KindString, String: tok.text})
	case tokenNumber:
		return p.scalar(tok, &Node{Kind: KindNumber, Number: tok.num})
	case tokenIdentifier:
		// Infinity and NaN reach the parser as identifiers unless they carry a
		// sign, because an unquoted object key may spell them too and only
		// position tells the two apart.
		switch tok.text {
		case "null":
			return p.scalar(tok, &Node{Kind: KindNull})
		case "true":
			return p.scalar(tok, &Node{Kind: KindBool, Bool: true})
		case "false":
			return p.scalar(tok, &Node{Kind: KindBool})
		case "Infinity":
			return p.scalar(tok, &Node{Kind: KindNumber, Number: math.Inf(1)})
		case "NaN":
			return p.scalar(tok, &Node{Kind: KindNumber, Number: math.NaN()})
		default:
			return nil, p.unexpected("a value")
		}
	case tokenObjectOpen:
		return p.parseObject()
	case tokenArrayOpen:
		return p.parseArray()
	default:
		return nil, p.unexpected("a value")
	}
}

func (p *parser) parseObject() (*Node, error) {
	node := &Node{Kind: KindObject, Start: p.tok.start}
	if err := p.enter(); err != nil {
		return nil, err
	}

	for p.tok.kind != tokenObjectClose {
		var member Member
		switch p.tok.kind {
		case tokenString, tokenIdentifier:
			member.Key = p.tok.text
			member.KeyStart, member.KeyEnd = p.tok.start, p.tok.end
		default:
			return nil, p.unexpected("an object key")
		}
		if err := p.advance(); err != nil {
			return nil, err
		}

		if p.tok.kind != tokenColon {
			return nil, p.unexpected(`":"`)
		}
		if err := p.advance(); err != nil {
			return nil, err
		}

		value, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		member.Value = value
		node.Members = append(node.Members, member)

		if err := p.separator(tokenObjectClose, `"," or "}"`); err != nil {
			return nil, err
		}
	}

	return p.leave(node)
}

func (p *parser) parseArray() (*Node, error) {
	node := &Node{Kind: KindArray, Start: p.tok.start}
	if err := p.enter(); err != nil {
		return nil, err
	}

	for p.tok.kind != tokenArrayClose {
		element, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		node.Elements = append(node.Elements, element)

		if err := p.separator(tokenArrayClose, `"," or "]"`); err != nil {
			return nil, err
		}
	}

	return p.leave(node)
}

func (p *parser) scalar(tok token, node *Node) (*Node, error) {
	node.Start, node.End = tok.start, tok.end
	return node, p.advance()
}

// enter consumes the opening bracket of a container and takes one step of the
// depth budget.
func (p *parser) enter() error {
	p.depth++
	if p.depth > maxDepth {
		return p.lex.errorf(p.tok.start, "nested more than %d levels deep", maxDepth)
	}
	return p.advance()
}

func (p *parser) leave(node *Node) (*Node, error) {
	node.End = p.tok.end
	p.depth--
	return node, p.advance()
}

// separator consumes the comma between two container entries. A trailing comma
// is legal in JSON5, so a comma may be followed by the closing bracket.
func (p *parser) separator(closing tokenKind, expected string) error {
	switch p.tok.kind {
	case tokenComma:
		return p.advance()
	case closing:
		return nil
	default:
		return p.unexpected(expected)
	}
}

func (p *parser) advance() error {
	tok, err := p.lex.next()
	if err != nil {
		return err
	}
	p.tok = tok
	return nil
}

func (p *parser) unexpected(expected string) error {
	return p.lex.errorf(p.tok.start, "expected %s, found %s", expected, p.describe(p.tok))
}

func (p *parser) describe(tok token) string {
	switch tok.kind {
	case tokenEOF:
		return "end of document"
	case tokenString:
		return "a string"
	case tokenNumber:
		return "a number"
	default:
		return strconv.Quote(string(p.lex.src[tok.start:tok.end]))
	}
}
