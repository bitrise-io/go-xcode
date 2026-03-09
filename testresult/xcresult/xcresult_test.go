package xcresult

import (
	"reflect"
	"testing"
)

func Test_filterIllegalChars(t *testing.T) {
	// \b == /u0008 -> backspace
	content := []byte("test\b text")

	if !reflect.DeepEqual(filterIllegalChars(content), []byte("test text")) {
		t.Fatal("illegal character is not removed")
	}
}
