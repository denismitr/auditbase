package mysql

import (
	"context"
	"database/sql"
	sq "github.com/Masterminds/squirrel"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/denismitr/auditbase/utils/uuid"
	"github.com/denismitr/auditbase/utils/validator"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"time"
)

type selectQuery struct {
	selectSQL string
	selectArgs []interface{}
	countSQL string
	countArgs []interface{}
}

type PropertyRepository struct {
	conn   *sqlx.DB
	log logger.Logger
	uuid4  uuid.UUID4Generator
}

func NewPropertyRepository(conn *sqlx.DB, uuid4 uuid.UUID4Generator, log logger.Logger) *PropertyRepository {
	return &PropertyRepository{
		conn: conn,
		uuid4: uuid4,
		log: log,
	}
}

type property struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	EntityID    string `db:"entity_id"`
	ChangeCount int    `db:"change_count"`
	LastEventAt sql.NullTime `db:"last_event_at"`
}

func (p *property) ToModel() *model.Property {
	var t *time.Time
	if p.LastEventAt.Valid {
		t = &p.LastEventAt.Time
	}

	return &model.Property{
		ID:          p.ID,
		EntityID:    p.EntityID,
		Name:        p.Name,
		ChangeCount: p.ChangeCount,
		LastEventAt: t,
	}
}

func (p *PropertyRepository) FirstByID(ID string) (*model.Property, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	q, args, err := createFirstByIDQuery(ID)
	if err != nil {
		return nil, err
	}

	stmt, err := p.conn.PreparexContext(ctx, q)
	if err != nil {
		return nil, errors.Wrapf(err, "could not prepare %s statement", q)
	}

	defer func() { _ = stmt.Close() }()

	var prop property
	if err := stmt.GetContext(ctx, &prop, args...); err != nil {
		return nil, errors.Wrapf(err, "could not get properties with ID %s", ID)
	}

	return prop.ToModel(), nil
}

func (p *PropertyRepository) Select(filter *model.Filter, sort *model.Sort, pagination *model.Pagination) ([]*model.Property, *model.Meta, error) {
	q, err := createSelectPropertiesQuery(filter, sort, pagination)
	if err != nil {
		return nil, nil, err
	}

	selectStmt, err := p.conn.Preparex(q.selectSQL)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "could not prepare %s statement", q.selectSQL)
	}

	defer func() { _ = selectStmt.Close() }()

	countStmt, err :=  p.conn.Preparex(q.countSQL)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "could not prepare %s statement", q.countSQL)
	}

	defer func() { _ = countStmt.Close() }()

	var props []property
	var m meta

	if err := selectStmt.Select(&props, q.selectArgs...); err != nil {
		return nil, nil, errors.Wrap(err, "could not select properties")
	}

	if err := countStmt.Get(&m, q.selectArgs...); err != nil {
		return nil, nil, errors.Wrap(err, "could not select properties")
	}

	var result = make([]*model.Property, len(props))
	for i := range props {
		result[i] = props[i].ToModel()
	}

	return result, m.ToModel(pagination), nil
}

func (r *PropertyRepository) GetIDOrCreate(name, entityID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	var result string

	createSql, createArgs, err := createInsertPropertyQuery(r.uuid4.Generate(), name, entityID)
	if err != nil {
		return result, errors.Wrap(err, "could not create insert properties query")
	}

	getSql, getArgs, err := createGetPropertyIDQuery(name, entityID)
	if err != nil {
		return result, errors.Wrap(err, "could not create get properties ID query")
	}

	createStmt, err := r.conn.PreparexContext(ctx, createSql)
	if err != nil {
		return result, errors.Wrap(err, "could not prepare insert properties query")
	}

	defer func() { _ = createStmt.Close() }()

	getStmt, err := r.conn.PreparexContext(ctx, getSql)
	if err != nil {
		return result, errors.Wrap(err, "could not prepare get properties ID query")
	}

	defer func() { _ = getStmt.Close() }()

	if _, err := createStmt.ExecContext(ctx, createArgs...); err != nil {
		r.log.Error(err)
	}

	rows, err := getStmt.Query(getArgs...)
	if err != nil {
		return result, errors.Wrapf(err, "could not select id from properties with name %s and eventID %s", name, entityID)
	}

	defer func() { _ = rows.Close() }()

	for rows.Next() {
		if err := rows.Scan(&result); err != nil {
			return result, errors.Wrapf(err, "could not parse properties ID for name %s and eventID %s", name, entityID)
		}

		return result, nil
	}

	return result, errors.Errorf("failed to create or retrieve properties with name %s and entityId %s", name, entityID)
}

