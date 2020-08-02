package mysql

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/denismitr/auditbase/db"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/denismitr/auditbase/utils/uuid"
	"github.com/denismitr/auditbase/utils/validator"
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

func (m *microservice) ToModel() *model.Microservice {
	return &model.Microservice{
		ID: m.ID,
		Name: m.Name,
		Description: m.Description,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

type MicroserviceRepository struct {
	conn  *sqlx.DB
	uuid4 uuid.UUID4Generator
	log logger.Logger
}

// NewMicroserviceRepository - constructor function
func NewMicroserviceRepository(
	conn *sqlx.DB,
	uuid4 uuid.UUID4Generator,
	log logger.Logger,
) *MicroserviceRepository {
	return &MicroserviceRepository{
		conn:  conn,
		uuid4: uuid4,
		log: log,
	}
}

// Create microservices in MySQL DB and return a newly created instance
func (r *MicroserviceRepository) Create(ctx context.Context, m *model.Microservice) (*model.Microservice, error) {
	createSQL, createArgs, err := createMicroserviceQuery(m)
	if err != nil {
		return nil, err
	}

	selectSQL, selectArgs, err := firstMicroserviceByIDQuery(m.ID)
	if err != nil {
		return nil, err
	}

	createStmt, err := r.conn.PreparexContext(ctx, createSQL)
	if err != nil {
		return nil, errors.Wrapf(err, "could not prepare insert statement %s", createSQL)
	}

	selectStmt, err := r.conn.PreparexContext(ctx, selectSQL)
	if err != nil {
		return nil, errors.Wrapf(err, "could not prepare select statement %s", selectSQL)
	}

	if _, err := createStmt.ExecContext(ctx, createArgs...); err != nil {
		return nil, errors.Wrapf(err, "cannot insert record into microservices table")
	}

	var ms microservice

	if err := selectStmt.GetContext(ctx, &ms, selectArgs...); err != nil {
		r.log.Error(err)
		return nil, model.ErrMicroserviceNotFound
	}

	return ms.ToModel(), nil
}

// SelectAll microservices
func (r *MicroserviceRepository) SelectAll() ([]*model.Microservice, error) {
	stmt := `SELECT BIN_TO_UUID(id) as id, name, description, created_at, updated_at FROM microservices`
	ms := []microservice{}

	if err := r.conn.Select(&ms, stmt); err != nil {
		return nil, errors.Wrap(err, "could not select all microservices")
	}

	result := make([]*model.Microservice, len(ms))

	for i := range ms {
		result[i] = &model.Microservice{
			ID:          ms[i].ID,
			Name:        ms[i].Name,
			Description: ms[i].Description,
			CreatedAt:   ms[i].CreatedAt,
			UpdatedAt:   ms[i].UpdatedAt,
		}
	}

	return result, nil
}

// Delete microservices by ID
func (r *MicroserviceRepository) Delete(ID model.ID) error {
	stmt := `DELETE FROM microservices WHERE id = UUID_TO_BIN(?)`

	if _, err := r.conn.Exec(stmt, ID.String()); err != nil {
		return errors.Wrapf(err, "could not delete microservices with ID %s", ID.String())
	}

	return nil
}

// Update microservices by ID
func (r *MicroserviceRepository) Update(ID model.ID, m *model.Microservice) error {
	stmt := `UPDATE microservices SET name = ?, description = ? WHERE id = UUID_TO_BIN(?)`

	if _, err := r.conn.Exec(stmt, m.Name, m.Description, ID.String()); err != nil {
		return errors.Wrapf(err, "could not update microservices with ID %s", m.ID)
	}

	return nil
}

// FirstByID - find one microservices by ID
func (r *MicroserviceRepository) FirstByID(ID model.ID) (*model.Microservice, error) {
	var m microservice

	q, args, err := firstMicroserviceByIDQuery(ID.String())
	if err != nil {
		return nil, err
	}

	stmt, err := r.conn.Preparex(q)
	if err != nil {
		return nil, errors.Wrapf(err, "could not prepare select statement %s", q)
	}

	if err := stmt.Get(&m, args...); err != nil {
		r.log.Error(err)
		return nil, model.ErrMicroserviceNotFound
	}

	return &model.Microservice{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}, nil
}

// FirstByName - gets first microservices by its name
func (r *MicroserviceRepository) FirstByName(ctx context.Context, name string) (*model.Microservice, error) {
	m := new(microservice)

	query := `
		SELECT 
			BIN_TO_UUID(id) as id, name, description, created_at, updated_at 
		FROM microservices 
			WHERE name = ?
	`
	stmt, err := r.conn.PreparexContext(ctx, query)
	if err != nil {
		return nil, errors.Wrapf(err, "could not prepare %s", query)
	}

	if err := stmt.GetContext(ctx, m, name); err != nil {
		return nil, errors.Wrapf(err, "could not get microservices with name %s from database", name)
	}

	return &model.Microservice{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}, nil
}

// FirstOrCreateByName - gets first microservices with given name or tries to create
// a new one, assigning new UUID4
// fixme: refactor to transaction
func (r *MicroserviceRepository) FirstOrCreateByName(ctx context.Context, name string) (*model.Microservice, error) {
	m, err := r.FirstByName(ctx, name)
	if err == nil {
		return m, nil
	} else {
		r.log.Error(err)
	}

	m = &model.Microservice{
		ID:          r.uuid4.Generate(),
		Name:        name,
		Description: "",
	}

	if _, err := r.Create(ctx, m); err != nil {
		return nil, errors.Wrapf(err, "microservices with name %s does not exist and cannot be created", name)
	}

	return m, nil
}

func createMicroserviceQuery(m *model.Microservice) (string, []interface{}, error) {
	if validator.IsEmptyString(m.ID) {
		return "", nil, db.ErrEmptyUUID4
	}

	if !validator.IsUUID4(m.ID) {
		return "", nil, errors.Errorf("%s is not a valid uuid4", m.ID)
	}

	return sq.Insert("microservices").
		Columns("id", "name", "description").
		Values(sq.Expr("UUID_TO_BIN(?)", m.ID), m.Name, m.Description).
		ToSql()
}

func firstMicroserviceByIDQuery(ID string) (string, []interface{}, error) {
	if validator.IsEmptyString(ID) {
		return "", nil, db.ErrEmptyUUID4
	}

	if !validator.IsUUID4(ID) {
		return "", nil, errors.Errorf("%s is not a valid uuid4", ID)
	}

	return sq.Select(
		"BIN_TO_UUID(id) as id",
			"name", "description",
			"created_at", "updated_at",
		).
		From("microservices").
		Where("id = UUID_TO_BIN(?)", ID).
		ToSql()
}
