package clock

import "time"

type Clock interface {
	CurrentTimestamp() int64
	CurrentTime() time.Time
}

type StdClock struct{}

func New() *StdClock {
	return &StdClock{}
}

func (c *StdClock) CurrentTimestamp() int64 {
	return time.Now().Unix()
}

func (c *StdClock) CurrentTime() time.Time {
	return time.Now()
}
