package xcodeproj

import (
	"testing"

	plist "github.com/bitrise-io/go-plist"
	"github.com/bitrise-io/go-utils/pretty"
	"github.com/bitrise-io/go-xcode/xcodeproject/serialized"
	"github.com/stretchr/testify/require"
)

func Test_parseProductReference(t *testing.T) {
	t.Log("PBXFileReference")
	{
		var raw serialized.Object
		_, err := plist.Unmarshal([]byte(rawProductReference), &raw)
		require.NoError(t, err)

		productReference, err := parseProductReference("13E76E0E1F4AC90A0028096E", raw)
		require.NoError(t, err)
		// fmt.Printf("productReference:\n%s\n", pretty.Object(productReference))
		require.Equal(t, expectedProductReference, pretty.Object(productReference))
	}
}

const rawProductReference = `{
	13E76E0E1F4AC90A0028096E /* code-sign-test.app */ = {isa = PBXFileReference; explicitFileType = wrapper.application; includeInIndex = 0; path = "code-sign-test.app"; sourceTree = BUILT_PRODUCTS_DIR; };
}`

const expectedProductReference = `{
	"Path": "code-sign-test.app"
}`
