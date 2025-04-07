package exportoptions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpgradeToXcode15_3MethodNames(t *testing.T) {
	tests := []struct {
		name   string
		method Method
		want   Method
	}{
		{
			method: "app-store",
			want:   "app-store-connect",
		},
		{
			method: "ad-hoc",
			want:   "release-testing",
		},
		{
			method: "development",
			want:   "debugging",
		},
		{
			method: "enterprise",
			want:   "enterprise",
		},
		{
			method: "developer-id",
			want:   "developer-id",
		},
		{
			method: "package",
			want:   "package",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, UpgradeToXcode15_3MethodNames(tt.method))
		})
	}
}
