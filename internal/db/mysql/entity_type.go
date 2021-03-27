package mysql

import (
	"context"
	"database/sql"
	sq "github.com/Masterminds/squirrel"
	"github.com/denismitr/auditbase/internal/db"
	"github.com/denismitr/auditbase/internal/model"
	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"time"
	"unicode/utf8"
)

type entityTypeRecord struct {
	ID          int    `db:"id"`
	Name        string    `db:"name"`
	ServiceID   int    `db:"service_id"`
	Description string    `db:"description"`
	EntitiesCnt int       `db:"entities_cnt"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type EntityTypeRepository struct {
	*Tx
}

// static check of correct interface implementation
var _ db.EntityTypeRepository = (*EntityTypeRepository)(nil)

func (r *EntityTypeRepository) Select(
	ctx context.Context,
	cursor *db.Cursor,
	filter *db.Filter,
) (*model.EntityTypeCollection, error) {
	var entityTypes []entityTypeRecord

	sQ, err := selectEntityTypesQuery(cursor, filter)
	if err != nil {
		return nil, err
	}

	stmt, err := r.mysqlTx.Preparex(sQ.selectSQL)
	if err != nil {
		return nil, errors.Wrap(err, "could not prepare selectSql to select entityTypes")
	}

	cntStmt, err := r.mysqlTx.Preparex(sQ.countSQL)
	if err != nil {
		return nil, errors.Wrap(err, "could not prepare selectSql to count entityTypes")
	}

	defer func() { _ = stmt.Close() }()
	defer func() { _ = cntStmt.Close() }()

	if err := stmt.SelectContext(ctx, &entityTypes, sQ.selectArgs...); err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, db.ErrNotFound
		default:
			return nil, errors.Wrap(err, "could not execute selectSql to select entityTypes")
		}
	}

	var cnt int
	if err := cntStmt.Get(ctx, &cnt); err != nil {
		return nil, errors.Wrap(err, "could not execute selectSql to count entityTypes")
	}

	return mapEntityTypesToCollection(entityTypes, cnt, cursor.Page, cursor.PerPage), nil
}

func (r *EntityTypeRepository) Create(ctx context.Context, e *model.EntityType) (*model.EntityType, error) {
	q, args, err := createEntityTypeQuery(e.ID, e.ServiceID, e.Name, e.Description, e.IsActor)
	if err != nil {
		panic(errors.Wrap(err, "how could createEntityTypeQuery func fail?"))
	}

	stmt, err := r.mysqlTx.Preparex(q)
	if err != nil {
		return nil, errors.Wrap(err, "could not prepare tx to create entity type")
	}

	result, err := stmt.ExecContext(ctx, args...);
	if err != nil {
		return nil, errors.Wrapf(err, "could not create new entity type with name %s", e.Name)
	}

	newID, err := result.LastInsertId()
	if err != nil {
		return nil, errors.Wrapf(err, "could not retrieve last insert ID for create entity type [%s]", e.Name)
	}

	return r.FirstByID(ctx, model.ID(newID))
}

func (r *EntityTypeRepository) FirstOrCreateByNameAndServiceID(
	ctx context.Context,
	name string,
	serviceID model.ID,
) (*model.EntityType, error) {
	if ent, err := r.FirstByNameAndServiceID(ctx, name, serviceID); err == nil {
		return ent, nil
	} else {
		switch err {
		case db.ErrNotFound:
			r.lg.Debugf(err.Error())
		default:
			return nil, err
		}
	}

	newEntity := &model.EntityType{
		Name:      name,
		ServiceID: serviceID,
	}

	created, err := r.Create(ctx, newEntity)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"entityTypeRecord with name [%s] and service ID [%d] does not exist and could not be created",
			name,
			serviceID,
		)
	}

	return created, nil
}

func (r *EntityTypeRepository) FirstByID(ctx context.Context, ID model.ID) (*model.EntityType, error) {
	return firstEntityTypeByID(ctx, r.mysqlTx, ID)
}

func firstEntityTypeByID(ctx context.Context, tx *sqlx.Tx, ID model.ID) (*model.EntityType, error) {
	q, args, err := firstEntityTypeByIDQuery(ID)
	if err != nil {
		return nil, err
	}

	ent := entityTypeRecord{}

	stmt, err := tx.Preparex(q)
	if err != nil {
		return nil, errors.Wrapf(err, "could not prepare selectSql statement %s to get entityRecord by id", q)
	}

	defer func() { _ = stmt.Close() }()

	if err := stmt.GetContext(ctx, &ent, args...); err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, db.ErrNotFound
		default:
			return nil, errors.Wrapf(err, "could not find entities with ID %d", ID)
		}
	}

	return mapEntityTypeRecordToModel(ent), nil
}

func (r *EntityTypeRepository) FirstByNameAndServiceID(
	ctx context.Context,
	name string,
	serviceID model.ID,
) (*model.EntityType, error) {
	q, args, err := firstEntityTypeByNameAndServiceIDQuery(name, serviceID)
	if err != nil {
		panic("how could firstEntityTypeByNameAndServiceIDQuery func fail?")
	}

	ent := entityTypeRecord{}

	stmt, err := r.mysqlTx.Preparex(q)
	if err != nil {
		return nil, errors.Wrapf(err, "could not prepare selectSql statement %s to get entityRecord by name and service ID", q)
	}

	defer func() { _ = stmt.Close() }()

	if err := stmt.GetContext(ctx, &ent, args...); err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, db.ErrNotFound
		default:
			return nil, errors.Wrapf(err, "could not find entity with name %s and serviceID %d", name, serviceID)
		}
	}

	return mapEntityTypeRecordToModel(ent), nil
}

func firstEntityTypeByIDQuery(ID model.ID) (string, []interface{}, error) {
	dialect := goqu.Dialect(MySQL8)

	q := dialect.Select(
		goqu.L("`et`.`id`").As("id"),
		goqu.L("`et`.`service_id`").As("service_id"),
		goqu.I("et.name").As("name"),
		goqu.I("et.description").As("description"),
		goqu.I("et.created_at").As("created_at"),
		goqu.I("et.updated_at").As("updated_at"),
		goqu.L("count(distinct `e`.`id`)").As("entities_cnt"),
	).
		From(goqu.T("entity_types").As("et")).
		LeftJoin(goqu.T("entities").As("e"), goqu.On(goqu.Ex{"e.entity_type_id": goqu.I("et.id")})).
		Where(goqu.L("`et`.`id` = ?", int(ID))).
		GroupBy(
			goqu.I("et.id"),
			goqu.I("et.service_id"),
			goqu.I("et.name"),
			goqu.I("et.description"),
			goqu.I("et.created_at"),
			goqu.I("et.updated_at"),
		).Limit(1)

	return q.Prepared(true).ToSQL()
}

func selectEntityTypesQuery(c *db.Cursor, f *db.Filter) (*selectQuery, error) {
	countQ := sq.Select("count(*) as cnt").From("entity_types")

	q := sq.Select(
		"id",
		"service_id",
		"name",
		"description",
		"created_at",
		"updated_at",
		"count(distinct entities.id) as entities_cnt",
	).From("entities").LeftJoin("entities on entities.entity_type_id = entity_types.id")

	if f.Has("entityTypeID") {
		countQ = countQ.Where(`service_id = ?`, f.MustInt("entityTypeID"))
		q = q.Where(`service_id = ?`, f.MustInt("entityTypeID"))
	}

	if f.Has("name") {
		countQ = countQ.Where(`name = ?`, f.MustString("name"))
		q = q.Where(`name = ?`, f.MustString("name"))
	}

	q = q.GroupBy("id", "service_id", "name", "description", "created_at", "updated_at")

	if c != nil {
		if c.Sort.Has("name") {
			q = q.OrderByClause("name ?", c.Sort.GetOrDefault("name", db.ASCOrder).String())
		} else {
			q = q.OrderBy("updated_at ?", db.DESCOrder.String())
		}

		q = q.Limit(uint64(c.PerPage))
		q = q.Offset(uint64(c.Offset()))
	} else {
		q = q.Offset(0)
		q = q.OrderBy("updated_at ?", db.DESCOrder.String())
	}

	sQ := selectQuery{}
	if queryStr, args, err := q.ToSql(); err != nil {
		return nil, errors.Wrap(err, "invalid select SQL for entities")
	} else {
		sQ.selectSQL = queryStr
		sQ.selectArgs = args
	}

	if queryStr, args, err := countQ.ToSql(); err != nil {
		return nil, errors.Wrap(err, "invalid count SQL for entities")
	} else {
		sQ.countSQL = queryStr
		sQ.countArgs = args
	}

	return &sQ, nil
}

func createEntityTypeQuery(ID, serviceID model.ID, name, description string, isActor bool) (string, []interface{}, error) {
	if !utf8.ValidString(name) {
		panic("how could name not be a valid ut8 string")
	}

	dialect := goqu.Dialect(MySQL8)

	q := dialect.Insert(goqu.T("entity_types")).Cols(
		"service_id", "name", "description", "is_actor",
	).Vals(
		goqu.Vals{
			goqu.L("?", int(serviceID)), // fixme
			name, description, isActor,
		},
	)

	return q.Prepared(true).ToSQL()
}

func firstEntityTypeByNameAndServiceIDQuery(name string, serviceID model.ID) (string, []interface{}, error) {
	if name == "" {
		panic("how can name be empty?")
	}

	dialect := goqu.Dialect(MySQL8)

	q := dialect.Select(
		goqu.L("`et`.`id`").As("id"),
		goqu.L("`et`.`service_id`").As("service_id"),
		goqu.I("et.name").As("name"),
		goqu.I("et.description").As("description"),
		goqu.I("et.created_at").As("created_at"),
		goqu.I("et.updated_at").As("updated_at"),
		goqu.L("count(distinct `e`.`id`)").As("entities_cnt"),
	).
		From(goqu.T("entity_types").As("et")).
		LeftJoin(goqu.T("entities").As("e"), goqu.On(goqu.Ex{"e.entity_type_id": goqu.I("et.id")})).
		Where(goqu.L("`et`.`service_id` = ?", int(serviceID))).
		Where(goqu.Ex{"name": name}).
		GroupBy(
			goqu.I("et.id"),
			goqu.I("et.service_id"),
			goqu.I("et.name"),
			goqu.I("et.description"),
			goqu.I("et.created_at"),
			goqu.I("et.updated_at"),
		).Limit(1)

	return q.Prepared(true).ToSQL()
}
