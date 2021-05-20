package xcodebuild

import (
	"reflect"
	"testing"

	"github.com/bitrise-io/go-xcode/xcodeproject/serialized"
)

func Test_parseShowBuildSettingsOutput(t *testing.T) {
	tests := []struct {
		name string
		out  string
		want serialized.Object
	}{
		{
			name: "empty output",
			out:  "",
			want: serialized.Object{},
		},
		{
			name: "simple output",
			out: `    ACTION = build
    AD_HOC_CODE_SIGNING_ALLOWED = NO
    ALTERNATE_GROUP = staff`,
			want: serialized.Object{"ACTION": "build", "AD_HOC_CODE_SIGNING_ALLOWED": "NO", "ALTERNATE_GROUP": "staff"},
		},
		{
			name: "output header",
			out: `Build settings for action build and target ios-simple-objc:
    ACTION = build
    AD_HOC_CODE_SIGNING_ALLOWED = NO`,
			want: serialized.Object{"ACTION": "build", "AD_HOC_CODE_SIGNING_ALLOWED": "NO"},
		},
		{
			name: "Build setting without value",
			out:  `    ACTION = `,
			want: serialized.Object{"ACTION": ""},
		},
		{
			name: "Build setting without =",
			out:  `    ACTION `,
			want: serialized.Object{},
		},
		{
			name: "Build setting without key",
			out:  `    = `,
			want: serialized.Object{},
		},
		{
			name: "Split the first = ",
			out:  `    ACTION = build+=test`,
			want: serialized.Object{"ACTION": "build+=test"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseShowBuildSettingsOutput(tt.out)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseShowBuildSettingsOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}
