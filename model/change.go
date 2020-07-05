package model

import (
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

type DataType int

const (
	UnknownDataType DataType = iota
	NullDataType
	StringDataType
	IntegerDataType
	FloatDataType
)

var dataTypeConstants = map[DataType]string{
	UnknownDataType: "unknown",
	NullDataType:    "null",
	StringDataType:  "string",
	IntegerDataType: "integer",
	FloatDataType:   "float",
}

func (dt DataType) String() string {
	return dataTypeConstants[dt]
}

func (dt DataType) MarshalJSON() ([]byte, error) {
	if dt == NullDataType || dt == UnknownDataType {
		return nil, nil
	}

	return []byte(fmt.Sprintf("\"%s\"", dt.String())), nil
}

func (dt *DataType) UnmarshalJSON(data []byte) error {
	s, _ := strconv.Unquote(string(data))

	for k, v := range dataTypeConstants {
		if v == s {
			*dt = k
			return nil
		}
	}

	return errors.Errorf("%s is not a valid data type", s)
}

type Change struct {
	ID              string    `json:"id"`
	EventID         string    `json:"eventId"`
	PropertyID      string    `json:"propertyId"`
	CurrentDataType DataType  `json:"currentDataType"`
	From            *string   `json:"from"`
	To              *string   `json:"to"`
	Property        *Property `json:"property,omitempty"`
	CreatedAt       time.Time `json:"createdAt,omitempty"`
}

type PropertyChange struct {
	ID              string   `json:"id"`
	EventID         string   `json:"eventId"`
	PropertyID      string   `json:"propertyId"`
	From            *string  `json:"from"`
	To              *string  `json:"to"`
	CurrentDataType DataType `json:"currentDataType"`
	PropertyName    string   `json:"property,omitempty"`
	EntityID        string   `json:"entityId,omitempty"`
}

type ChangeRepository interface {
	Select(*Filter, *Sort, *Pagination) ([]*Change, *Meta, error)
	FirstByID(string) (*Change, error)
}
