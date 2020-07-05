package mysql

import (
	"bytes"
	"database/sql"
	sq "github.com/Masterminds/squirrel"
	"github.com/denismitr/auditbase/db"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/denismitr/auditbase/utils/uuid"
	"github.com/denismitr/auditbase/utils/validator"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"time"
)

type change struct {
	ID              string         `db:"id"`
	EventID         string         `db:"event_id"`
	PropertyID      string         `db:"property_id"`
	CurrentDataType int            `db:"current_data_type"`
	CreatedAt       time.Time      `db:"created_at"`
	FromValue       sql.NullString `db:"from_value"`
	ToValue         sql.NullString `db:"to_value"`
}

type propertyChange struct {
	ID              string         `db:"id"`
	EventID         string         `db:"event_id"`
	PropertyID      string         `db:"property_id"`
	FromValue       sql.NullString `db:"from_value"`
	ToValue         sql.NullString `db:"to_value"`
	CurrentDataType int            `db:"current_data_type"`
	PropertyName    string         `db:"property_name"`
	EntityID        string         `db:"entity_id"`
}

func (c *propertyChange) ToModel() *model.PropertyChange {
	return &model.PropertyChange{
		ID:              c.ID,
		EventID:         c.EventID,
		EntityID:        c.EntityID,
		From:            db.PointerFromNullString(c.FromValue),
		To:              db.PointerFromNullString(c.ToValue),
		CurrentDataType: model.DataType(c.CurrentDataType),
		PropertyID:      c.PropertyID,
		PropertyName:    c.PropertyName,
	}
}

func (c *change) ToModel() *model.Change {
	m := &model.Change{
		ID:              c.ID,
		EventID:         c.EventID,
		From:            db.PointerFromNullString(c.FromValue),
		To:              db.PointerFromNullString(c.ToValue),
		CurrentDataType: model.DataType(c.CurrentDataType),
		PropertyID:      c.PropertyID,
		CreatedAt:       c.CreatedAt,
	}

	return m
}

type ChangeRepository struct {
	conn   *sqlx.DB
	logger logger.Logger
	uuid4  uuid.UUID4Generator
}

func NewChangeRepository(
	conn *sqlx.DB,
	logger logger.Logger,
	uuid4 uuid.UUID4Generator,
) *ChangeRepository {
	return &ChangeRepository{
		conn:   conn,
		logger: logger,
		uuid4:  uuid4,
	}
}

func (c *ChangeRepository) Select(
	f *model.Filter,
	s *model.Sort,
	p *model.Pagination,
) ([]*model.Change, *model.Meta, error) {
	q, err := selectChangesQuery(f, s, p)
	if err != nil {
		return nil, nil, err
	}

	selectStmt, err := c.conn.Preparex(q.selectSQL)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "could not prepare %s statement", q.selectSQL)
	}
	countStmt, err := c.conn.Preparex(q.countSQL)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "could not prepare %s statement", q.countSQL)
	}

	var cc []change
	var m meta

	if err := selectStmt.Select(&cc, q.selectArgs...); err != nil {
		return nil, nil, errors.Wrap(err, "could not select changes")
	}

	if err := countStmt.Get(&m, q.countArgs...); err != nil {
		return nil, nil, errors.Wrap(err, "could not select changes")
	}

	changes := make([]*model.Change, len(cc))
	for i := range cc {
		changes[i] = cc[i].ToModel()
	}

	return changes, m.ToModel(p), nil
}

func (c *ChangeRepository) FirstByID(ID string) (*model.Change, error) {
	q, args, err := firstChangeByIDQuery(ID)
	if err != nil {
		return nil, err
	}

	stmt, err := c.conn.Preparex(q)
	if err != nil {
		return nil, errors.Wrapf(err, "could not prepare %s statement", q)
	}

	var chng change
	if err := stmt.Get(&chng, args...); err != nil {
		c.logger.Error(errors.Wrapf(err, "could not get change with ID %s", ID))
		return nil, model.ErrChangeNotFound
	}

	return chng.ToModel(), nil
}

func firstChangeByIDQuery(ID string) (string, []interface{}, error) {
	if !validator.IsUUID4(ID) {
		return "", nil, errors.Errorf("%s not a valid uuid4", ID)
	}

	return sq.Select("BIN_TO_UUID(c.id) as id",
		"BIN_TO_UUID(c.property_id) as property_id",
		"BIN_TO_UUID(c.event_id) as event_id",
		"e.emitted_at as created_at",
		"c.current_data_type",
		"c.from_value", "c.to_value",
	).
		From("changes as c").
		Join("events as e on e.id = c.event_id").
		Where("c.id = UUID_TO_BIN(?)", ID).
		Limit(1).
		ToSql()
}

