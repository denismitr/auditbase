package model

type Microservice struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type MicroserviceRepository interface {
	Create(Microservice) error
	Delete(ID string) error
	Update(ID string, m Microservice) error
	GetOneByID(ID string) (Microservice, error)
	SelectAll() ([]Microservice, error)
}
