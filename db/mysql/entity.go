package mysql

import (
	"context"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/denismitr/auditbase/db"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/validator"
	"github.com/pkg/errors"
)

type entityRecord struct {
	ID               string `db:"id"`
	ExternalID       string `db:"external_id"`
	EntityTypeID     string `db:"entity_type_id"`
	IsActor          bool   `db:"is_actor"`
	TargetActionsCnt int    `db:"target_actions_cnt"`
	CreatedAt        string `db:"created_at"`
	UpdatedAt        string `db:"updated_at"`
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

// Create an entities
func (r *EntityRepository) Create(ctx context.Context, e *model.Entity) (*model.Entity, error) {
	q, args, err := createEntityQuery(e.ID, e.EntityTypeID, e.ExternalID, e.IsActor)
	if err != nil {
		panic(errors.Wrap(err, "how could query builder fail?"))
	}

	stmt, err := r.mysqlTx.PreparexContext(ctx, q)
	if err != nil {
		return nil, errors.Wrap(err, "could not prepare tx to create entityRecord")
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
		return nil, errors.Wrapf(
			err,
			"could not find entities with externalID [%s] and entityTypeID [%s]",
			externalID,
			entityTypeID.String(),
		)
	}

	return mapEntityRecordToModel(ent), nil
}

func (r *EntityRepository) FirstByID(ctx context.Context, ID model.ID) (*model.Entity, error) {
	sql, args, err := firstEntityByIDQuery(ID)
	if err != nil {
		return nil, err
	}

	ent := entityRecord{}

	stmt, err := r.mysqlTx.Preparex(sql)
	if err != nil {
		return nil, errors.Errorf("could not prepare sql statement %s to get entityRecord by id", sql)
	}

	defer func() { _ = stmt.Close() }()

	if err := stmt.GetContext(ctx, &ent, args...); err != nil {
		return nil, errors.Wrapf(err, "could not find entities with ID %s", ID)
	}

	return mapEntityRecordToModel(ent), nil
}

// FirstOrCreateByNameAndService ...
func (r *EntityRepository) FirstOrCreateByExternalIDAndEntityTypeID(
	ctx context.Context,
	externalID string,
	entityTypeID model.ID,
	isActor bool,
) (*model.Entity, error) {
	if ent, err := r.FirstByExternalIDAndTypeID(ctx, externalID, entityTypeID); err == nil {
		return ent, nil
	} else {
		r.lg.Debugf(err.Error())
	}

	ent := &model.Entity{
		ID:           model.ID(r.uuid4.Generate()),
		EntityTypeID: entityTypeID,
		ExternalID:   externalID,
		IsActor:      isActor,
	}

	created, err := r.Create(ctx, ent);
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

func selectEntitiesQuery(c *db.Cursor, f *db.Filter) (*selectQuery, error) {
	countQ := sq.Select("count(*) as cnt").From("entities")

	q := sq.Select(
		"bin_to_uuid(id) as id",
		"bin_to_uuid(entity_type_id) as entity_type_id",
		"external_id",
		"is_actor",
		"created_at",
		"updated_at",
		"count(distinct actions.id) as target_actions_cnt",
	).From("entities").LeftJoin("actions on actions.target_entity_id = entities.id")

	if f.Has("entityTypeId") {
		countQ = countQ.Where(`entity_type_id = uuid_to_bin(?)`, f.StringOrDefault("entityTypeId", ""))
		q = q.Where(`entity_type_id = uuid_to_bin(?)`, f.StringOrDefault("entityTypeId", ""))
	}

	if c.Sort.Has("externalId") {
		q = q.OrderByClause("external_id ?", c.Sort.GetOrDefault("externalId", db.ASCOrder).String())
	} else {
		q = q.OrderBy("updated_at ?", c.Sort.GetOrDefault("updatedAt", db.DESCOrder).String())
	}

	q = q.GroupBy("id", "entity_type_id", "external_id", "is_actor", "created_at", "updated_at")
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

	return sq.Select(
		"bin_to_uuid(id) as id",
		"count(distinct actions.id) as target_actions_cnt",
		"bin_to_uuid(entity_type_id) as entity_type_id",
		"external_id", "is_active", "created_at", "updated_at").
		From("entities").
		LeftJoin("actions on actions.target_entity_id = entities.id").
		Where("id = uuid_to_bin(?)", ID.String()).
		GroupBy("id", "entity_type_id", "external_id", "is_active", "created_at", "updated_at").
		Limit(1).
		ToSql()
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
		"count(distinct actions.id) as target_actions_cnt",
		"bin_to_uuid(entity_type_id) as entity_type_id",
		"external_id", "is_active", "created_at", "updated_at",
	).
		From("entities").
		LeftJoin("actions on actions.target_entity_id = entities.id").
		Where("entity_type_id = uuid_to_bin(?)", entityTypeID.String()).
		Where("external_id = ?", externalID).
		GroupBy("id", "entity_type_id", "external_id", "is_active", "created_at", "updated_at").
		Limit(1).
		ToSql()
}

func createEntityQuery(id, entityTypeID model.ID, externalID string, isActor bool) (string, []interface{}, error) {
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
		Columns("id", "external_id", "entity_type_id", "is_actor").
		Values(sq.Expr("uuid_to_bin(?)", id), externalID, sq.Expr("uuid_to_bin(?)", entityTypeID), isActor).
		ToSql()
}

