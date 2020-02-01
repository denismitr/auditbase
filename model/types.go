package model

type ActorType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt,omitempty"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
}

// TargetType - is an entity that is a subject of some action
// e.g. subscription, account, order that can be creatd, modified, updated
// by some actor represented by ActorType
type TargetType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt,omitempty"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
}

// TargetTypeRepository - abstracts away the methods for woring with
// target types
type TargetTypeRepository interface {
	Select() ([]TargetType, error)
	Create(TargetType) error
	FirstByName(string) (TargetType, error)
	FirstByID(string) (TargetType, error)
	FirstOrCreateByName(string) (TargetType, error)
}

// ActorTypeRepository governs actor types schema
type ActorTypeRepository interface {
	Select() ([]ActorType, error)
	Create(ActorType) error
	FirstByName(string) (ActorType, error)
	FirstByID(string) (ActorType, error)
	FirstOrCreateByName(string) (ActorType, error)
}
