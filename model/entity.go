package model

// Entity - represents something that can act on data
// or something that can be a subject to change or both
type Entity struct {
	ID          string       `json:"id"`
	Service     Microservice `json:"service"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	CreatedAt   string       `json:"createdAt,omitempty"`
	UpdatedAt   string       `json:"updatedAt,omitempty"`
}

// EntityRepository governs entities data interactions
type EntityRepository interface {
	Select() ([]*Entity, error)
	Create(*Entity) error
	FirstByNameAndService(string, *Microservice) (*Entity, error)
	FirstByID(string) (*Entity, error)
	FirstOrCreateByNameAndService(string, *Microservice) (*Entity, error)
}
