package mysql

import (
	"context"
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/denismitr/auditbase/internal/db"
	"github.com/denismitr/auditbase/internal/model"
	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"time"
)

type entityRecord struct {
	ID               int       `db:"id"`
	ExternalID       string    `db:"external_id"`
	EntityTypeID     int       `db:"entity_type_id"`
	IsActor          bool      `db:"is_actor"`
	TargetActionsCnt int       `db:"target_actions_cnt"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

type entityRecordAllJoined struct {
	EntityID              int       `db:"entity_id"`
	EntityExternalID      string    `db:"entity_external_id"`
	EntityTypeID          int       `db:"entity_type_id"`
	IsActor               bool      `db:"is_actor"`
	TargetActionsCnt      int       `db:"target_actions_cnt"`
	EntityCreatedAt       time.Time `db:"entity_created_at"`
	EntityUpdatedAt       time.Time `db:"entity_updated_at"`
	EntityTypeName        string    `db:"entity_type_name"`
	ServiceID             int       `db:"service_id"`
	EntityTypeDescription string    `db:"entity_type_description"`
	EntityTypeCreatedAt   time.Time `db:"entity_type_created_at"`
	EntityTypeUpdatedAt   time.Time `db:"entity_type_updated_at"`
	ServiceName           string    `db:"service_name"`
	ServiceDescription    string    `db:"service_description"`
	ServiceCreatedAt      time.Time `db:"service_created_at"`
	ServiceUpdatedAt      time.Time `db:"service_updated_at"`
}

type EntityRepository struct {
	*Tx
}

// static check of correct interface implementation
var _ db.EntityRepository = (*EntityRepository)(nil)

// Select all entities
func (r *EntityRepository) Select(
	ctx context.Context,
	cursor *db.Cursor,
	filter *db.Filter,
) (*model.EntityCollection, error) {
	var entities []entityRecord

	sQ, err := selectEntitiesQuery(cursor, filter)
	if err != nil {
		return nil, err
	}

	stmt, err := r.mysqlTx.Preparex(sQ.selectSQL)
	if err != nil {
		return nil, errors.Wrap(err, "could not prepare selectSql to select entities")
	}

	cntStmt, err := r.mysqlTx.Preparex(sQ.countSQL)
	if err != nil {
		return nil, errors.Wrap(err, "could not prepare selectSql to count entities")
	}

	defer func() { _ = stmt.Close() }()
	defer func() { _ = cntStmt.Close() }()

	if err := stmt.SelectContext(ctx, &entities, sQ.selectArgs...); err != nil {
		return nil, errors.Wrap(err, "could not execute selectSql to select entities")
	}

	var cnt int
	if err := cntStmt.Get(ctx, &cnt); err != nil {
		return nil, errors.Wrap(err, "could not execute selectSql to count entities")
	}

	return mapEntitiesToCollection(entities, cnt, cursor.Page, cursor.PerPage), nil
}

// MakeNewActions an entities
func (r *EntityRepository) Create(ctx context.Context, e *model.Entity) (*model.Entity, error) {
	q, args, err := createEntityQuery(e.ID, e.EntityTypeID, e.ExternalID)
	if err != nil {
		panic(errors.Wrap(err, "how could query builder fail?"))
	}

	stmt, err := r.mysqlTx.PreparexContext(ctx, q)
	if err != nil {
		return nil, errors.Wrap(err, "could not prepare tx to create entity")
	}

	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "could not create new entities with externalID %s", e.ExternalID)
	}

	newID, err := result.LastInsertId()
	if err != nil {
		return nil, errors.Wrapf(err, "could not retrieve new ID for created entity [%s]", e.ExternalID)
	}

	return r.FirstByID(ctx, model.ID(newID))
}

func (r *EntityRepository) FirstByExternalIDAndTypeID(
	ctx context.Context,
	externalID string,
	entityTypeID model.ID,
) (*model.Entity, error) {
	q, args, err := firstByExternalIDAndTypeIDQuery(externalID, entityTypeID)
	if err != nil {
		panic(fmt.Sprintf("could not build firstByExternalIDAndTypeIDQuery query for %s - %d", externalID, entityTypeID))
	}

	stmt, err := r.mysqlTx.PreparexContext(ctx, q)
	if err != nil {
		return nil, errors.Wrapf(err, "could not prepare query %s", q)
	}

	ent := entityRecord{}

	if err := stmt.GetContext(ctx, &ent, args...); err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, db.ErrNotFound
		default:
			return nil, errors.Wrapf(
				err,
				"could not find entities with externalID [%s] and entityTypeID [%d]",
				externalID,
				entityTypeID,
			)
		}
	}

	return mapEntityRecordToModel(ent), nil
}

func (r *EntityRepository) FirstByID(ctx context.Context, ID model.ID) (*model.Entity, error) {
	return firstEntityByID(ctx, r.mysqlTx, ID)
}

func (r *EntityRepository) FirstByIDWithEntityType(ctx context.Context, ID model.ID) (*model.Entity, error) {
	entity, err := firstEntityByID(ctx, r.mysqlTx, ID)
	if err != nil {
		return nil, err
	}

	entityType, err := firstEntityTypeByID(ctx, r.mysqlTx, entity.EntityTypeID)
	if err != nil {
		return nil, errors.Wrap(err, "could not join entity type to entity")
	}

	entity.EntityType = entityType

	return entity, nil
}

// FirstOrCreateByNameAndService ...
func (r *EntityRepository) FirstOrCreateByExternalIDAndEntityTypeID(
	ctx context.Context,
	externalID string,
	entityTypeID model.ID,
) (*model.Entity, error) {
	if ent, err := r.FirstByExternalIDAndTypeID(ctx, externalID, entityTypeID); err == nil {
		return ent, nil
	} else {
		switch err {
		case db.ErrNotFound:
			r.lg.Debugf(err.Error())
		default:
			return nil, err
		}
	}

	ent := &model.Entity{
		EntityTypeID: entityTypeID,
		ExternalID:   externalID,
	}

	created, err := r.Create(ctx, ent)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"entity with external ID [%s] and entity type ID [%d] does not exist and could not be created",
			externalID,
			entityTypeID,
		)
	}

	return created, nil
}

func firstEntityByID(ctx context.Context, tx *sqlx.Tx, ID model.ID) (*model.Entity, error) {
	q, args, err := firstEntityByIDQuery(ID)
	if err != nil {
		return nil, err
	}

	ent := entityRecordAllJoined{}

	stmt, err := tx.Preparex(q)
	if err != nil {
		return nil, errors.Wrapf(err, "could not prepare selectSql statement %s to get entityRecord by id", q)
	}

	defer func() { _ = stmt.Close() }()

	if err := stmt.GetContext(ctx, &ent, args...); err != nil {
		return nil, errors.Wrapf(err, "could not find entities with ID %d", ID)
	}

	return mapEntityRecordAllJoinedToModel(ent), nil
}

func selectEntitiesQuery(c *db.Cursor, f *db.Filter) (*selectQuery, error) {
	dialect := goqu.Dialect(MySQL8)

	countQ := dialect.Select(goqu.L("count(*)").As("cnt")).From(goqu.I("entities").As("e"))

	q := dialect.Select(
		goqu.I("e.id"),
		goqu.I("e.entity_type_id"),
		goqu.I("e.external_id"),
		goqu.I("e.created_at"),
		goqu.I("e.updated_at"),
	).From(goqu.I("entities").As("e"))

	if f.Has("entityTypeId") {
		countQ = countQ.Where(goqu.I(`e.entity_type_id`).Eq(f.MustInt("entityTypeId")))
		q = q.Where(goqu.I(`e.entity_type_id`).Eq(f.MustInt("entityTypeId")))
	}

	if c.Sort.Has("externalId") {
		ind := goqu.I("e.external_id")
		order := c.Sort.GetOrDefault("externalId", db.ASCOrder)
		if order == db.ASCOrder {
			q = q.OrderAppend(ind.Asc())
		} else {
			q = q.OrderAppend(ind.Desc())
		}
	} else {
		ind := goqu.I("e.updated_at")
		order := c.Sort.GetOrDefault("updatedAt", db.DESCOrder)
		if order == db.DESCOrder {
			q = q.OrderAppend(ind.Desc())
		} else {
			q = q.OrderAppend(ind.Asc())
		}
	}

	q = q.Limit(c.PerPage)
	q = q.Offset(c.Offset())

	sQ := selectQuery{}
	if q, args, err := q.Prepared(true).ToSQL(); err != nil {
		return nil, errors.Wrap(err, "invalid select SQL for entities")
	} else {
		sQ.selectSQL = q
		sQ.selectArgs = args
	}

	if q, args, err := countQ.Prepared(true).ToSQL(); err != nil {
		return nil, errors.Wrap(err, "invalid count SQL for entities")
	} else {
		sQ.countSQL = q
		sQ.countArgs = args
	}

	return &sQ, nil
}

func firstEntityByIDQuery(ID model.ID) (string, []interface{}, error) {
	dialect := goqu.Dialect(MySQL8)

	return dialect.Select(
		goqu.I("e.id").As("entity_id"),
		goqu.I("e.entity_type_id").As("entity_type_id"),
		goqu.I("et.service_id").As("service_id"),
		goqu.I("e.external_id").As("entity_external_id"),
		goqu.I("et.name").As("entity_type_name"),
		goqu.I("et.description").As("entity_type_description"),
		goqu.I("ms.name").As("service_name"),
		goqu.I("ms.description").As("service_description"),
		goqu.I("e.created_at").As("entity_created_at"),
		goqu.I("et.created_at").As("entity_type_created_at"),
		goqu.I("ms.created_at").As("service_created_at"),
		goqu.I("e.updated_at").As("entity_updated_at"),
		goqu.I("et.updated_at").As("entity_type_updated_at"),
		goqu.I("ms.updated_at").As("service_updated_at"),
	).From(goqu.T("entities").As("e")).InnerJoin(
		goqu.T("entity_types").As("et"),
		goqu.On(goqu.Ex{"e.entity_type_id": goqu.I("et.id")}),
	).InnerJoin(
		goqu.T("microservices").As("ms"),
		goqu.On(goqu.Ex{"et.service_id": goqu.I("ms.id")}),
	).Where(
		goqu.L("`e`.`id` = ?", int(ID)), // fixme
	).Limit(1).Prepared(true).ToSQL()
}

func firstByExternalIDAndTypeIDQuery(externalID string, entityTypeID model.ID) (string, []interface{}, error) {
	if externalID == "" {
		panic("how can external id be empty?")
	}

	return sq.Select(
		"id",
		"entity_type_id",
		"external_id", "created_at", "updated_at",
	).
		From("entities").
		Where("entity_type_id = ?", int(entityTypeID)).
		Where("external_id = ?", externalID).
		GroupBy("id", "entity_type_id", "external_id", "created_at", "updated_at").
		Limit(1).
		ToSql()
}

func createEntityQuery(id, entityTypeID model.ID, externalID string) (string, []interface{}, error) {
	if externalID == "" {
		panic("how can external id be empty?")
	}

	return sq.Insert("entities").
		Columns("external_id", "entity_type_id").
		Values(externalID, sq.Expr("?", int(entityTypeID))).
		ToSql() // fixme: refactor to goqu
}
