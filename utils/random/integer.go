package random

import (
	"math/rand"
	"sync"
	"time"
)

var intMu sync.Mutex
var seededInt = rand.New(rand.NewSource(time.Now().UnixNano()))

func Int(min, max int) int {
	intMu.Lock()
	defer intMu.Unlock()
	return seededInt.Intn(max - min + 1) + min
}

func Timestamp(min time.Time, max time.Time) int64 {
	intMu.Lock()
	defer intMu.Unlock()

	ts := seededInt.Intn(int(max.Unix()) - int(min.Unix()) + 1) + int(min.Unix())
	return int64(ts)
}

func Time(min time.Time, max time.Time) time.Time {
	ts := Timestamp(min, max)
	return time.Unix(ts, 0)
}
