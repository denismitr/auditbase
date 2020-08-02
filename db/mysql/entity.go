package mysql

import (
	"context"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/denismitr/auditbase/utils/uuid"
	"github.com/denismitr/auditbase/utils/validator"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	sq "github.com/Masterminds/squirrel"
	"time"
)

type entity struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	ServiceID   string `db:"service_id"`
	Description string `db:"description"`
	CreatedAt   string `db:"created_at"`
	UpdatedAt   string `db:"updated_at"`
}

func (e entity) ToModel() *model.Entity {
	return &model.Entity{
		ID:          e.ID,
		Name:        e.Name,
		ServiceID:   e.ServiceID,
		Description: e.Description,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

type EntityRepository struct {
	conn   *sqlx.DB
	logger logger.Logger
	uuid4  uuid.UUID4Generator
}

func NewEntityRepository(conn *sqlx.DB, uuid4 uuid.UUID4Generator, logger logger.Logger) *EntityRepository {
	return &EntityRepository{
		conn:   conn,
		uuid4:  uuid4,
		logger: logger,
	}
}

const selectEntities = `
	SELECT 
		BIN_TO_UUID(id) as id, BIN_TO_UUID(service_id) as service_id, name, description, created_at, updated_at 
	FROM entities
`

// Select all entities
func (r *EntityRepository) Select(f *model.Filter, s *model.Sort, p *model.Pagination) ([]*model.Entity, error) {
	var entities []entity

	sql, args, err := createSelectEntitiesQuery(f, s, p)
	if err != nil {
		return nil, errors.Wrap(err, "could not create sql to select entities")
	}

	stmt, err := r.conn.Preparex(sql)
	if err != nil {
		return nil, errors.Wrap(err, "could not prepare sql to select entities")
	}

	if err := stmt.Select(&entities, args...); err != nil {
		return nil, errors.Wrap(err, "could not execute sql to select entities")
	}

	result := make([]*model.Entity, len(entities))

	for i := range entities {
		result[i] = entities[i].ToModel()
	}

	return result, nil
}

func (r *EntityRepository) Properties(ID string) ([]*model.Property, error) {
	query, err := createSelectPropertiesForEntity(ID)

	if err != nil {
		return nil, err
	}

	r.logger.Debugf(query)

	var properties []property

	if err := r.conn.Select(&properties, query, ID); err != nil {
		return nil, errors.Wrapf(err, "could not get properties stat from entities with ID [%s]", ID)
	}

	stats := make([]*model.Property, len(properties))

	for i := range properties {
		stats[i] = properties[i].ToModel()
	}

	return stats, nil
}

func createSelectPropertiesForEntity(_ string) (string, error) {
	query, _, err := sq.Select(
		"BIN_TO_UUID(p.id) as id",
		"BIN_TO_UUID(p.entity_id) as entity_id",
		"COUNT(c.id) as change_count",
		"p.name", "p.type",
	).From(
		"properties as p",
	).Join(
		"changes as c ON c.property_id = p.id",
	).Where(
		"p.entity_id = UUID_TO_BIN(?)",
	).GroupBy(
		"p.id", "p.entity_id", "p.name", "p.type",
	).ToSql()

	return query, err
}

// Create an entities
func (r *EntityRepository) Create(e *model.Entity) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2 * time.Second)
	defer cancel()

	stmt := `
		INSERT INTO entities (id, service_id, name, description) VALUES (
			UUID_TO_BIN(:id), UUID_TO_BIN(:service_id), :name, :description
		)
	`

	tt := entity{
		ID:          e.ID,
		Name:        e.Name,
		ServiceID:   e.Service.ID,
		Description: e.Description,
	}

	if _, err := r.conn.NamedExecContext(ctx, stmt, tt); err != nil {
		return errors.Wrapf(err, "could not create new entities with name %s", e.Name)
	}

	return nil
}

func (r *EntityRepository) FirstByNameAndService(name string, service *model.Microservice) (*model.Entity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2 * time.Second)
	defer cancel()

	stmt := `
		SELECT 
			BIN_TO_UUID(id) as id, 
			BIN_TO_UUID(service_id) as service_id, 
			name, 
			description, 
			created_at, 
			updated_at 
		FROM entities
			WHERE service_id = UUID_TO_BIN(?) AND name = ?
		LIMIT 1
	`

	ent := new(entity)

	if err := r.conn.GetContext(ctx, ent, stmt, service.ID, name); err != nil {
		return nil, errors.Wrapf(err, "could not find entities with name %s", name)
	}

	return ent.ToModel(), nil
}

func (r *EntityRepository) FirstByID(ID string) (*model.Entity, error) {
	sql, args, err := createFirstEntityByIDQuery(ID)
	if err != nil {
		return nil, err
	}

	ent := entity{}

	stmt, err := r.conn.Preparex(sql)
	if err != nil {
		return nil, errors.Errorf("could not prepare sql statement %s", sql)
	}

	if err := stmt.Get(&ent, args...); err != nil {
		return nil, errors.Wrapf(err, "could not find entities with ID %s", ID)
	}

	return ent.ToModel(), nil
}

// FirstOrCreateByNameAndService - fetches first or creates an entities by its name and service
func (r *EntityRepository) FirstOrCreateByNameAndService(name string, service *model.Microservice) (*model.Entity, error) {
	ent, err := r.FirstByNameAndService(name, service)
	if err == nil {
		return ent, nil
	}

	//r.logger.Debugf(err.Error())

	ent = &model.Entity{
		ID:          r.uuid4.Generate(),
		Service:     service,
		Name:        name,
		Description: "",
	}

	if err := r.Create(ent); err != nil {
		return nil, errors.Wrapf(err, "entities %s with service ID %s does not exist and cannot be created", name, service)
	}

	return ent, nil
}

func createSelectEntitiesQuery(f *model.Filter, s *model.Sort, p *model.Pagination) (string, []interface{}, error) {
	q := sq.Select(
		"BIN_TO_UUID(id) as id",
		"BIN_TO_UUID(service_id) as service_id",
		"name", "description", "created_at", "updated_at",
	).From("entities")

	if f.Has("serviceId") {
		q = q.Where(`service_id = uuid_to_bin(?)`, f.StringOrDefault("serviceId", ""))
	}

	if s.Has("name") {
		q = q.OrderByClause("name ?", s.GetOrDefault("name", model.ASCOrder).String())
	} else if !f.Has("serviceId") {
		q = q.OrderBy("created_at DESC")
	} else {
		q = q.OrderBy("service_id DESC")
	}

	q = q.Limit(uint64(p.PerPage))
	q = q.Offset(uint64(p.Offset()))

	return q.ToSql()
}

func createFirstEntityByIDQuery(ID string) (string, []interface{}, error) {
	if ! validator.IsUUID4(ID) {
		return "", nil, errors.Errorf("%s is not a valid UUID4", ID)
	}

	return sq.Select(
		"BIN_TO_UUID(id) as id",
		"BIN_TO_UUID(service_id) as service_id",
		"name", "description", "created_at", "updated_at").
	From("entities").
	Where("id = UUID_TO_BIN(?)", ID).
	Limit(1).
	ToSql()
}
