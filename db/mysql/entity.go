package mysql

import (
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/denismitr/auditbase/utils/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	sq "github.com/Masterminds/squirrel"
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
	query, args := createSelectEntitiesQuery(f, s, p)
	r.logger.SQL(query, args)
	stmt, err := r.conn.PrepareNamed(query)
	if err != nil {
		return nil, errors.Wrap(err, "could not prepare named stmt to select entities")
	}

	if err := stmt.Select(&entities, args); err != nil {
		return nil, errors.Wrap(err, "could not select all entities")
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
		return nil, errors.Wrapf(err, "could not get property stat from entity with ID [%s]", ID)
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

// Create an entity
func (r *EntityRepository) Create(e *model.Entity) error {
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

	if _, err := r.conn.NamedExec(stmt, tt); err != nil {
		return errors.Wrapf(err, "could not create new entity with name %s", e.Name)
	}

	return nil
}

func (r *EntityRepository) FirstByNameAndService(name string, service *model.Microservice) (*model.Entity, error) {
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

	if err := r.conn.Get(ent, stmt, service.ID, name); err != nil {
		return nil, errors.Wrapf(err, "could not find entity with name %s", name)
	}

	return ent.ToModel(), nil
}

func (r *EntityRepository) FirstByID(ID string) (*model.Entity, error) {
	stmt := `
		SELECT 
			BIN_TO_UUID(id) as id, BIN_TO_UUID(service_id) as service_id, name, description, created_at, updated_at 
		FROM entities
			WHERE id = UUID_TO_BIN(?)
		LIMIT 1
	`

	ent := entity{}

	if err := r.conn.Get(&ent, stmt, ID); err != nil {
		return nil, errors.Wrapf(err, "could not find entity with ID %s", ID)
	}

	return ent.ToModel(), nil
}

// FirstOrCreateByNameAndService - fetches first or creates an entity by its name and service
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
		return nil, errors.Wrapf(err, "entity %s with service ID %s does not exist and cannot be created", name, service)
	}

	return ent, nil
}

func createSelectEntitiesQuery(f *model.Filter, s *model.Sort, p *model.Pagination) (string, map[string]interface{}) {
	stmt := selectEntities
	args := make(map[string]interface{})

	if f.Has("serviceId") {
		stmt += ` where service_id = uuid_to_bin(:serviceId)`
		args["serviceId"] = f.StringOrDefault("serviceId", "")
	}

	if s.Has("name") {
		stmt += ` order by name ` + s.GetOrDefault("name", model.ASCOrder).String()
	} else if !f.Has("serviceId") {
		stmt += ` order by created_at DESC`
	} else {
		stmt += ` order by service_id DESC`
	}

	stmt += ` limit :offset, :limit`
	args["limit"] = p.PerPage
	args["offset"] = p.Offset()

	return stmt, args
}
