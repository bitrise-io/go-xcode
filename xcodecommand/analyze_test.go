package xcodecommand

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAnalyze(t *testing.T) {
	tests := []struct {
		name   string
		params AnalyzeParams
		want   []string
	}{
		{
			name:   "steps-xcode-analyze: result bundle, no code signing, additional options last",
			params: AnalyzeParams{ProjectPath: "App.xcodeproj", Scheme: "App", ResultBundlePath: "/tmp/Analyze.xcresult", DisableCodeSigning: true, AdditionalOptions: additional("COMPILER_INDEX_STORE_ENABLE=NO")},
			want:   []string{"analyze", "-project", "App.xcodeproj", "-scheme", "App", "-resultBundlePath", "/tmp/Analyze.xcresult", "CODE_SIGNING_ALLOWED=NO", "COMPILER_INDEX_STORE_ENABLE=NO"},
		},
		{
			// The step's -destination and -resultBundlePath are defaults: the user's replace them.
			name:   "a user -destination and -resultBundlePath replace the step's defaults",
			params: AnalyzeParams{Scheme: "App", Destination: "generic/platform=iOS", ResultBundlePath: "/tmp/Analyze.xcresult", AdditionalOptions: additional("-destination", "generic/platform=tvOS", "-resultBundlePath", "/tmp/mine.xcresult")},
			want:   []string{"analyze", "-scheme", "App", "-destination", "generic/platform=tvOS", "-resultBundlePath", "/tmp/mine.xcresult"},
		},
		{
			// "-resultBundlePath -quiet": the parser takes -quiet as the value, as xcodebuild does.
			// The step's default stays, and xcodebuild refuses the repeat instead of writing the
			// bundle to ./-quiet.
			name:   "a -resultBundlePath that took the next flag as its value replaces nothing",
			params: AnalyzeParams{Scheme: "App", ResultBundlePath: "/tmp/Analyze.xcresult", AdditionalOptions: additional("-resultBundlePath", "-quiet")},
			want:   []string{"analyze", "-scheme", "App", "-resultBundlePath", "/tmp/Analyze.xcresult", "-resultBundlePath", "-quiet"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := Analyze(tt.params)
			require.NoError(t, err)
			require.Equal(t, tt.want, cmd.Args())
		})
	}
}
