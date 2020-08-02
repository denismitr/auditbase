package persister

import (
	"context"
	"github.com/denismitr/auditbase/cache"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/pkg/errors"
	"sync"
	"time"
)

type semaphore struct{}

// persister is responsible for inspecting events thoroughly and
// writing events
type persister struct {
	factory model.RepositoryFactory
	logger  logger.Logger
	cacher  cache.Cacher

	sem       chan semaphore
	remember  cache.RememberFunc

	persist chan *payload

	omu       sync.RWMutex
	running   bool
	observers []chan<- model.EventPersistenceResult

	actors int64
	targets int64
	properties int64
	events int64
}

func New(
	factory model.RepositoryFactory,
	logger logger.Logger,
	cacher cache.Cacher,
	maxEvents int,
) model.EventPersister {
	remember := cacher.Remember(func(v, target interface{}) error {
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

	return &persister{
		factory:   factory,
		logger:    logger,
		cacher:    cacher,

		sem:      make(chan semaphore, maxEvents),
		remember: remember,
		running:  false,

		persist:   make(chan *payload),
		observers: make([]chan<- model.EventPersistenceResult, 0),
	}
}

// Run persister in blocking way
func (p *persister) Run(ctx context.Context) error {
	p.omu.Lock()
	p.running = true
	p.omu.Unlock()

	actorAssigns := p.preAssignActors(ctx, p.persist)
	targetAssigns := p.preAssignTargets(ctx, actorAssigns)
	propertyAssigns := p.preAssignProperties(ctx, targetAssigns)
	savedEvents := p.saveEvent(ctx, propertyAssigns)

	for {
		select {
		case <-ctx.Done():
			p.logger.Debugf("Persister context was canceled")
			close(p.persist)
			p.running = false
			return ctx.Err()
		case pl := <-savedEvents:
			p.success(pl)
		}
	}
}

func (p *persister) preAssignActors(ctx context.Context, in <-chan *payload) chan *payload {
	next := make(chan *payload)

	go func() {
		for pl := range in {
			if pl.isRejected() {
				p.logger.Debugf("PAYLOAD REJECTED")
				continue
			}

			go func(pl *payload) {
				ae := pl.actorEntity()
				as := pl.actorService()

				ctx, cancel := context.WithTimeout(ctx, 5 * time.Second)
				defer cancel()

				service, err := prepareService(ctx, as, p.remember, p.factory)
				if err != nil {
					p.reject(pl, err)
					return
				}

				entity, err := prepareEntity(ctx, service, ae, p.remember, p.factory)
				if err != nil {
					p.reject(pl, err)
					return
				}

				_ = pl.update(func(e *model.Event) error {
					e.ActorEntity = *entity
					e.ActorService = *service
					return nil
				})

				next <- pl
			}(pl)
		}

		p.logger.Debugf("Closing down preAssignActors next channel")
		close(next)
	}()

	return next
}

func (p *persister) preAssignTargets(ctx context.Context, in <-chan *payload) chan *payload {
	next := make(chan *payload)

	go func() {
		for pl := range in {
			if pl.isRejected() {
				p.logger.Debugf("PAYLOAD REJECTED")
				continue
			}

			go func(pl *payload) {
				te := pl.targetEntity()
				ts := pl.targetService()

				ctx, cancel := context.WithTimeout(ctx, 5 * time.Second)
				defer cancel()

				service, err := prepareService(ctx, ts, p.remember, p.factory)
				if err != nil {
					p.reject(pl, err)
					return
				}

				targetEntity, err := prepareEntity(ctx, service, te, p.remember, p.factory)
				if err != nil {
					p.reject(pl, err)
					return
				}

				_ = pl.update(func(e *model.Event) error {
					e.TargetEntity = *targetEntity
					e.TargetService = *service
					return nil
				})

				next <- pl
			}(pl)
		}

		close(next)
	}()

	return next
}

func (p *persister) preAssignProperties(ctx context.Context, in <-chan *payload) chan *payload {
	next := make(chan *payload)

	go func() {
		for pl := range in {
			if pl.isRejected() {
				continue
			}

			go func(pl *payload) {
				targetEntity := pl.targetEntity()
				propertyNames := pl.changingProperties()

				ctx, cancel := context.WithTimeout(ctx, 5 * time.Second)
				defer cancel()

				props, err := prepareEntityProperties(ctx, propertyNames, p.factory, targetEntity)
				if err != nil {
					p.reject(pl, err)
					return
				}

				if err := pl.update(func(e *model.Event) error {
					for id, name := range props {
						var match bool

						for i := range e.Changes {
							if e.Changes[i].PropertyName == name {
								e.Changes[i].PropertyID = id
								match = true
								break
							}
						}

						if !match {
							return errors.Errorf("no match found between event [%s] changes and property name [%s]", name)
						}
					}

					return nil
				}); err != nil {
					p.reject(pl, err)
					return
				}

				next <- pl
			}(pl)
		}

		close(next)
	}()

	return next
}

func (p *persister) saveEvent(ctx context.Context, in <-chan *payload) chan *payload {
	next := make(chan *payload)

	go func() {
		for pl := range in {
			if pl.isRejected() {
				continue
			}

			go func(pl *payload) {
				e := pl.event()

				if err := p.factory.Events().Create(e); err != nil {
					p.reject(pl, err)
					return
				}

				next <- pl
			}(pl)
		}

		close(next)
	}()

	return next
}

func (p *persister) reject(pl *payload, err error) {
	if pl.isRejected() {
		return
	}

	go func() {
		pl.markAsRejected()

		<-p.sem

		r := failedResult(pl.eventID(), err)

		p.omu.Lock()
		defer p.omu.Unlock()
		for i := range p.observers {
			p.observers[i] <- r
		}
	}()
}

func (p *persister) success(pl *payload) {
	if pl.isRejected() {
		return
	}

	go func() {
		<-p.sem

		pl.markAsSucceeded()
		r := successResult(pl.eventID())

		p.omu.Lock()
		defer p.omu.Unlock()
		for i := range p.observers {
			p.observers[i] <- r
		}
	}()
}

func (p *persister) Persist(e *model.Event) {
	//p.omu.Lock()
	//if p.running == false {
	//	p.omu.Unlock()
	//	panic("Persister is not running... You forgot to call the Run method")
	//}
	//p.omu.Unlock()

	go func() {
		p.sem <- struct{}{}
		p.persist <- wrap(e)
	}()
}

func (p *persister) NotifyOnResult(observer chan<- model.EventPersistenceResult) {
	p.omu.Lock()
	defer p.omu.Unlock()
	p.observers = append(p.observers, observer)
}
