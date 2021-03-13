package mysql

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/denismitr/auditbase/db"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/validator"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"time"
)

type microserviceRecord struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (m *microserviceRecord) ToModel() *model.Microservice {
	return &model.Microservice{
		ID: model.ID(m.ID),
		Name: m.Name,
		Description: m.Description,
		CreatedAt: model.JSONTime{Time: m.CreatedAt},
		UpdatedAt: model.JSONTime{Time: m.UpdatedAt},
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

	createStmt, err := r.mysqlTx.PreparexContext(ctx, createSQL)
	if err != nil {
		return nil, errors.Wrapf(err, "could not prepare insert statement %s", createSQL)
	}

	defer func() { _ = createStmt.Close() }()

	if _, err := createStmt.ExecContext(ctx, createArgs...); err != nil {
		return nil, errors.Wrapf(err, "cannot insert record into microservices table")
	}

	return r.FirstByID(ctx, m.ID)
}

// SelectAll microservices
// fixme: pagination
func (r *MicroserviceRepository) SelectAll(ctx context.Context) (*model.MicroserviceCollection, error) {
	q := `SELECT BIN_TO_UUID(id) as id, name, description, created_at, updated_at FROM microservices`
	var msr []microserviceRecord

	stmt, err := r.mysqlTx.PreparexContext(ctx, q)
	if err != nil {
		return nil, errors.Wrap(err, "could not prepare select all microservices query")
	}

	if err := stmt.SelectContext(ctx, &msr); err != nil {
		return nil, errors.Wrap(err, "could not select all microservices")
	}

	var result model.MicroserviceCollection

	for _, ms := range msr {
		result.Items = append(result.Items, model.Microservice{
			ID:          model.ID(ms.ID),
			Name:        ms.Name,
			Description: ms.Description,
			CreatedAt:   model.JSONTime{Time: ms.CreatedAt},
			UpdatedAt:   model.JSONTime{Time: ms.UpdatedAt},
		})
	}

	result.Meta.Total = len(result.Items)
	result.Meta.Page = 1
	result.Meta.PerPage = 10000

	return nil, nil
}

func selectAllMicroservicesQuery() (string, interface{}, error) {
	dialect := goqu.Dialect(MySQL8)

	return dialect.Select(
		goqu.L("bin_to_uuid(id)").As("id"),
		"name", "description",
		"created_at", "updated_at",
	).Prepared(true).ToSQL()
}

// Delete microservices by ID
func (r *MicroserviceRepository) Delete(ctx context.Context, ID model.ID) error {
	q, args, err := deleteMicroserviceQuery(ID)
	if err != nil {
		panic(err)
	}

	stmt, err := r.mysqlTx.PreparexContext(ctx, q)
	if err != nil {
		return errors.Wrap(err, "could not prepare delete microservice query")
	}

	if _, err := stmt.ExecContext(ctx, args); err != nil {
		return errors.Wrapf(err, "could not delete microservices with ID %s", ID.String())
	}

	return nil
}

func deleteMicroserviceQuery(ID model.ID) (string, []interface{}, error) {
	if ! validator.IsUUID4(ID.String()) {
		return "", nil, errors.New("how could id not be a valid UUID")
	}

	dialect := goqu.Dialect(MySQL8)
	expr := goqu.L("id = uuid_to_bin(?)", ID.String())
	return dialect.Delete("microservices").Where(expr).Prepared(true).ToSQL()
}

// Update microservices by ID
func (r *MicroserviceRepository) Update(ctx context.Context, ID model.ID, m *model.Microservice) (*model.Microservice, error) {
	q, args, err := updateMicroserviceQuery(ID, m)
	if err != nil {
		panic(err)
	}

	stmt, err := r.mysqlTx.PreparexContext(ctx, q)
	if err != nil {
		return nil, errors.Wrap(err, "could not prepare update microservices query")
	}

	if _, err := stmt.ExecContext(ctx, args...); err != nil {
		return nil, errors.Wrapf(err, "could not update microservices with ID %s", m.ID.String())
	}

	return nil, nil
}

func updateMicroserviceQuery(ID model.ID, m *model.Microservice) (string, []interface{}, error) {
	if ! validator.IsUUID4(ID.String()) {
		return "", nil, errors.New("how could id not be a valid UUID")
	}

	if m.Name == "" {
		return "", nil, errors.New("how can microservice name be empty on update?")
	}

	if m.UpdatedAt.IsZero() {
		return "", nil, errors.New("how can microservice updated at time be zero?")
	}

	dialect := goqu.Dialect(MySQL8)
	whereExpr := goqu.L("id = uuid_to_bin(?)", ID.String())
	return dialect.Update("microservices").Where(whereExpr).Set(goqu.Record{
		"name": m.Name,
		"description": m.Description,
		"updated_at": m.UpdatedAt.Unix(),
	}).Prepared(true).ToSQL()
 }

// FirstByID - find one microservices by ID
func (r *MicroserviceRepository) FirstByID(ctx context.Context, ID model.ID) (*model.Microservice, error) {
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
		CreatedAt:   model.JSONTime{Time: m.CreatedAt},
		UpdatedAt:   model.JSONTime{Time: m.UpdatedAt},
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
		CreatedAt:   model.JSONTime{Time: m.CreatedAt},
		UpdatedAt:   model.JSONTime{Time: m.UpdatedAt},
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
