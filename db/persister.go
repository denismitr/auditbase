package db

import (
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils"
	"github.com/pkg/errors"
)

type PersistenceResult int

const (
	EventCouldNotBeProcessed PersistenceResult = iota
	EventFlowFailed
	DatabaseFailed
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

func (p *DBPersister) NotifyOnResult(r chan<- PersistenceResult) {
	p.results = append(p.results, r)
}

func (p *DBPersister) Persist(e *model.Event) error {
	if err := p.AssignActorTypeTo(e); err != nil {
		p.handlePersistenceError(err, DatabaseFailed)
		return err
	}

	if err := p.AssignActorServiceTo(e); err != nil {
		p.handlePersistenceError(err, DatabaseFailed)
		return err
	}

	if err := p.AssignTargetTypeTo(e); err != nil {
		p.handlePersistenceError(err, DatabaseFailed)
		return err
	}

	if err := p.AssignTargetServiceTo(e); err != nil {
		p.handlePersistenceError(err, DatabaseFailed)
		return err
	}

	if err := p.events.Create(e); err != nil {
		p.handlePersistenceError(err, DatabaseFailed)
		return err
	}

	return nil
}

func (p *DBPersister) AssignTargetTypeTo(e *model.Event) error {
	// TODO: cache these checks
	if e.TargetType.ID != "" {
		tt, err := p.targetTypes.FirstByID(e.TargetType.ID)
		if err != nil {
			return errors.Wrapf(err, "target type with ID %s does not exist", e.TargetType.ID)
		}

		e.TargetType = tt
		return nil
	}

	tt, err := p.targetTypes.FirstOrCreateByName(e.TargetType.Name)
	if err != nil {
		return err
	}

	e.TargetType = tt
	return nil
}

func (p *DBPersister) AssignTargetServiceTo(e *model.Event) error {
	// TODO: cache these checks
	if e.TargetService.ID != "" {
		ts, err := p.microservices.FirstByID(model.ID(e.TargetService.ID))
		if err != nil {
			return errors.Wrapf(err, "target type ID %s does not exist", e.TargetService.ID)
		}

		e.TargetService = ts
		return nil
	}

	ts, err := p.microservices.FirstOrCreateByName(e.TargetService.Name)
	if err != nil {
		return err
	}

	e.TargetService = ts

	return nil
}

func (p *DBPersister) AssignActorTypeTo(e *model.Event) error {
	// TODO: cache these checks
	if e.ActorType.ID != "" {
		at, err := p.actorTypes.FirstByID(e.ActorType.ID)
		if err != nil {
			return errors.Wrapf(err, "actor type with id %s does not exist", e.ActorType.ID)
		}

		e.ActorType = at
		return nil
	}

	at, err := p.actorTypes.FirstOrCreateByName(e.ActorType.Name)
	if err != nil {
		return nil
	}

	e.ActorType = at
	return nil
}

func (p *DBPersister) AssignActorServiceTo(e *model.Event) error {
	// TODO: cache these checks
	if e.ActorService.ID != "" {
		as, err := p.microservices.FirstByID(model.ID(e.ActorService.ID))
		if err != nil {
			return errors.Wrapf(err, "actor service with id %s does not exist", e.ActorService.ID)
		}

		e.ActorService = as
		return nil
	}

	as, err := p.microservices.FirstOrCreateByName(e.ActorService.Name)
	if err != nil {
		return err
	}

	e.ActorService = as
	return nil
}

func (p *DBPersister) handlePersistenceError(
	err error,
	result PersistenceResult,
) {
	p.logger.Error(err)

	for _, r := range p.results {
		r <- result //fixme: use error types
	}
}
