package mysql

import (
	"context"
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/denismitr/auditbase/db"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/validator"
	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"time"
)

type entityRecord struct {
	ID               string    `db:"id"`
	ExternalID       string    `db:"external_id"`
	EntityTypeID     string    `db:"entity_type_id"`
	IsActor          bool      `db:"is_actor"`
	TargetActionsCnt int       `db:"target_actions_cnt"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

type entityRecordAllJoined struct {
	EntityID              string    `db:"entity_id"`
	EntityExternalID      string    `db:"entity_external_id"`
	EntityTypeID          string    `db:"entity_type_id"`
	IsActor               bool      `db:"is_actor"`
	TargetActionsCnt      int       `db:"target_actions_cnt"`
	EntityCreatedAt       time.Time `db:"entity_created_at"`
	EntityUpdatedAt       time.Time `db:"entity_updated_at"`
	EntityTypeName        string    `db:"entity_type_name"`
	ServiceID             string    `db:"service_id"`
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
		return nil, errors.Wrap(err, "could not prepare sql to select entities")
	}

	cntStmt, err := r.mysqlTx.Preparex(sQ.countSQL)
	if err != nil {
		return nil, errors.Wrap(err, "could not prepare sql to count entities")
	}

	defer func() { _ = stmt.Close() }()
	defer func() { _ = cntStmt.Close() }()

	if err := stmt.SelectContext(ctx, &entities, sQ.selectArgs...); err != nil {
		return nil, errors.Wrap(err, "could not execute sql to select entities")
	}

	var cnt int
	if err := cntStmt.Get(ctx, &cnt); err != nil {
		return nil, errors.Wrap(err, "could not execute sql to count entities")
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

	if _, err := stmt.ExecContext(ctx, args...); err != nil {
		return nil, errors.Wrapf(err, "could not create new entities with externalID %s", e.ExternalID)
	}

	return r.FirstByID(ctx, e.ID)
}

func (r *EntityRepository) FirstByExternalIDAndTypeID(
	ctx context.Context,
	externalID string,
	entityTypeID model.ID,
) (*model.Entity, error) {
	q, args, err := firstByExternalIDAndTypeIDQuery(externalID, entityTypeID)
	if err != nil {
		panic(fmt.Sprintf("could not build firstByExternalIDAndTypeIDQuery query for %s - %s", externalID, entityTypeID.String()))
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
				"could not find entities with externalID [%s] and entityTypeID [%s]",
				externalID,
				entityTypeID.String(),
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
		ID:           model.ID(r.uuid4.Generate()),
		EntityTypeID: entityTypeID,
		ExternalID:   externalID,
	}

	created, err := r.Create(ctx, ent)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"entity with external ID [%s] and entity type ID [%s] does not exist and could not be created",
			externalID,
			entityTypeID,
		)
	}

	return created, nil
}

func firstEntityByID(ctx context.Context, tx *sqlx.Tx, ID model.ID) (*model.Entity, error) {
	sql, args, err := firstEntityByIDQuery(ID)
	if err != nil {
		return nil, err
	}

	ent := entityRecordAllJoined{}

	stmt, err := tx.Preparex(sql)
	if err != nil {
		return nil, errors.Wrapf(err, "could not prepare sql statement %s to get entityRecord by id", sql)
	}

	defer func() { _ = stmt.Close() }()

	if err := stmt.GetContext(ctx, &ent, args...); err != nil {
		return nil, errors.Wrapf(err, "could not find entities with ID %s", ID)
	}

	return mapEntityRecordAllJoinedToModel(ent), nil
}

func selectEntitiesQuery(c *db.Cursor, f *db.Filter) (*selectQuery, error) {
	countQ := sq.Select("count(*) as cnt").From("entities")

	q := sq.Select(
		"bin_to_uuid(id) as id",
		"bin_to_uuid(entity_type_id) as entity_type_id",
		"external_id",
		"created_at",
		"updated_at",
	).From("entities")

	if f.Has("entityTypeId") {
		countQ = countQ.Where(`entity_type_id = uuid_to_bin(?)`, f.StringOrDefault("entityTypeId", ""))
		q = q.Where(`entity_type_id = uuid_to_bin(?)`, f.StringOrDefault("entityTypeId", ""))
	}

	if c.Sort.Has("externalId") {
		q = q.OrderByClause("external_id ?", c.Sort.GetOrDefault("externalId", db.ASCOrder).String())
	} else {
		q = q.OrderBy("updated_at ?", c.Sort.GetOrDefault("updatedAt", db.DESCOrder).String())
	}

	q = q.GroupBy("id", "entity_type_id", "external_id", "created_at", "updated_at")
	q = q.Limit(uint64(c.PerPage))
	q = q.Offset(uint64(c.Offset()))

	sQ := selectQuery{}
	if sql, args, err := q.ToSql(); err != nil {
		return nil, errors.Wrap(err, "invalid select SQL for entities")
	} else {
		sQ.selectSQL = sql
		sQ.selectArgs = args
	}

	if sql, args, err := countQ.ToSql(); err != nil {
		return nil, errors.Wrap(err, "invalid count SQL for entities")
	} else {
		sQ.countSQL = sql
		sQ.countArgs = args
	}

	return &sQ, nil
}

func firstEntityByIDQuery(ID model.ID) (string, []interface{}, error) {
	if !validator.IsUUID4(ID.String()) {
		panic(fmt.Sprintf("%s is not a valid entityRecord id (UUID4)", ID))
	}

	dialect := goqu.Dialect(MySQL8)

	return dialect.Select(
		goqu.L("bin_to_uuid(`e`.`id`)").As("entity_id"),
		goqu.L("bin_to_uuid(`e`.`entity_type_id`)").As("entity_type_id"),
		goqu.L("bin_to_uuid(`et`.`service_id`)").As("service_id"),
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
		goqu.L("`e`.`id` = uuid_to_bin(?)", ID.String()),
	).Limit(1).Prepared(true).ToSQL()
}

func firstByExternalIDAndTypeIDQuery(externalID string, entityTypeID model.ID) (string, []interface{}, error) {
	if !validator.IsUUID4(entityTypeID.String()) {
		panic(fmt.Sprintf("entity type ID [%s] is not a valid UUID4", entityTypeID))
	}

	if externalID == "" {
		panic("how can external id be empty?")
	}

	return sq.Select(
		"bin_to_uuid(id) as id",
		"bin_to_uuid(entity_type_id) as entity_type_id",
		"external_id", "created_at", "updated_at",
	).
		From("entities").
		Where("entity_type_id = uuid_to_bin(?)", entityTypeID.String()).
		Where("external_id = ?", externalID).
		GroupBy("id", "entity_type_id", "external_id", "created_at", "updated_at").
		Limit(1).
		ToSql()
}

func createEntityQuery(id, entityTypeID model.ID, externalID string) (string, []interface{}, error) {
	if !validator.IsUUID4(id.String()) {
		panic(fmt.Sprintf("entity ID [%s] is not a valid UUID4", entityTypeID))
	}

	if !validator.IsUUID4(entityTypeID.String()) {
		panic(fmt.Sprintf("entity type ID [%s] is not a valid UUID4", entityTypeID))
	}

	if externalID == "" {
		panic("how can external id be empty?")
	}

	return sq.Insert("entities").
		Columns("id", "external_id", "entity_type_id").
		Values(sq.Expr("uuid_to_bin(?)", id), externalID, sq.Expr("uuid_to_bin(?)", entityTypeID)).
		ToSql()
}
