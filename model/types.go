package model

type ActorType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type TargetType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type TargetTypeRepository interface {
	Create(TargetType) error
	FirstByName(string) (TargetType, error)
}

type ActorTypeRepository interface {
	Create(ActorType) error
	FirstByName(string) (ActorType, error)
}
