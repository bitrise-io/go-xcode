// Package xcjson reads the document format Xcode writes project.xcproj in, and
// records where in the file every value came from.
//
// Every Node carries the byte range it occupies in the source. That is what
// makes the package useful beyond decoding: a writer can replace one value by
// splicing over its range, which leaves comments, formatting and every key
// nobody has modelled byte for byte as Xcode left them. Comments are therefore
// not part of the value tree at all; they sit in the gaps between ranges and
// survive untouched.
package xcjson

// Parser reads a project document.
//
// Reading sits behind an interface because Xcode accepts a wider dialect than
// its own writer ever produces, so how much of that width to cover is a choice
// a parser makes. The model does not depend on that choice: Document, Node and
// SyntaxError describe any parse of any accepted dialect.
type Parser interface {
	// Parse reads one document. It returns a *SyntaxError, which carries the
	// line and column, when src is malformed.
	//
	// src is not copied, so neither the returned Document nor any Node stays
	// correct if src is modified afterwards.
	Parse(src []byte) (*Document, error)
}

// Document is a parsed project file: the bytes it was read from and the value
// tree that indexes into them.
type Document struct {
	// Source is the document as it was parsed. Node offsets are only valid
	// against these exact bytes.
	Source []byte

	// Root is the document's single top-level value. Xcode writes an object,
	// but the format permits any value and so does this package.
	Root *Node
}

// Raw returns node's bytes exactly as they appear in the document. The result
// aliases Source rather than copying it.
func (d *Document) Raw(node *Node) []byte {
	return d.Source[node.Start:node.End]
}
