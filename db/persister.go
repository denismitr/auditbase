package db

import (
	"github.com/denismitr/auditbase/cache"
	"sync"
	"time"

	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/errtype"
	"github.com/denismitr/auditbase/utils/logger"
)

type PersistenceResult int

const (
	EventCouldNotBeProcessed PersistenceResult = iota
	EventFlowFailed
	CriticalDatabaseFailure
	LogicalError
	UnknownError
	Success
)

const ErrDBTransactionFailed = errtype.StringError("db transaction failed")

// Persister persists event to DB
type Persister interface {
	Persist(*model.Event) error
	NotifyOnResult(chan<- PersistenceResult)
}

type DBPersister struct {
	microservices model.MicroserviceRepository
	events        model.EventRepository
	entities      model.EntityRepository
	logger        logger.Logger
	cacher        cache.Cacher

	results []chan<- PersistenceResult
}

// NewDBPersister - creates neew persister
func NewDBPersister(
	microservices model.MicroserviceRepository,
	events model.EventRepository,
	entities model.EntityRepository,
	logger logger.Logger,
	cacher cache.Cacher,
) *DBPersister {
	return &DBPersister{
		microservices: microservices,
		events:        events,
		entities:      entities,
		logger:        logger,
		cacher:        cacher,
	}
}

func (dbp *DBPersister) NotifyOnResult(r chan<- PersistenceResult) {
	dbp.results = append(dbp.results, r)
}

func (dbp *DBPersister) Persist(e *model.Event) error {
	p := wrap(e)
	dbp.prepare(p)

	if p.hasErrors() {
		for _, err := range p.errors() {
			dbp.notifyResultObservers(covertToPersistenceResult(err))
		}

		return ErrPersisterCouldNotPrepareEvent
	}

	p.update(func(e *model.Event) {
		e.ActorEntity.Service = &e.ActorService
		e.TargetEntity.Service = &e.TargetService
	})

	if err := dbp.events.Create(p.event()); err != nil {
		dbp.logger.Error(err)
		dbp.notifyResultObservers(covertToPersistenceResult(err))
		return ErrCouldNotCreateEvent
	}

	return nil
}

func (dbp *DBPersister) prepare(p *payload) {
	var wg sync.WaitGroup
	wg.Add(2)

	go dbp.assignActorEntity(p, &wg)
	go dbp.assignTargetEntity(p, &wg)

	wg.Wait()
}

func (dbp *DBPersister) assignTargetEntity(p *payload, wg *sync.WaitGroup) {
	defer wg.Done()

	te := p.targetEntity()
	ts := p.targetService()

	service := new(model.Microservice)
	err := dbp.cacher.Remember(ts.Name, 1*time.Minute, service, func() (interface{}, error) {
		v, err := dbp.microservices.FirstOrCreateByName(ts.Name)
		if err != nil {
			return nil, err
		}

		return v, nil
	})
	dbp.logger.Debugf("SERVICE: %#v", service)
	if err != nil {
		p.appendError(ErrDBWriteFailed)
		dbp.logger.Error(err)
	}

	targetEntity := new(model.Entity)
	if err := dbp.cacher.Remember(te.Name, 1*time.Minute, targetEntity, func() (interface{}, error) {
		v, err := dbp.entities.FirstOrCreateByNameAndService(te.Name, service)
		if err != nil {
			return nil, err
		}

		return v, nil
	}); err != nil {
		p.appendError(ErrDBWriteFailed)
		dbp.logger.Error(err)
		return
	}


	p.update(func(e *model.Event) {
		e.TargetEntity = *targetEntity
		e.TargetService = *service
	})
}

func (dbp *DBPersister) assignActorEntity(p *payload, wg *sync.WaitGroup) {
	defer wg.Done()

	ae := p.actorEntity()
	as := p.actorService()

	service := new(model.Microservice)
	if err := dbp.cacher.Remember(as.Name, 1 * time.Minute, service, func() (interface{}, error) {
		v, err := dbp.microservices.FirstOrCreateByName(as.Name)
		if err != nil {
			return nil, err
		}
		return v, nil
	}); err != nil {
		p.appendError(ErrDBWriteFailed)
		dbp.logger.Error(err)
		return
	}
	dbp.logger.Debugf("SERVICE 2: %#v", service)
	entity := new(model.Entity)
	if err := dbp.cacher.Remember(ae.Name, 1 * time.Minute, entity, func() (interface{}, error) {
		v, err := dbp.entities.FirstOrCreateByNameAndService(ae.Name, service)
		if err != nil {
			return nil, err
		}
		return v, nil
	}); err != nil {
		p.appendError(ErrDBWriteFailed)
		dbp.logger.Error(err)
		return
	}

	p.update(func(e *model.Event) {
		e.ActorEntity = *entity
		e.ActorService = *service
	})
}

func (dbp *DBPersister) notifyResultObservers(result PersistenceResult) {
	for _, r := range dbp.results {
		select {
		case r <- result:
		default:
		}
	}
}
