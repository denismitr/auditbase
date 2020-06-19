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

	mu            sync.Mutex
	changes       *ChangeRepository
	properties    *PropertyRepository
	events        *EventRepository
	entities      *EntityRepository
	microservices *MicroserviceRepository
}

func (r *RepositoryFactory) Changes() model.ChangeRepository {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.changes == nil {
		r.changes = NewChangeRepository(r.conn, r.log, r.uuid4)
	}

	return r.changes
}

func (r *RepositoryFactory) Properties() model.PropertyRepository {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.properties == nil {
		r.properties = NewPropertyRepository(r.conn, r.uuid4, r.log)
	}

	return r.properties
}

func (r *RepositoryFactory) Events() model.EventRepository {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.events == nil {
		r.events = NewEventRepository(r.conn, r.uuid4, r.log)
	}

	return r.events
}

func (r *RepositoryFactory) Entities() model.EntityRepository {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.entities == nil {
		r.entities = NewEntityRepository(r.conn, r.uuid4, r.log)
	}

	return r.entities
}

func (r *RepositoryFactory) Microservices() model.MicroserviceRepository {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.microservices == nil {
		r.microservices = NewMicroserviceRepository(r.conn, r.uuid4, r.log)
	}

	return r.microservices
}

func NewRepositoryFactory(conn *sqlx.DB, uuid4 uuid.UUID4Generator, log logger.Logger) *RepositoryFactory {
	return &RepositoryFactory{
		conn: conn,
		log: log,
		uuid4: uuid4,
	}
}



