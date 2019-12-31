package model

type Microservice struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type MicroserviceRepository interface {
	Create(Microservice) error
	Delete(ID int) error
	Update(Microservice) error
	FindOneByID(ID int) (Microservice, error)
}
