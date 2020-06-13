package mysql

import (
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/denismitr/auditbase/utils/uuid"
	"github.com/jmoiron/sqlx"
	"sync"
)

type RepositoryFactory struct {
	conn  *sqlx.DB
	log logger.Logger
	uuid4 uuid.UUID4Generator

	mu sync.Mutex
	property *PropertyRepository
	event *EventRepository
	entity *EntityRepository
	microservice *MicroserviceRepository
}

func (r *RepositoryFactory) Properties() model.PropertyRepository {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.property == nil {
		r.property = NewPropertyRepository(r.conn, r.uuid4, r.log)
	}

	return r.property
}

func (r *RepositoryFactory) Events() model.EventRepository {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.event == nil {
		r.event = NewEventRepository(r.conn, r.uuid4, r.log)
	}

	return r.event
}

func (r *RepositoryFactory) Entities() model.EntityRepository {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.entity == nil {
		r.entity = NewEntityRepository(r.conn, r.uuid4, r.log)
	}

	return r.entity
}

func (r *RepositoryFactory) Microservices() model.MicroserviceRepository {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.microservice == nil {
		r.microservice = NewMicroserviceRepository(r.conn, r.uuid4, r.log)
	}

	return r.microservice
}

func NewRepositoryFactory(conn *sqlx.DB, uuid4 uuid.UUID4Generator, log logger.Logger) *RepositoryFactory {
	return &RepositoryFactory{
		conn: conn,
		log: log,
		uuid4: uuid4,
	}
}



