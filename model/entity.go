package model

import "fmt"

// Entity - represents something that can act on data
// or something that can be a subject to change or both
type Entity struct {
	ID          string        `json:"id"`
	ServiceID   string        `json:"serviceId"`
	Service     *Microservice `json:"service,omitempty"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	CreatedAt   string        `json:"createdAt,omitempty"`
	UpdatedAt   string        `json:"updatedAt,omitempty"`
}

// EntityRepository governs entities data interactions
type EntityRepository interface {
	Select(*Filter, *Sort, *Pagination) ([]*Entity, error)
	Create(*Entity) error
	FirstByNameAndService(string, *Microservice) (*Entity, error)
	FirstByID(string) (*Entity, error)
	//Properties(string) ([]Property, error)
	FirstOrCreateByNameAndService(string, *Microservice) (*Entity, error)
}

func EntityItemCacheKey(name string, microservice *Microservice) string {
	return fmt.Sprintf("entity_%s_%s", name, microservice.ID)
}
