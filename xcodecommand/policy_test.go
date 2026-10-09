package xcodecommand

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

// One row per command: what its policy replaces and keeps both of.
// The constructors are wired to these policies; the merge rules themselves are covered in
// merge_test.go.
func TestPolicies(t *testing.T) {
	tests := []struct {
		policy     actionPolicy
		defaults   []string
		appendable []string
	}{
		{
			policy:     archivePolicy,
			defaults:   []string{"-destination"},
			appendable: []string{"-arch"},
		},
		{
			policy:     buildPolicy,
			defaults:   []string{"-destination"},
			appendable: []string{"-arch"},
		},
		{
			policy:     analyzePolicy,
			defaults:   []string{"-destination", "-resultBundlePath"},
			appendable: []string{"-arch"},
		},
		{
			// build-for-testing bakes the test selection into the xctestrun.
			policy:     buildForTestingPolicy,
			defaults:   []string{"-destination"},
			appendable: []string{"-arch"},
		},
		{
			policy: exportArchivePolicy,
		},
		{
			policy: resolvePackagesPolicy,
		},
	}
	for _, tt := range tests {
		t.Run(tt.policy.name, func(t *testing.T) {
			require.Equal(t, tt.defaults, slices.Sorted(slices.Values(tt.policy.defaults)))
			require.Equal(t, tt.appendable, slices.Sorted(slices.Values(tt.policy.appendable)))
		})
	}
}
