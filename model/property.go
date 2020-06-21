package model

import "time"

type Property struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	EntityID    string     `json:"entityId"`
	ChangeCount int        `json:"changeCount"`
	Changes     []Change   `json:"changes"`
	LastEventAt *time.Time `json:"lastEventAt,omitempty"`
}

type PropertyStat struct {
	Name       string `json:"name"`
	EventCount int    `json:"eventCount"`
}

type PropertyRepository interface {
	GetIDOrCreate(name, entityID string) (string, error)
	FirstByID(ID string) (*Property, error)
	Select(*Filter, *Sort, *Pagination) ([]*Property, *Meta, error)
}
