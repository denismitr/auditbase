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

type entityTypeRecord struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	ServiceID   string `db:"service_id"`
	Description string `db:"description"`
	EntitiesCnt int    `db:"entities_cnt"`
	CreatedAt   string `db:"created_at"`
	UpdatedAt   string `db:"updated_at"`
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
		return nil, errors.Wrap(err, "could not prepare sql to select entityTypes")
	}

	cntStmt, err := r.mysqlTx.Preparex(sQ.countSQL)
	if err != nil {
		return nil, errors.Wrap(err, "could not prepare sql to count entityTypes")
	}

	defer func() { _ = stmt.Close() }()
	defer func() { _ = cntStmt.Close() }()

	if err := stmt.SelectContext(ctx, &entityTypes, sQ.selectArgs...); err != nil {
		return nil, errors.Wrap(err, "could not execute sql to select entityTypes")
	}

	var cnt int
	if err := cntStmt.Get(ctx, &cnt); err != nil {
		return nil, errors.Wrap(err, "could not execute sql to count entityTypes")
	}

	return mapEntityTypesToCollection(entityTypes, cnt, cursor.Page, cursor.PerPage), nil
}

func (r *EntityTypeRepository) Create(ctx context.Context, e *model.EntityType) (*model.EntityType, error) {
	q, args, err := createEntityTypeQuery(e.ID, e.ServiceID, e.Name, e.Description)
	if err != nil {
		panic(errors.Wrap(err, "how could createEntityTypeQuery func fail?"))
	}

	stmt, err := r.mysqlTx.Preparex(q)
	if err != nil {
		return nil, errors.Wrap(err, "could not prepare tx to create entityRecord")
	}

	if _, err := stmt.ExecContext(ctx, args...); err != nil {
		return nil, errors.Wrapf(err, "could not create new entity type with name %s", e.Name)
	}

	return r.FirstByID(ctx, e.ID)
}

func (r *EntityTypeRepository) FirstOrCreateByNameAndServiceID(
	ctx context.Context,
	name string,
	serviceID model.ID,
) (*model.EntityType, error) {
	if ent, err := r.FirstByNameAndServiceID(ctx, name, serviceID); err == nil {
		return ent, nil
	} else {
		r.lg.Debugf(err.Error())
	}

	newEntity := &model.EntityType{
		ID:           model.ID(r.uuid4.Generate()),
		Name: name,
		ServiceID:   serviceID,
	}

	created, err := r.Create(ctx, newEntity);
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"entityTypeRecord with name [%s] and service ID [%s] does not exist and could not be created",
			name,
			serviceID.String(),
		)
	}

	return created, nil
}

func (r *EntityTypeRepository) FirstByID(ctx context.Context, ID model.ID) (*model.EntityType, error) {
	if !validator.IsUUID4(ID.String()) {
		panic("how can id be invalid?")
	}

	sql, args, err := firstEntityTypeByIDQuery(ID.String())
	if err != nil {
		return nil, err
	}

	ent := entityTypeRecord{}

	stmt, err := r.mysqlTx.Preparex(sql)
	if err != nil {
		return nil, errors.Errorf("could not prepare sql statement %s to get entityRecord by id", sql)
	}

	defer func() { _ = stmt.Close() }()

	if err := stmt.GetContext(ctx, &ent, args...); err != nil {
		return nil, errors.Wrapf(err, "could not find entities with ID %s", ID)
	}

	return mapEntityTypeRecordToModel(ent), nil
}

func (r *EntityTypeRepository) FirstByNameAndServiceID(
	ctx context.Context,
	name string,
	serviceID model.ID,
) (*model.EntityType, error) {
	sql, args, err := firstEntityTypeByNameAndServiceIDQuery(name, serviceID)
	if err != nil {
		panic("how could irstEntityTypeByNameAndServiceIDQuery func fail?")
	}

	ent := entityTypeRecord{}

	stmt, err := r.mysqlTx.Preparex(sql)
	if err != nil {
		return nil, errors.Errorf("could not prepare sql statement %s to get entityRecord by name and service ID", sql)
	}

	defer func() { _ = stmt.Close() }()

	if err := stmt.GetContext(ctx, &ent, args...); err != nil {
		return nil, errors.Wrapf(err, "could not find entity with name %s and serviceID %s", name, serviceID.String())
	}

	return mapEntityTypeRecordToModel(ent), nil
}

