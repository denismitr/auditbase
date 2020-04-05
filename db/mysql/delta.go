package mysql

import (
	"encoding/json"
	"errors"
)

type delta map[string][]interface{}

func (d *delta) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed while converting to delta")
	}

	return json.Unmarshal(b, &d)
}
