package consumer

import (
	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils"
	"github.com/pkg/errors"
)

type persistenceResult int

const (
	eventCouldNotBeProcessed = iota
	eventFlowFailed
	databaseFailed
	success
)

type persister interface {
	persist(flow.ReceivedEvent)
}

type dbPersister struct {
	microservices model.MicroserviceRepository
	events        model.EventRepository
	targetTypes   model.TargetTypeRepository
	actorTypes    model.ActorTypeRepository
	logger        utils.Logger

	resultCh chan persistenceResult
}

func newDBPersister(
	microservices model.MicroserviceRepository,
	events model.EventRepository,
	targetTypes model.TargetTypeRepository,
	actorTypes model.ActorTypeRepository,
	logger utils.Logger,
	resultCh chan persistenceResult,
) *dbPersister {
	return &dbPersister{
		microservices: microservices,
		events:        events,
		actorTypes:    actorTypes,
		targetTypes:   targetTypes,
		logger:        logger,
		resultCh:      resultCh,
	}
}

func (p *dbPersister) persist(re flow.ReceivedEvent) {
	e, err := re.Event()
	if err != nil {
		p.handlePersistenceError(err, re, eventFlowFailed)
		return
	}

	if err := p.assignActorTypeTo(&e); err != nil {
		p.handlePersistenceError(err, re, databaseFailed)
		return
	}

	if err := p.assignActorServiceTo(&e); err != nil {
		p.handlePersistenceError(err, re, databaseFailed)
		return
	}

	if err := p.assignTargetTypeTo(&e); err != nil {
		p.handlePersistenceError(err, re, databaseFailed)
		return
	}

	if err := p.assignTargetServiceTo(&e); err != nil {
		p.handlePersistenceError(err, re, databaseFailed)
		return
	}

	if err := p.events.Create(e); err != nil {
		p.handlePersistenceError(err, re, databaseFailed)
		return
	}

	re.Ack()
}

func (p *dbPersister) assignTargetTypeTo(e *model.Event) error {
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

func (p *dbPersister) assignTargetServiceTo(e *model.Event) error {
	// TODO: cache these checks
	if e.TargetService.ID != "" {
		ts, err := p.microservices.FirstByID(e.TargetService.ID)
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

func (p *dbPersister) assignActorTypeTo(e *model.Event) error {
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

func (p *dbPersister) assignActorServiceTo(e *model.Event) error {
	// TODO: cache these checks
	if e.ActorService.ID != "" {
		as, err := p.microservices.FirstByID(e.ActorService.ID)
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

func (p *dbPersister) handlePersistenceError(
	err error,
	re flow.ReceivedEvent,
	result persistenceResult,
) {
	p.logger.Error(err)
	re.Reject()
	p.resultCh <- result
}
