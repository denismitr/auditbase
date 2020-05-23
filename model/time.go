package model

import (
	"fmt"
	"time"
)

const DefaultTimeFormat = time.RFC822

type JSONTime struct {
	time.Time
}

func (t JSONTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", t.Format(DefaultTimeFormat))
	return []byte(stamp), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// The time is expected to be a quoted string in RFC 3339 format.
func (t *JSONTime) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" {
		return nil
	}
	// Fractional seconds are handled implicitly by Parse.
	var err error
	tt, err := time.Parse(`"`+DefaultTimeFormat+`"`, string(data))
	*t = JSONTime{Time: tt}
	return err
}
