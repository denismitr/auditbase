package mysql

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/denismitr/auditbase/utils/uuid"
	"github.com/denismitr/auditbase/utils/validator"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
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
}

func (p *property) ToModel() *model.Property {
	return &model.Property{
		ID:          p.ID,
		EntityID:    p.EntityID,
		Name:        p.Name,
		ChangeCount: p.ChangeCount,
	}
}

func (p *PropertyRepository) FirstByID(ID string) (*model.Property, error) {
	panic("implement me")
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

	countStmt, err :=  p.conn.Preparex(q.countSQL)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "could not prepare %s statement", q.countSQL)
	}

	var props []property
	var m meta

	if err := selectStmt.Select(&props, q.selectArgs...); err != nil {
		return nil, nil, errors.Wrap(err, "could not select properties")
	}

	if err := countStmt.Select(&m, q.selectArgs...); err != nil {
		return nil, nil, errors.Wrap(err, "could not select properties")
	}

	var result = make([]*model.Property, len(props))
	for i := range props {
		result[i] = props[i].ToModel()
	}

	return result, m.ToModel(pagination), nil
}

func (r *PropertyRepository) GetIDOrCreate(name, entityID string) (string, error) {
	var result string

	_, err := sq.Insert("properties").
		Columns("id", "name", "entity_id").
		Values(
			sq.Expr("UUID_TO_BIN(?)", r.uuid4.Generate()),
			name,
			sq.Expr("UUID_TO_BIN(?)", entityID),
		).RunWith(r.conn).Query()

	if err != nil {
		r.log.Error(err)
	}

	id := sq.
		Select("BIN_TO_UUID(id) as id").
		From("properties").
		Where(sq.Eq{"name": name}).
		Where("entity_id = UUID_TO_BIN(?)", entityID).
		Limit(1)

	sql, args, _ := id.ToSql()
	r.log.Debugf("%s -- %#v", sql, args)

	rows, err := id.RunWith(r.conn).Query()
	if err != nil {
		return result, errors.Wrapf(err, "could not select id from property with name %s and entityID %s", name, entityID)
	}

	for rows.Next() {
		if err := rows.Scan(&result); err != nil {
			return result, errors.Wrapf(err, "could not parse property ID for name %s and entityID %s", name, entityID)
		}

		return result, nil
	}

	return result, errors.Errorf("failed to create or retrieve property with name %s and entityId %s", name, entityID)
}

func createSelectPropertiesQuery(
	f *model.Filter,
	sort *model.Sort,
	pg *model.Pagination,
) (*selectQuery, error) {
	sQ := sq.Select(
		"BIN_TO_UUID(p.id) as id",
			"BIN_TO_UUID(p.entity_id) as entity_id",
			"p.name").From("properties as p")

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

	if sort.Empty() {
		sQ = sQ.OrderBy("id ASC")
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


