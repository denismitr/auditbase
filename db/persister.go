package db

import (
	"sync"

	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils"
	"github.com/pkg/errors"
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

// Persister persists event to DB
type Persister interface {
	Persist(*model.Event) error
	NotifyOnResult(chan<- PersistenceResult)
}

type DBPersister struct {
	microservices model.MicroserviceRepository
	events        model.EventRepository
	targetTypes   model.TargetTypeRepository
	actorTypes    model.ActorTypeRepository
	logger        utils.Logger

	results []chan<- PersistenceResult
}

// NewDBPersister - creates neew persister
func NewDBPersister(
	microservices model.MicroserviceRepository,
	events model.EventRepository,
	targetTypes model.TargetTypeRepository,
	actorTypes model.ActorTypeRepository,
	logger utils.Logger,
) *DBPersister {
	return &DBPersister{
		microservices: microservices,
		events:        events,
		actorTypes:    actorTypes,
		targetTypes:   targetTypes,
		logger:        logger,
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

	if err := dbp.events.Create(p.event()); err != nil {
		dbp.logger.Error(err)
		dbp.notifyResultObservers(covertToPersistenceResult(err))
		return ErrCouldNotCreateEvent
	}

	return nil
}

func (dbp *DBPersister) prepare(p *payload) {
	var wg sync.WaitGroup
	wg.Add(4)

	go dbp.assignActorType(p, &wg)
	go dbp.assignActorService(p, &wg)
	go dbp.assignTargetService(p, &wg)
	go dbp.assignTargetType(p, &wg)

	wg.Wait()
}

func (dbp *DBPersister) assignTargetType(p *payload, wg *sync.WaitGroup) {
	defer wg.Done()

	ett := p.targetType()

	// TODO: cache these checks
	if ett.ID != "" {
		tt, err := dbp.targetTypes.FirstByID(ett.ID)
		if err != nil {
			p.appendError(ErrEntityDoesNotExist)
			dbp.logger.Error(
				errors.Wrapf(err, "target type with ID %s does not exist", ett.ID))
			return
		}

		p.update(func(e *model.Event) {
			e.TargetType = tt
		})

		return
	}

	tt, err := dbp.targetTypes.FirstOrCreateByName(ett.Name)
	if err != nil {
		p.appendError(ErrDBWriteFailed)
		dbp.logger.Error(err)
		return
	}

	p.update(func(e *model.Event) {
		e.TargetType = tt
	})
}

func (dbp *DBPersister) assignTargetService(p *payload, wg *sync.WaitGroup) {
	defer wg.Done()

	ets := p.targetService()

	// TODO: cache these checks
	if ets.ID != "" {
		ts, err := dbp.microservices.FirstByID(model.ID(ets.ID))
		if err != nil {
			dbp.logger.Error(
				errors.Wrapf(err, "target type ID %s does not exist", ets.ID))
			p.appendError(ErrEntityDoesNotExist)
			return
		}

		p.update(func(e *model.Event) {
			e.TargetService = ts
		})

		return
	}

	ts, err := dbp.microservices.FirstOrCreateByName(ets.Name)
	if err != nil {
		dbp.logger.Error(err)
		p.appendError(ErrDBWriteFailed)
		return
	}

	p.update(func(e *model.Event) {
		e.TargetService = ts
	})
}

func (dbp *DBPersister) assignActorType(p *payload, wg *sync.WaitGroup) {
	defer wg.Done()

	eat := p.ActorType()

	if eat.ID != "" {
		at, err := dbp.actorTypes.FirstByID(eat.ID)
		if err != nil {
			dbp.logger.Error(
				errors.Wrapf(err, "actor type with id %s does not exist", eat.ID))
			p.appendError(ErrEntityDoesNotExist)
			return
		}

		p.update(func(e *model.Event) {
			e.ActorType = at
		})

		return
	}

	at, err := dbp.actorTypes.FirstOrCreateByName(eat.Name)
	if err != nil {
		p.appendError(ErrDBWriteFailed)
		dbp.logger.Error(err)
		return
	}

	p.update(func(e *model.Event) {
		e.ActorType = at
	})
}

func (dbp *DBPersister) assignActorService(p *payload, wg *sync.WaitGroup) {
	defer wg.Done()

	eas := p.ActorService()

	if eas.ID != "" {
		as, err := dbp.microservices.FirstByID(model.ID(eas.ID))
		if err != nil {
			p.appendError(ErrEntityDoesNotExist)
			dbp.logger.Error(
				errors.Wrapf(err, "actor service with id %s does not exist", eas.ID))

			return
		}

		p.update(func(e *model.Event) {
			e.ActorService = as
		})

		return
	}

	as, err := dbp.microservices.FirstOrCreateByName(eas.Name)
	if err != nil {
		p.appendError(ErrDBWriteFailed)
		dbp.logger.Error(err)
		return
	}

	p.update(func(e *model.Event) {
		e.ActorService = as
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
