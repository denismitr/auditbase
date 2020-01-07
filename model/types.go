package model

type ActorType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt,omitempty"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
}

type TargetType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt,omitempty"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
}

type TargetTypeRepository interface {
	Create(TargetType) error
	FirstByName(string) (TargetType, error)
}

type ActorTypeRepository interface {
	Create(ActorType) error
	FirstByName(string) (ActorType, error)
}
