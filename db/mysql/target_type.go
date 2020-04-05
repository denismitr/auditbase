package mysql

import (
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

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

// TargetTypeRepository - implements model.TargetTypeRepository
type TargetTypeRepository struct {
	conn  *sqlx.DB
	uuid4 utils.UUID4Generatgor
}

func NewTargetTypeRepository(conn *sqlx.DB, uuid4 utils.UUID4Generatgor) *TargetTypeRepository {
	return &TargetTypeRepository{
		conn:  conn,
		uuid4: uuid4,
	}
}

// Select all target types in DB
func (r *TargetTypeRepository) Select() ([]model.TargetType, error) {
	targetTypes := []targetType{}
	result := []model.TargetType{}

	if err := r.conn.Select(&targetTypes, selectTargetTypes); err != nil {
		return result, errors.Wrap(err, "could not select all target types")
	}

	for i := range targetTypes {
		result = append(result, targetTypes[i].ToModel())
	}

	return result, nil
}

// Create a new target type from model DTO
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

	if _, err := r.conn.NamedExec(stmt, tt); err != nil {
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

	if err := r.conn.Get(&tt, stmt, name); err != nil {
		return tt.ToModel(), errors.Wrapf(err, "could not find target type with name %s", name)
	}

	return tt.ToModel(), nil
}

// FirstByID - fetches a target type by ID
func (r *TargetTypeRepository) FirstByID(ID string) (model.TargetType, error) {
	stmt := `
		SELECT 
			BIN_TO_UUID(id) as id, name, description, created_at, updated_at 
		FROM target_types
			WHERE id = UUID_TO_BIN(?)
		LIMIT 1
	`

	tt := targetType{}

	if err := r.conn.Get(&tt, stmt, ID); err != nil {
		return tt.ToModel(), errors.Wrapf(err, "could not find target type with ID %s", ID)
	}

	return tt.ToModel(), nil
}

func (r *TargetTypeRepository) FirstOrCreateByName(name string) (model.TargetType, error) {
	tt, err := r.FirstByName(name)
	if err == nil {
		return tt, nil
	}

	tt = model.TargetType{
		ID:          r.uuid4.Generate(),
		Name:        name,
		Description: "",
	}

	if err := r.Create(tt); err != nil {
		return model.TargetType{}, errors.Wrapf(err, "target type %s does not exist and cannot be created", name)
	}

	return tt, nil
}
