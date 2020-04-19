package mysql

import (
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type entity struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	ServiceID   string `db:"service_id"`
	Description string `db:"description"`
	CreatedAt   string `db:"created_at"`
	UpdatedAt   string `db:"updated_at"`
}

func (e entity) ToModel() *model.Entity {
	return &model.Entity{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

type EntityRepository struct {
	conn  *sqlx.DB
	uuid4 uuid.UUID4Generator
}

func NewEntityRepository(conn *sqlx.DB, uuid4 uuid.UUID4Generator) *EntityRepository {
	return &EntityRepository{
		conn:  conn,
		uuid4: uuid4,
	}
}

const selectEntities = `
	SELECT 
		BIN_TO_UUID(id) as id, BIN_TO_UUID(service_id) as id, name, description, created_at, updated_at 
	FROM entities
`

// Select all entities
func (r *EntityRepository) Select() ([]*model.Entity, error) {
	entities := []entity{}

	if err := r.conn.Select(&entities, selectEntities); err != nil {
		return nil, errors.Wrap(err, "could not select all entitys")
	}

	result := make([]*model.Entity, len(entities))

	for i := range entities {
		result[i] = entities[i].ToModel()
	}

	return result, nil
}

// Create an entity
func (r *EntityRepository) Create(e *model.Entity) error {
	stmt := `
		INSERT INTO entities (id, service_id, name, description) VALUES (
			UUID_TO_BIN(:id), UUID_TO_BIN(:service_id), :name, :description
		)
	`

	tt := entity{
		ID:          e.ID,
		Name:        e.Name,
		ServiceID:   e.Service.ID,
		Description: e.Description,
	}

	if _, err := r.conn.NamedExec(stmt, tt); err != nil {
		return errors.Wrapf(err, "could not create new entity with name %s", e.Name)
	}

	return nil
}

func (r *EntityRepository) FirstByNameAndService(name string, service *model.Microservice) (*model.Entity, error) {
	stmt := `
		SELECT 
			BIN_TO_UUID(id) a
			s id, BIN_TO_UUID(service_id) as service_id, name, description, created_at, updated_at 
		FROM entities
			WHERE service_id = UUID_TO_BIN(?) AND name = ?
		LIMIT 1
	`

	ent := entity{}

	if err := r.conn.Get(&ent, stmt, name, service.ID); err != nil {
		return nil, errors.Wrapf(err, "could not find entity with name %s", name)
	}

	return ent.ToModel(), nil
}

func (r *EntityRepository) FirstByID(ID string) (*model.Entity, error) {
	stmt := `
		SELECT 
			BIN_TO_UUID(id) as id, BIN_TO_UUID(service_id) as service_id, name, description, created_at, updated_at 
		FROM entities
			WHERE id = UUID_TO_BIN(?)
		LIMIT 1
	`

	ent := entity{}

	if err := r.conn.Get(&ent, stmt, ID); err != nil {
		return nil, errors.Wrapf(err, "could not find entity with ID %s", ID)
	}

	return ent.ToModel(), nil
}

// FirstOrCreateByNameAndService - fetches first or creates an entity by its name and service
func (r *EntityRepository) FirstOrCreateByNameAndService(name string, service *model.Microservice) (*model.Entity, error) {
	ent, err := r.FirstByNameAndService(name, service)
	if err == nil {
		return ent, nil
	}

	ent = &model.Entity{
		ID:          r.uuid4.Generate(),
		Service:     *service,
		Name:        name,
		Description: "",
	}

	if err := r.Create(ent); err != nil {
		return nil, errors.Wrapf(err, "entity %s does not exist and cannot be created", name)
	}

	return ent, nil
}
