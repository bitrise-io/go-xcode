// Package serialized reads typed values from the maps that decoding plist and pbxproj files
// produces. A false ok means the key is absent or holds another type; Has, or the Optional
// accessors' WrongTypeError, tell the two apart.
package serialized

// Object is a decoded plist/pbxproj node: a set of string keys mapped to
// arbitrary values.
type Object map[string]any

// Get type-asserts rather than converts, so T must be the decoder's concrete type: string,
// bool, int64 or float64. Use Object for nested dictionaries (map[string]any) and StringSlice
// for arrays ([]any).
func Get[T any](o Object, key string) (value T, ok bool) {
	raw, exists := o[key]
	if !exists {
		return value, false
	}
	value, ok = raw.(T)
	return value, ok
}

// Keys returns the object's keys in unspecified order.
func (o Object) Keys() []string {
	keys := make([]string, 0, len(o))
	for key := range o {
		keys = append(keys, key)
	}
	return keys
}

// Has reports whether key is present, regardless of the stored value's type.
func (o Object) Has(key string) bool {
	_, ok := o[key]
	return ok
}

// Value returns the raw value stored at key. ok mirrors Go map semantics: it is
// false only when key is absent.
func (o Object) Value(key string) (value any, ok bool) {
	value, ok = o[key]
	return value, ok
}

func (o Object) Bool(key string) (bool, bool) {
	return Get[bool](o, key)
}

func (o Object) String(key string) (string, bool) {
	return Get[string](o, key)
}

func (o Object) Int64(key string) (int64, bool) {
	return Get[int64](o, key)
}

// StringSlice returns the []string stored at key. The value must be a []any
// whose every element is a string. ok is false if key is absent, the value is
// not a []any, or any element is not a string.
func (o Object) StringSlice(key string) ([]string, bool) {
	casted, ok := Get[[]any](o, key)
	if !ok {
		return nil, false
	}

	slice := make([]string, 0, len(casted))
	for _, v := range casted {
		item, ok := v.(string)
		if !ok {
			return nil, false
		}
		slice = append(slice, item)
	}

	return slice, true
}

// Object returns the nested Object stored at key. The value must be a
// map[string]any. ok is false if key is absent or the value is not a map.
func (o Object) Object(key string) (Object, bool) {
	casted, ok := Get[map[string]any](o, key)
	if !ok {
		return nil, false
	}
	return casted, true
}

// OptionalString returns the string stored at key. An absent key is "" with no
// error, as a plist omits what is not set; a value that is not a string is a
// WrongTypeError.
func (o Object) OptionalString(key string) (string, error) {
	if value, ok := o.String(key); ok {
		return value, nil
	}
	if raw, present := o.Value(key); present {
		return "", WrongTypeError{Key: key, Value: raw, Expected: "string"}
	}
	return "", nil
}

// OptionalStringSlice returns the []string stored at key. An absent key is nil
// with no error; a value that is not an array of strings is a WrongTypeError.
func (o Object) OptionalStringSlice(key string) ([]string, error) {
	if value, ok := o.StringSlice(key); ok {
		return value, nil
	}
	if raw, present := o.Value(key); present {
		return nil, WrongTypeError{Key: key, Value: raw, Expected: "string array"}
	}
	return nil, nil
}
