package db

import (
	"github.com/denismitr/auditbase/cache"
	"github.com/pkg/errors"
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
	factory model.RepositoryFactory
	logger  logger.Logger
	cacher  cache.Cacher

	results []chan<- PersistenceResult
}

// NewDBPersister - creates neew persister
func NewDBPersister(
	factory model.RepositoryFactory,
	logger logger.Logger,
	cacher cache.Cacher,
) *DBPersister {
	return &DBPersister{
		factory: factory,
		logger:  logger,
		cacher:  cacher,
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

	if err := dbp.factory.Events().Create(p.event()); err != nil {
		dbp.logger.Error(err)
		dbp.notifyResultObservers(covertToPersistenceResult(err))
		return ErrCouldNotCreateEvent
	}

	return nil
}

func (dbp *DBPersister) prepare(p *payload) {
	remember := dbp.cacher.Remember(func(v, target interface{}) error {
		switch t := target.(type) {
		case *model.Microservice:
			*t = *v.(*model.Microservice)
		case *model.Entity:
			*t = *v.(*model.Entity)
		case *model.Property:
			*t = *v.(*model.Property)
		case *string:
			*t = *v.(*string)
		default:
			return cache.ErrCouldNotRawValueToTarget
		}

		return nil
	})

	var wg sync.WaitGroup
	wg.Add(2)

	go dbp.assignActorEntity(remember, p, &wg)
	go dbp.assignTargetEntityAndProperties(remember, p, &wg)

	wg.Wait()
}

func (dbp *DBPersister) assignTargetEntityAndProperties(remember cache.RememberFunc, p *payload, wg *sync.WaitGroup) {
	defer wg.Done()

	te := p.targetEntity()
	ts := p.targetService()

	service := new(model.Microservice)
	serviceCacheKey := model.MicroserviceItemCacheKey(ts.Name)

	err := remember(serviceCacheKey, 3*time.Minute, service, func() (interface{}, error) {
		v, err := dbp.factory.Microservices().FirstOrCreateByName(ts.Name)
		if err != nil {
			return nil, err
		}

		return v, nil
	})

	if err != nil {
		p.appendError(ErrDBWriteFailed)
		dbp.logger.Error(err)
	}

	targetEntity := new(model.Entity)
	entityCacheKey := model.EntityItemCacheKey(te.Name, service)

	if err := remember(entityCacheKey, 5*time.Minute, targetEntity, func() (interface{}, error) {
		v, err := dbp.factory.Entities().FirstOrCreateByNameAndService(te.Name, service)
		if err != nil {
			return nil, err
		}

		return v, nil
	}); err != nil {
		p.appendError(ErrDBWriteFailed)
		dbp.logger.Error(err)
		return
	}

	props, err := dbp.getOrCreatePropertyIds(p.changingProperties(), targetEntity)
	if err != nil {
		dbp.logger.Error(err)
		p.appendError(ErrDBWriteFailed)
		return
	}

	p.update(func(e *model.Event) {
		e.TargetEntity = *targetEntity
		e.TargetService = *service

		for name, id := range props {
			for i := range e.Changes {
				if e.Changes[i].PropertyName == name {
					e.Changes[i].PropertyID = id
				}
			}
		}
	})
}

func (dbp *DBPersister) assignActorEntity(remember cache.RememberFunc, p *payload, wg *sync.WaitGroup) {
	defer wg.Done()

	ae := p.actorEntity()
	as := p.actorService()

	service := new(model.Microservice)
	serviceCacheKey := model.MicroserviceItemCacheKey(as.Name)

	if err := remember(serviceCacheKey, 3*time.Minute, service, func() (interface{}, error) {
		v, err := dbp.factory.Microservices().FirstOrCreateByName(as.Name)
		if err != nil {
			return nil, err
		}
		return v, nil
	}); err != nil {
		p.appendError(ErrDBWriteFailed)
		dbp.logger.Error(err)
		return
	}

	entity := new(model.Entity)
	entityCacheKey := model.EntityItemCacheKey(ae.Name, service)

	if err := remember(entityCacheKey, 5*time.Minute, entity, func() (interface{}, error) {
		v, err := dbp.factory.Entities().FirstOrCreateByNameAndService(ae.Name, service)
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

func (dbp *DBPersister) getOrCreatePropertyIds(propertyNames []string, entity *model.Entity) (map[string]string, error) {
	sink := newPropertySink()

	var wg sync.WaitGroup

	for _, name := range propertyNames {
		wg.Add(1)
		go func(propertyName string) {
			defer wg.Done()

			id, err := dbp.factory.Properties().GetIDOrCreate(propertyName, entity.ID)

			if err != nil {
				sink.err(err)
			} else {
				sink.add(propertyName, id)
			}
		}(name)
	}

	wg.Wait()

	if sink.hasErrors() {
		return nil, sink.firstError()
	}

	if sink.count() != len(propertyNames) {
		return nil, errors.Errorf("could not get all IDs for all properties: %#v", propertyNames)
	}

	return sink.all(), nil
}
