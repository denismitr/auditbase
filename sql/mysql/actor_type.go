package mysql

import (
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type actorType struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	CreatedAt   string `db:"created_at"`
	UpdatedAt   string `db:"updated_at"`
}

func (at actorType) ToModel() model.ActorType {
	return model.ActorType{
		ID:          at.ID,
		Name:        at.Name,
		Description: at.Description,
		CreatedAt:   at.CreatedAt,
		UpdatedAt:   at.UpdatedAt,
	}
}

type ActorTypeRepository struct {
	conn  *sqlx.DB
	uuid4 utils.UUID4Generatgor
}

func NewActorTypeRepository(conn *sqlx.DB, uuid4 utils.UUID4Generatgor) *ActorTypeRepository {
	return &ActorTypeRepository{
		conn:  conn,
		uuid4: uuid4,
	}
}

func (r *ActorTypeRepository) Select() ([]model.ActorType, error) {
	stmt := `
		SELECT 
			BIN_TO_UUID(id) as id, name, description, created_at, updated_at 
		FROM actor_types
	`

	actorTypes := []actorType{}
	result := []model.ActorType{}

	if err := r.conn.Select(&actorTypes, stmt); err != nil {
		return result, errors.Wrap(err, "could not select all actor types")
	}

	for i := range actorTypes {
		result = append(result, model.ActorType{
			ID:          actorTypes[i].ID,
			Name:        actorTypes[i].Name,
			Description: actorTypes[i].Description,
			CreatedAt:   actorTypes[i].CreatedAt,
			UpdatedAt:   actorTypes[i].UpdatedAt,
		})
	}

	return result, nil
}

func (r *ActorTypeRepository) Create(mat model.ActorType) error {
	stmt := `
		INSERT INTO actor_types (id, name, description) VALUES (
			UUID_TO_BIN(:id), :name, :description
		)
	`

	tt := actorType{
		ID:          mat.ID,
		Name:        mat.Name,
		Description: mat.Description,
	}

	if _, err := r.conn.NamedExec(stmt, tt); err != nil {
		return errors.Wrapf(err, "could not create new actor type with name %s", mat.Name)
	}

	return nil
}

func (r *ActorTypeRepository) FirstByName(name string) (model.ActorType, error) {
	stmt := `
		SELECT 
			BIN_TO_UUID(id) as id, name, description, created_at, updated_at 
		FROM actor_types
			WHERE name = ?
		LIMIT 1
	`

	at := actorType{}

	if err := r.conn.Get(&at, stmt, name); err != nil {
		return at.ToModel(), errors.Wrapf(err, "could not find actor type with name %s", name)
	}

	return at.ToModel(), nil
}

func (r *ActorTypeRepository) FirstByID(ID string) (model.ActorType, error) {
	stmt := `
		SELECT 
			BIN_TO_UUID(id) as id, name, description, created_at, updated_at 
		FROM actor_types
			WHERE id = UUID_TO_BIN(?)
		LIMIT 1
	`

	at := actorType{}

	if err := r.conn.Get(&at, stmt, ID); err != nil {
		return at.ToModel(), errors.Wrapf(err, "could not find actor type with ID %s", ID)
	}

	return at.ToModel(), nil
}

func (r *ActorTypeRepository) FirstOrCreateByName(name string) (model.ActorType, error) {
	at, err := r.FirstByName(name)
	if err == nil {
		return at, nil
	}

	at = model.ActorType{
		ID:          r.uuid4.Generate(),
		Name:        name,
		Description: "",
	}

	if err := r.Create(at); err != nil {
		return model.ActorType{}, errors.Wrapf(err, "actor type %s does not exist and cannot be created", name)
	}

	return at, nil
}
