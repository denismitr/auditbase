package mysql

import (
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type microservice struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	CreatedAt   string `db:"created_at"`
	UpdatedAt   string `db:"updated_at"`
}

type MicroserviceRepository struct {
	conn  *sqlx.DB
	uuid4 utils.UUID4Generatgor
}

// NewMicroserviceRepository - constructor function
func NewMicroserviceRepository(
	conn *sqlx.DB,
	uuid4 utils.UUID4Generatgor,
) *MicroserviceRepository {
	return &MicroserviceRepository{
		conn:  conn,
		uuid4: uuid4,
	}
}

func (r *MicroserviceRepository) Create(m model.Microservice) error {
	stmt := "INSERT INTO microservices (id, name, description) VALUES (UUID_TO_BIN(?), ?, ?)"

	if _, err := r.conn.Exec(stmt, m.ID, m.Name, m.Description); err != nil {
		return errors.Wrapf(err, "cannot insert into microservices table")
	}

	return nil
}

// SelectAll microservices
func (r *MicroserviceRepository) SelectAll() ([]model.Microservice, error) {
	stmt := `SELECT BIN_TO_UUID(id) as id, name, description, created_at, updated_at FROM microservices`
	ms := []microservice{}

	if err := r.conn.Select(&ms, stmt); err != nil {
		return []model.Microservice{}, errors.Wrap(err, "could not select all microservices")
	}

	result := make([]model.Microservice, len(ms))

	for i := range ms {
		result[i] = model.Microservice{
			ID:          ms[i].ID,
			Name:        ms[i].Name,
			Description: ms[i].Description,
			CreatedAt:   ms[i].CreatedAt,
			UpdatedAt:   ms[i].UpdatedAt,
		}
	}

	return result, nil
}

// Delete microservice by ID
func (r *MicroserviceRepository) Delete(ID string) error {
	stmt := `DELETE FROM microservices WHERE id = UUID_TO_BIN(?)`

	if _, err := r.conn.Exec(stmt, ID); err != nil {
		return errors.Wrapf(err, "could not delete microservice with ID %s", ID)
	}

	return nil
}

// Update microservice by ID
func (r *MicroserviceRepository) Update(ID string, m model.Microservice) error {
	stmt := `UPDATE microservices SET name = ?, description = ? WHERE id = UUID_TO_BIN(?)`

	if _, err := r.conn.Exec(stmt, m.Name, m.Description, ID); err != nil {
		return errors.Wrapf(err, "could not update microservice with ID %s", m.ID)
	}

	return nil
}

// FirstByID - find first microservice by ID
func (r *MicroserviceRepository) FirstByID(ID string) (model.Microservice, error) {
	m := new(microservice)

	stmt := `SELECT BIN_TO_UUID(id) as id, name, description, created_at, updated_at FROM microservices WHERE id = UUID_TO_BIN(?)`

	if err := r.conn.Get(m, stmt, ID); err != nil {
		return model.Microservice{}, errors.Wrapf(err, "could not get microservice with ID %s from database", ID)
	}

	return model.Microservice{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}, nil
}

// FirstByName - gets first microservice by its name
func (r *MicroserviceRepository) FirstByName(name string) (model.Microservice, error) {
	m := new(microservice)

	stmt := `
		SELECT 
			BIN_TO_UUID(id) as id, name, description, created_at, updated_at 
		FROM microservices 
			WHERE name = ?
	`

	if err := r.conn.Get(m, stmt, name); err != nil {
		return model.Microservice{},
			errors.Wrapf(err, "could not get microservice with name %s from database", name)
	}

	return model.Microservice{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}, nil
}

// FirstOrCreateByName - gets first microservice with given name or tries to create
// a new one, assigning new UUID4
func (r *MicroserviceRepository) FirstOrCreateByName(name string) (model.Microservice, error) {
	m, err := r.FirstByName(name)
	if err == nil {
		return m, nil
	}

	m = model.Microservice{
		ID:          r.uuid4.Generate(),
		Name:        name,
		Description: "",
	}

	if err := r.Create(m); err != nil {
		return model.Microservice{},
			errors.Wrapf(err, "microservice with name %s does not exist and cannot be created", name)
	}

	return m, nil
}
