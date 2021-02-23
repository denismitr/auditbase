package mysql

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/denismitr/auditbase/db"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/validator"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

type microserviceRecord struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	CreatedAt   string `db:"created_at"`
	UpdatedAt   string `db:"updated_at"`
}

func (m *microserviceRecord) ToModel() *model.Microservice {
	return &model.Microservice{
		ID: model.ID(m.ID),
		Name: m.Name,
		Description: m.Description,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

type MicroserviceRepository struct {
	*Tx
}

// Create microservices in MySQL DB and return a newly created instance
func (r *MicroserviceRepository) Create(ctx context.Context, m *model.Microservice) (*model.Microservice, error) {
	createSQL, createArgs, err := createMicroserviceQuery(m)
	if err != nil {
		return nil, err
	}

	selectSQL, selectArgs, err := firstMicroserviceByIDQuery(m.ID.String())
	if err != nil {
		return nil, err
	}

	createStmt, err := r.mysqlTx.PreparexContext(ctx, createSQL)
	if err != nil {
		return nil, errors.Wrapf(err, "could not prepare insert statement %s", createSQL)
	}

	defer func() { _ = createStmt.Close() }()

	selectStmt, err := r.mysqlTx.PreparexContext(ctx, selectSQL)
	if err != nil {
		return nil, errors.Wrapf(err, "could not prepare select statement %s", selectSQL)
	}

	defer func() { _ = selectStmt.Close() }()

	if _, err := createStmt.ExecContext(ctx, createArgs...); err != nil {
		return nil, errors.Wrapf(err, "cannot insert record into microservices table")
	}

	var ms microserviceRecord

	if err := selectStmt.GetContext(ctx, &ms, selectArgs...); err != nil {
		r.lg.Error(err)
		return nil, model.ErrMicroserviceNotFound
	}

	return ms.ToModel(), nil
}

// SelectAll microservices
func (r *MicroserviceRepository) SelectAll() ([]*model.Microservice, error) {
	stmt := `SELECT BIN_TO_UUID(id) as id, name, description, created_at, updated_at FROM microservices`
	ms := []microserviceRecord{}

	if err := r.mysqlTx.Select(&ms, stmt); err != nil {
		return nil, errors.Wrap(err, "could not select all microservices")
	}

	result := make([]*model.Microservice, len(ms))

	for i := range ms {
		result[i] = &model.Microservice{
			ID:          model.ID(ms[i].ID),
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

	if _, err := r.mysqlTx.Exec(stmt, ID.String()); err != nil {
		return errors.Wrapf(err, "could not delete microservices with ID %s", ID.String())
	}

	return nil
}

// Update microservices by ID
func (r *MicroserviceRepository) Update(ID model.ID, m *model.Microservice) error {
	stmt := `UPDATE microservices SET name = ?, description = ? WHERE id = UUID_TO_BIN(?)`

	if _, err := r.mysqlTx.Exec(stmt, m.Name, m.Description, ID.String()); err != nil {
		return errors.Wrapf(err, "could not update microservices with ID %s", m.ID)
	}

	return nil
}

// FirstByID - find one microservices by ID
func (r *MicroserviceRepository) FirstByID(ID model.ID) (*model.Microservice, error) {
	var m microserviceRecord

	q, args, err := firstMicroserviceByIDQuery(ID.String())
	if err != nil {
		return nil, err
	}

	stmt, err := r.mysqlTx.Preparex(q)
	if err != nil {
		return nil, errors.Wrapf(err, "could not prepare select statement %s", q)
	}

	defer func() { _ = stmt.Close() }()

	if err := stmt.Get(&m, args...); err != nil {
		r.lg.Error(err)
		return nil, model.ErrMicroserviceNotFound
	}

	return &model.Microservice{
		ID:          model.ID(m.ID),
		Name:        m.Name,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}, nil
}

// FirstByName - gets first microservices by its name
func (r *MicroserviceRepository) FirstByName(ctx context.Context, name string) (*model.Microservice, error) {
	m := new(microserviceRecord)

	query := `
		SELECT 
			BIN_TO_UUID(id) as id, name, description, created_at, updated_at 
		FROM microservices 
			WHERE name = ?
	`
	stmt, err := r.mysqlTx.PreparexContext(ctx, query)
	if err != nil {
		return nil, errors.Wrapf(err, "could not prepare %s", query)
	}

	defer func() { _ = stmt.Close() }()

	if err := stmt.GetContext(ctx, m, name); err != nil {
		return nil, errors.Wrapf(err, "could not get microservices with name %s from database", name)
	}

	return &model.Microservice{
		ID:          model.ID(m.ID),
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
		r.lg.Error(err)
	}

	m = &model.Microservice{
		ID:          model.ID(r.uuid4.Generate()),
		Name:        name,
		Description: "",
	}

	if _, err := r.Create(ctx, m); err != nil {
		return nil, errors.Wrapf(err, "microservices with name %s does not exist and cannot be created", name)
	}

	return m, nil
}

func createMicroserviceQuery(m *model.Microservice) (string, []interface{}, error) {
	if validator.IsEmptyString(m.ID.String()) {
		return "", nil, db.ErrEmptyUUID4
	}

	if !validator.IsUUID4(m.ID.String()) {
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
