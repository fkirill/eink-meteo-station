package utils

import "time"

type TimeProvider interface {
	LocalNow() time.Time
	UtcNow() time.Time
}

type locationTimeProvider struct {
	location *time.Location
}

func (o *locationTimeProvider) LocalNow() time.Time {
	return time.Now().In(o.location)
}

func (o *locationTimeProvider) UtcNow() time.Time {
	return time.Now().UTC()
}

type testTimeProvider struct {
	offset time.Duration
}

func (o *testTimeProvider) LocalNow() time.Time {
	return time.Now().UTC().Add(o.offset)
}

func (o *testTimeProvider) UtcNow() time.Time {
	return time.Now().UTC()
}

func NewTestTimeProvider(startTime time.Time) TimeProvider {
	now := time.Now()
	_, tzOffset := now.Zone()
	testOffset := int(startTime.Sub(now).Seconds())
	return &testTimeProvider{time.Duration(tzOffset+testOffset) * time.Second}
}

func NewTimeProvider() TimeProvider {
	return &locationTimeProvider{time.Local}
}
