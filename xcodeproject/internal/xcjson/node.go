package xcjson

// Kind classifies the value a Node holds.
type Kind uint8

// The value kinds. Integers and floats share KindNumber because the format,
// like JSON, has a single number type.
const (
	KindNull Kind = iota
	KindBool
	KindNumber
	KindString
	KindArray
	KindObject
)

// Node is one value and the byte range it occupies in the document it was
// parsed from.
//
// Only the fields listed for the node's Kind carry meaning; the rest are zero.
type Node struct {
	Kind Kind

	// Start is the offset of the value's first byte, and End the offset one
	// past its last. The range bounds the value exactly: it excludes
	// surrounding whitespace, comments and any trailing comma, so replacing
	// these bytes rewrites the value and nothing else.
	Start int
	End   int

	// Bool is the value of a KindBool node.
	Bool bool
	// Number is the value of a KindNumber node. Hexadecimal literals,
	// Infinity and NaN all decode into it.
	Number float64
	// String is the value of a KindString node, unquoted and unescaped.
	String string
	// Elements are the values of a KindArray node, in source order.
	Elements []*Node
	// Members are the key/value pairs of a KindObject node, in source order.
	Members []Member
}

// Member is one key/value pair of an object.
type Member struct {
	// Key is the member name, unquoted and unescaped. A parser that accepts
	// unquoted keys yields the identifier itself.
	Key string

	// KeyStart and KeyEnd bound the key token, quotes included when it has
	// any.
	KeyStart int
	KeyEnd   int

	Value *Node
}

// Lookup returns the value of the member named key, or nil if n is not an
// object or has no such member. A key may legally repeat, in which case the
// last occurrence wins, as it does in an ECMAScript object literal.
func (n *Node) Lookup(key string) *Node {
	if n == nil || n.Kind != KindObject {
		return nil
	}

	for i := len(n.Members) - 1; i >= 0; i-- {
		if n.Members[i].Key == key {
			return n.Members[i].Value
		}
	}

	return nil
}
