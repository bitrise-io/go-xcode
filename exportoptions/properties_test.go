package exportoptions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpgradeExportMethod(t *testing.T) {
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
			method: "developer-id",
			want:   "developer-id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, UpgradeExportMethod(tt.method))
		})
	}
}
