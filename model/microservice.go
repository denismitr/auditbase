package model

type Microservice struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt,omitempty"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
}

type MicroserviceRepository interface {
	Create(Microservice) error
	Delete(ID string) error
	Update(ID string, m Microservice) error
	FirstByID(ID string) (Microservice, error)
	FirstByName(name string) (Microservice, error)
	FirstOrCreateByName(name string) (Microservice, error)
	SelectAll() ([]Microservice, error)
}

func (m *Microservice) Validate(ve Validator) ValidationErrors {
	if !ve.IsUUID4(m.ID) {
		ve.Add("ID", ":id must be a valid UUID4 or be null for auto assigning")
	}

	if m.Name == "" {
		ve.Add("name", ":name field is required")
	}

	if len(m.Name) > 36 {
		ve.Add("name", ":name should be less than 36 characters")
	}

	if len(m.Description) > 255 {
		ve.Add("description", ":description should be less than 255 characters")
	}

	return ve.Errors()
}
