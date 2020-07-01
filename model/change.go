package model

import "time"

type Change struct {
	ID              string    `json:"id"`
	EventID         string    `json:"eventId"`
	PropertyID      string    `json:"propertyId"`
	CurrentDataType *string   `json:"currentDataType"`
	From            *string   `json:"from"`
	To              *string   `json:"to"`
	Property        *Property `json:"property,omitempty"`
	CreatedAt       time.Time `json:"createdAt,omitempty"`
}

type PropertyChange struct {
	ID              string  `json:"id"`
	EventID         string  `json:"eventId"`
	PropertyID      string  `json:"propertyId"`
	From            *string `json:"from"`
	To              *string `json:"to"`
	CurrentDataType *string `json:"currentDataType"`
	PropertyName    string  `json:"property,omitempty"`
	EntityID        string  `json:"entityId,omitempty"`
}

type ChangeRepository interface {
	Select(*Filter, *Sort, *Pagination) ([]*Change, *Meta, error)
	FirstByID(string) (*Change, error)
}
