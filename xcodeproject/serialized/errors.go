package serialized

import "fmt"

// WrongTypeError reports a key whose value is not of the type the caller asked
// for. Value is the stored value, so the file it came from can be corrected.
type WrongTypeError struct {
	Key      string
	Value    any
	Expected string
}

func (e WrongTypeError) Error() string {
	return fmt.Sprintf("%s is not a %s: %#v (%T)", e.Key, e.Expected, e.Value, e.Value)
}
