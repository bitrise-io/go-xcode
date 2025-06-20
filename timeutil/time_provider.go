package timeutil

import "time"

type TimeProvider interface {
	CurrentTime() time.Time
}

type DefaultTimeProvider struct{}

func NewDefaultTimeProvider() DefaultTimeProvider {
	return DefaultTimeProvider{}
}

func (d DefaultTimeProvider) CurrentTime() time.Time {
	return time.Now()
}
