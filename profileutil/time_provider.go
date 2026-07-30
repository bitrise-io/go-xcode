package profileutil

import "github.com/bitrise-io/go-xcode/v2/timeutil"

// TimeProvider is an alias kept for backwards compatibility; the canonical declaration moved to
// timeutil so certificateutil can share it.
type TimeProvider = timeutil.TimeProvider

// DefaultTimeProvider is an alias kept for backwards compatibility; the canonical declaration moved
// to timeutil so certificateutil can share it.
type DefaultTimeProvider = timeutil.DefaultTimeProvider