func selectChangesQuery(
	f *model.Filter,
	s *model.Sort,
	p *model.Pagination,
) (*selectQuery, error) {
	selectQ := sq.Select(
		"BIN_TO_UUID(c.id) as id",
		"BIN_TO_UUID(c.property_id) as property_id",
		"BIN_TO_UUID(c.event_id) as event_id",
		"e.emitted_at as created_at",
		"c.current_data_type",
		"c.from_value", "c.to_value",
	).From("changes as c").Join("events as e on e.id = c.event_id")

	countQ := sq.Select("count(*) as total").From("changes as c")

	if s.Empty() {
		selectQ = selectQ.OrderBy("e.emitted_at DESC")
	}

	if f.Has("propertyId") {
		propertyID := f.MustString("propertyId")
		if !validator.IsUUID4(propertyID) {
			return nil, errors.Errorf("%s is not a valid properties ID", propertyID)
		}
		selectQ = selectQ.Where("c.property_id = UUID_TO_BIN(?)", propertyID)
		countQ = countQ.Where("c.property_id = UUID_TO_BIN(?)", propertyID)
	}

	if f.Has("eventId") {
		eventID := f.MustString("eventId")
		if !validator.IsUUID4(eventID) {
			return nil, errors.Errorf("%s is not a valid events ID", eventID)
		}
		selectQ = selectQ.Where("c.event_id = UUID_TO_BIN(?)", eventID)
		countQ = countQ.Where("c.event_id = UUID_TO_BIN(?)", eventID)
	}

	selectSQL, selectArgs, err := selectQ.Limit(uint64(p.PerPage)).Offset(uint64(p.Offset())).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "could not build select changes query")
	}
	countSQl, countArgs, err := countQ.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "could not build select changes query")
	}

	return &selectQuery{
		selectSQL:  selectSQL,
		selectArgs: selectArgs,
		countSQL:   countSQl,
		countArgs:  countArgs,
	}, nil
}

func selectChangesByEventIDsQuery(ids []string) (string, []interface{}, error) {
	if len(ids) == 0 {
		return "", nil, db.ErrEmptyWhereInList
	}

	q := sq.Select(
		"BIN_TO_UUID(c.id) as id",
		"BIN_TO_UUID(c.property_id) as property_id",
		"BIN_TO_UUID(c.event_id) as event_id",
		"BIN_TO_UUID(p.entity_id) as entity_id",
		"p.name as property_name",
		"c.current_data_type",
		"c.from_value", "c.to_value",
	)

	q = q.From("changes as c")
	q = q.Join("properties as p ON p.id = c.property_id")

	var expr bytes.Buffer
	var args []interface{}
	expr.WriteString("event_id IN (")
	for i := range ids {
		expr.WriteString("UUID_TO_BIN(?)")
		if i+1 < len(ids) {
			expr.WriteString(",")
		}

		args = append(args, ids[i])
	}
	expr.WriteString(")")

	return q.Where(expr.String(), args...).ToSql()
}

func createChangeQuery(c *change) (string, []interface{}, error) {
	if !validator.IsUUID4(c.PropertyID) {
		return "", nil, errors.New("change property id is not a valid uuid4")
	}

	if !validator.IsUUID4(c.EventID) {
		return "", nil, errors.New("change event id is not a valid uuid4")
	}

	return sq.Insert("changes").
		Columns("id", "property_id", "event_id", "current_data_type", "from_value", "to_value").
		Values(
			sq.Expr("UUID_TO_BIN(?)", c.ID),
			sq.Expr("UUID_TO_BIN(?)", c.PropertyID),
			sq.Expr("UUID_TO_BIN(?)", c.EventID),
			c.CurrentDataType,
			c.FromValue,
			c.ToValue,
		).ToSql()
}

func selectChangesByEventIDQuery(ID string) (string, []interface{}, error) {
	if !validator.IsUUID4(ID) {
		return "", nil, db.ErrInvalidUUID4
	}

	return sq.Select(
		"BIN_TO_UUID(c.id) as id",
		"BIN_TO_UUID(c.event_id) as event_id",
		"BIN_TO_UUID(c.property_id) as property_id",
		"BIN_TO_UUID(p.entity_id) as entity_id",
		"c.current_data_type",
		"p.name as property_name",
		"from_value",
		"to_value",
	).
		From("changes as c").
		Join("properties as p ON p.id = c.property_id").
		Where("event_id = UUID_TO_BIN(?)", ID).
		ToSql()
}
