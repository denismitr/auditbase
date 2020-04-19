package clock

import "time"

// TimestampToTime - converts timetamp in seconds to time.Time structur
func TimestampToTime(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}
