package serialized

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeys(t *testing.T) {
	o := Object{"key": "value", "key1": "value", "key2": "value"}
	keys := o.Keys()
	require.Equal(t, 3, len(keys))
	require.Contains(t, keys, "key")
	require.Contains(t, keys, "key1")
	require.Contains(t, keys, "key2")
}

func TestHas(t *testing.T) {
	o := Object{"present": "value", "wrongType": 42}

	require.True(t, o.Has("present"))
	require.True(t, o.Has("wrongType"))
	require.False(t, o.Has("missing"))
}

func TestValue(t *testing.T) {
	o := Object{"key": "value"}

	{
		v, ok := o.Value("key")
		require.True(t, ok)
		require.Equal(t, "value", v)
	}

	{
		v, ok := o.Value("not_existing_key")
		require.False(t, ok)
		require.Nil(t, v)
	}
}

func TestBool(t *testing.T) {
	o := Object{"key": true, "wrong": "not_a_bool"}

	{
		v, ok := o.Bool("key")
		require.True(t, ok)
		require.Equal(t, true, v)
	}

	{
		v, ok := o.Bool("missing")
		require.False(t, ok)
		require.Equal(t, false, v)
	}

	{
		v, ok := o.Bool("wrong")
		require.False(t, ok)
		require.Equal(t, false, v)
	}
}

func TestString(t *testing.T) {
	o := Object{"key": "value", "wrong": 42}

	{
		v, ok := o.String("key")
		require.True(t, ok)
		require.Equal(t, "value", v)
	}

	{
		v, ok := o.String("missing")
		require.False(t, ok)
		require.Equal(t, "", v)
	}

	{
		v, ok := o.String("wrong")
		require.False(t, ok)
		require.Equal(t, "", v)
	}
}

func TestInt64(t *testing.T) {
	o := Object{"key": int64(42), "wrong": "not_an_int"}

	{
		v, ok := o.Int64("key")
		require.True(t, ok)
		require.Equal(t, int64(42), v)
	}

	{
		v, ok := o.Int64("missing")
		require.False(t, ok)
		require.Equal(t, int64(0), v)
	}

	{
		v, ok := o.Int64("wrong")
		require.False(t, ok)
		require.Equal(t, int64(0), v)
	}
}

func TestObject(t *testing.T) {
	o := Object{"key": map[string]any{"object_key": "object_value"}, "wrong": "not_an_object"}

	{
		v, ok := o.Object("key")
		require.True(t, ok)
		require.Equal(t, Object{"object_key": "object_value"}, v)
	}

	{
		v, ok := o.Object("missing")
		require.False(t, ok)
		require.Nil(t, v)
	}

	{
		v, ok := o.Object("wrong")
		require.False(t, ok)
		require.Nil(t, v)
	}
}

func TestStringSlice(t *testing.T) {
	o := Object{
		"buildConfigurations": []any{"13E76E3B1F4AC90A0028096E", "13E76E3C1F4AC90A0028096E"},
		"notASlice":           "value",
		"mixedElements":       []any{"a", 42},
	}

	{
		v, ok := o.StringSlice("buildConfigurations")
		require.True(t, ok)
		require.Equal(t, []string{"13E76E3B1F4AC90A0028096E", "13E76E3C1F4AC90A0028096E"}, v)
	}

	{
		v, ok := o.StringSlice("missing")
		require.False(t, ok)
		require.Nil(t, v)
	}

	{
		v, ok := o.StringSlice("notASlice")
		require.False(t, ok)
		require.Nil(t, v)
	}

	{
		v, ok := o.StringSlice("mixedElements")
		require.False(t, ok)
		require.Nil(t, v)
	}
}

func TestGet(t *testing.T) {
	o := Object{"str": "value", "num": int64(7)}

	{
		v, ok := Get[string](o, "str")
		require.True(t, ok)
		require.Equal(t, "value", v)
	}

	{
		v, ok := Get[int64](o, "num")
		require.True(t, ok)
		require.Equal(t, int64(7), v)
	}

	{
		v, ok := Get[string](o, "num")
		require.False(t, ok)
		require.Equal(t, "", v)
	}

	{
		v, ok := Get[string](o, "missing")
		require.False(t, ok)
		require.Equal(t, "", v)
	}
}

// Get asserts, it does not convert: it only succeeds for the exact concrete
// type the decoder produced (scalars). Composite values need the dedicated
// accessors, which is what these cases guard against a naive "just use Get"
// refactor of Object/StringSlice.
func TestGet_compositesRequireDedicatedAccessors(t *testing.T) {
	o := Object{
		"dict":  map[string]any{"k": "v"}, // decoded nested dictionary
		"array": []any{"a", "b"},          // decoded array
	}

	// Nested dicts are map[string]any, not the named Object type.
	_, ok := Get[Object](o, "dict")
	require.False(t, ok, "Get[Object] must not match a stored map[string]any")

	obj, ok := o.Object("dict") // Object() converts
	require.True(t, ok)
	require.Equal(t, Object{"k": "v"}, obj)

	// Arrays are []any, not []string.
	_, ok = Get[[]string](o, "array")
	require.False(t, ok, "Get[[]string] must not match a stored []any")

	slice, ok := o.StringSlice("array") // StringSlice() casts element-wise
	require.True(t, ok)
	require.Equal(t, []string{"a", "b"}, slice)
}
