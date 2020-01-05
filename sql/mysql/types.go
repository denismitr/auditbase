package mysql

import (
	"github.com/denismitr/auditbase/model"
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

type targetType struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	CreatedAt   string `db:"created_at"`
	UpdatedAt   string `db:"updated_at"`
}

func (tt targetType) ToModel() model.TargetType {
	return model.TargetType{
		ID:          tt.ID,
		Name:        tt.Name,
		Description: tt.Description,
		CreatedAt:   tt.CreatedAt,
		UpdatedAt:   tt.UpdatedAt,
	}
}

type TargetTypeRepository struct {
	Conn *sqlx.DB
}

func (r *TargetTypeRepository) Create(mtt model.TargetType) error {
	stmt := `
		INSERT INTO target_types (id, name, description) VALUES (
			UUID_TO_BIN(:id), :name, :description
		)
	`

	tt := targetType{
		ID:          mtt.ID,
		Name:        mtt.Name,
		Description: mtt.Description,
	}

	if _, err := r.Conn.NamedExec(stmt, tt); err != nil {
		return errors.Wrapf(err, "could not create new target type with name %s", mtt.Name)
	}

	return nil
}

func (r *TargetTypeRepository) FirstByName(name string) (model.TargetType, error) {
	stmt := `
		SELECT 
			BIN_TO_UUID(id) as id, name, description, created_at, updated_at 
		FROM target_types
			WHERE name = ?
		LIMIT 1
	`

	tt := targetType{}

	if err := r.Conn.Get(&tt, stmt, name); err != nil {
		return tt.ToModel(), errors.Wrapf(err, "could not find target type with name %s", name)
	}

	return tt.ToModel(), nil
}

type ActorTypeRepository struct {
	Conn *sqlx.DB
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

	if _, err := r.Conn.NamedExec(stmt, tt); err != nil {
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

	if err := r.Conn.Get(&at, stmt, name); err != nil {
		return at.ToModel(), errors.Wrapf(err, "could not find actor type with name %s", name)
	}

	return at.ToModel(), nil
}