func firstEntityTypeByIDQuery(ID string) (string, []interface{}, error) {
	if !validator.IsUUID4(ID) {
		panic(fmt.Sprintf("%s is not a valid entityRecord id (UUID4)", ID))
	}

	return sq.Select(
		"bin_to_uuid(id) as id",
		"bin_to_uuid(service_id) as service_id",
		"name",
		"description",
		"created_at",
		"updated_at",
		"count(distinct entities.id) as entities_cnt",
	).
		From("entities").
		LeftJoin("entities on entities.entity_type_id = entity_types.id").
		Where(`id = uuid_to_bin(?)`, ID).
		GroupBy("id", "service_id", "name", "description", "created_at", "updated_at").
		Limit(1).
		ToSql()
}

func selectEntityTypesQuery(c *db.Cursor, f *db.Filter) (*selectQuery, error) {
	countQ := sq.Select("count(*) as cnt").From("entity_types")

	q := sq.Select(
		"bin_to_uuid(id) as id",
		"bin_to_uuid(service_id) as service_id",
		"name",
		"description",
		"created_at",
		"updated_at",
		"count(distinct entities.id) as entities_cnt",
	).From("entities").LeftJoin("entities on entities.entity_type_id = entity_types.id")

	if f.Has("serviceId") {
		countQ = countQ.Where(`service_id = uuid_to_bin(?)`, f.StringOrDefault("serviceId", ""))
		q = q.Where(`service_id = uuid_to_bin(?)`, f.StringOrDefault("serviceId", ""))
	}

	if f.Has("name") {
		countQ = countQ.Where(`name = ?`, f.StringOrDefault("name", ""))
		q = q.Where(`name = ?`, f.StringOrDefault("name", ""))
	}

	if c.Sort.Has("name") {
		q = q.OrderByClause("name ?", c.Sort.GetOrDefault("name", db.ASCOrder).String())
	} else {
		q = q.OrderBy("updated_at ?", c.Sort.GetOrDefault("updatedAt", db.DESCOrder).String())
	}

	q = q.GroupBy("id", "service_id", "name", "description", "created_at", "updated_at")
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

func createEntityTypeQuery(ID, serviceID model.ID, name, description string) (string, []interface{}, error) {
	if ! validator.IsUUID4(ID.String()) {
		panic("how can entity type id not be valid UUID4?")
	}

	if ! validator.IsUUID4(serviceID.String()) {
		panic("how can entity type service id not be a valid UUID4?")
	}

	return sq.Insert("entity_types").
		Columns("id", "name", "description", "service_id", "is_actor").
		Values(sq.Expr("uuid_to_bin(?)", ID.String()), name, description).
		Values(sq.Expr("uuid_to_bin(?)", serviceID.String())).
		ToSql()
}

func firstEntityTypeByNameAndServiceIDQuery(name string, serviceID model.ID) (string, []interface{}, error) {
	if name == "" {
		panic("how can name be empty?")
	}

	if !validator.IsUUID4(serviceID.String()) {
		panic(fmt.Sprintf("%s is not a valid microservice id (UUID4)", serviceID.String()))
	}

	return sq.Select(
		"bin_to_uuid(id) as id",
		"bin_to_uuid(service_id) as service_id",
		"name",
		"description",
		"created_at",
		"updated_at",
		"count(distinct entities.id) as entities_cnt",
	).
		From("entities").
		LeftJoin("entities on entities.entity_type_id = entity_types.id").
		Where(`service_id = uuid_to_bin(?)`, serviceID.String()).
		Where("name = ?", name).
		GroupBy("id", "service_id", "name", "description", "created_at", "updated_at").
		Limit(1).
		ToSql()
}