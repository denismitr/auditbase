package utils

import "time"

type Clock interface {
	CurrentTimestamp() int64
}

type StdClock struct{}

func NewClock() *StdClock {
	return &StdClock{}
}

func (c *StdClock) CurrentTimestamp() int64 {
	return time.Now().Unix()
}