func createInsertPropertyQuery(ID, name, entityID string) (string, []interface{}, error) {
	if validator.IsEmptyString(name) {
		return "", nil, errors.New("properties name is empty")
	}

	if ! validator.IsUUID4(entityID) {
		return "", nil, errors.Errorf("%s is not a valid uuid4", entityID)
	}

	if ! validator.IsUUID4(ID) {
		return "", nil, errors.Errorf("%s is not a valid uuid4", ID)
	}

	return sq.Insert("properties").
		Columns("id", "name", "entity_id").
		Values(
			sq.Expr("UUID_TO_BIN(?)", ID),
			name,
			sq.Expr("UUID_TO_BIN(?)", entityID),
		).ToSql()
}

func createGetPropertyIDQuery(name, entityID string) (string, []interface{}, error) {
	if validator.IsEmptyString(name) {
		return "", nil, errors.New("properties name is empty")
	}

	if ! validator.IsUUID4(entityID) {
		return "", nil, errors.Errorf("%s is not a valid uuid4", entityID)
	}

	return sq.
		Select("BIN_TO_UUID(id) as id").
		From("properties").
		Where(sq.Eq{"name": name}).
		Where("entity_id = UUID_TO_BIN(?)", entityID).
		Limit(1).ToSql()
}

func createFirstByIDQuery(ID string) (string, []interface{}, error) {
	if ! validator.IsUUID4(ID) {
		return "", nil, errors.Errorf("%s is not a valid uuid4", ID)
	}

	return sq.Select(
		"BIN_TO_UUID(p.id) as id",
		"BIN_TO_UUID(p.entity_id) as entity_id",
		"p.name",
	).
		From("properties as p").
		Where("p.id = UUID_TO_BIN(?)", ID).
		ToSql()
}

func createSelectPropertiesQuery(
	f *model.Filter,
	sort *model.Sort,
	pg *model.Pagination,
) (*selectQuery, error) {
	sQ := sq.Select(
		"BIN_TO_UUID(p.id) as id",
			"BIN_TO_UUID(p.entity_id) as entity_id",
			"p.name",
			"max(e.emitted_at) as last_event_at",
			"count(c.id) as change_count",
		).
		From("properties as p").
		Join("changes c ON c.property_id = p.id").
		Join("events e ON c.event_id = e.id")

	cQ := sq.Select("COUNT(*) as total").From("properties as p")

	if f.Has("entityId") {
		entityId := f.MustString("entityId")
		if ! validator.IsUUID4(entityId) {
			return nil, errors.Errorf("%s is not a valid UUID 4", entityId)
		}
		sQ = sQ.Where("p.entity_id = UUID_TO_BIN(?)", entityId)
		cQ = cQ.Where("p.entity_id = UUID_TO_BIN(?)", entityId)
	}

	if f.Has("name") {
		name := f.MustString("name")
		if validator.IsEmptyString(name) {
			return nil, errors.Errorf("name %s cannot be an empty string", name)
		}
		sQ = sQ.Where("p.name = name", name)
		cQ = cQ.Where("p.name = name", name)
	}

	sQ = sQ.GroupBy("p.id", "p.entity_id", "p.name")

	if sort.Empty() {
		sQ = sQ.OrderBy("max(e.emitted_at) DESC")
	} else {
		for column, order := range sort.All() {
			sQ = sQ.OrderByClause("? ?", column, order.String())
		}
	}

	selectSQL, selectArgs, err := sQ.Limit(uint64(pg.PerPage)).Offset(uint64(pg.Offset())).ToSql()
	if err != nil {
		return nil, errors.New("could not build select properties query")
	}

	countSQL, countArgs, err := cQ.ToSql()
	if err != nil {
		return nil, errors.New("could not build count properties query for pagination")
	}

	return &selectQuery{
		selectSQL: selectSQL,
		selectArgs: selectArgs,
		countSQL: countSQL,
		countArgs: countArgs,
	}, nil
}


