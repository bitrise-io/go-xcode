package xcodecommand

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAnalyze(t *testing.T) {
	cmd, err := Analyze(AnalyzeParams{ProjectPath: "App.xcodeproj", Scheme: "App", ResultBundlePath: "/tmp/Analyze.xcresult", DisableCodeSigning: true, AdditionalOptions: additional("COMPILER_INDEX_STORE_ENABLE=NO")})
	require.NoError(t, err)
	require.Equal(t, []string{"analyze", "-project", "App.xcodeproj", "-scheme", "App", "-resultBundlePath", "/tmp/Analyze.xcresult", "CODE_SIGNING_ALLOWED=NO", "COMPILER_INDEX_STORE_ENABLE=NO"}, cmd.Args())
}
