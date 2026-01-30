package profileutil

import "time"

type TimeProvider interface {
	Now() time.Time
}

type DefaultTimeProvider struct{}

func (DefaultTimeProvider) Now() time.Time {
	return time.Now()
}
