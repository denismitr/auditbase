package model

import (
	"fmt"
	"github.com/denismitr/auditbase/utils/errbag"
	"github.com/denismitr/auditbase/utils/validator"
)

type Microservice struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt,omitempty"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
}

type MicroserviceRepository interface {
	Create(*Microservice) (*Microservice, error)
	Delete(ID) error
	Update(ID, *Microservice) error
	FirstByID(ID ID) (*Microservice, error)
	FirstByName(name string) (*Microservice, error)
	FirstOrCreateByName(name string) (*Microservice, error)
	SelectAll() ([]*Microservice, error)
}

func (m *Microservice) Validate() *errbag.ErrorBag {
	eb := errbag.New()

	if !validator.IsUUID4(m.ID) {
		eb.Add("ID", ErrInvalidUUID4)
	}

	if validator.IsEmptyString(m.Name) {
		eb.Add("name", ErrNameIsRequired)
	}

	if validator.StringLenGt(m.Name, 36) {
		eb.Add("name", ErrMicroserviceNameTooLong)
	}

	if validator.StringLenGt(m.Description, 255) {
		eb.Add("description", ErrMicroserviceDescriptionTooLong)
	}

	return eb
}

func MicroserviceItemCacheKey(name string) string {
	return fmt.Sprintf("microservice_name_%s", name)
}
