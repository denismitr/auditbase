package persister

import (
	"context"
	"github.com/denismitr/auditbase/cache"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/denismitr/auditbase/utils/validator"
	"github.com/pkg/errors"
	"sync"
	"time"
)

type semaphore struct{}

type persister struct {
	factory model.RepositoryFactory
	logger  logger.Logger
	cacher  cache.Cacher

	sem chan semaphore
	running bool
	remember cache.RememberFunc

	actorAssigns chan *payload
	targetAssigns chan *payload
	propertiesAssigns chan *payload
	eventSaves chan *payload

	omu sync.RWMutex
	observers []chan<- model.EventPersistenceResult
}

func New(
	factory model.RepositoryFactory,
	logger  logger.Logger,
	cacher  cache.Cacher,
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
		factory: factory,
		logger: logger,
		cacher: cacher,

		sem: make(chan semaphore, maxEvents),
		running: false,
		remember: remember,

		actorAssigns: make(chan *payload),
		targetAssigns: make(chan *payload),
		propertiesAssigns: make(chan *payload),
		eventSaves: make(chan *payload),

		observers: make([]chan<- model.EventPersistenceResult, 0),
	}
}

func (p *persister) Run(ctx context.Context) error {
	p.running = true

	preAssignActorsNextCh := make(chan *payload)
	preAssignTargetsNextCh := make(chan *payload)
	preAssignPropertiesNextCh := make(chan *payload)
	saveEventNextCh := make(chan *payload)

	go p.preAssignActors(ctx, preAssignActorsNextCh)
	go p.preAssignTargets(ctx, preAssignTargetsNextCh)
	go p.preAssignProperties(ctx, preAssignPropertiesNextCh)
	go p.saveEvent(ctx, saveEventNextCh)

	for {
		select {
			case <-ctx.Done():
				close(p.actorAssigns)
				close(p.targetAssigns)
				close(p.propertiesAssigns)
				close(p.eventSaves)
				p.running = false
				return ctx.Err()
			case pl := <-preAssignActorsNextCh:
				p.targetAssigns <- pl
			case pl := <-preAssignTargetsNextCh:
				p.propertiesAssigns <- pl
			case pl := <-preAssignPropertiesNextCh:
				p.eventSaves <- pl
			case pl := <-saveEventNextCh:
				p.success(pl)
		}
	}
}

func (p *persister) preAssignActors(ctx context.Context, next chan<- *payload) {
	for pl := range p.actorAssigns {
		if pl.isRejected() {
			continue
		}

		go func(pl *payload) {
			ae := pl.actorEntity()
			as := pl.actorService()
			service := new(model.Microservice)
			serviceCacheKey := model.MicroserviceItemCacheKey(as.Name)

			if err := p.remember(serviceCacheKey, 3*time.Minute, service, func() (interface{}, error) {
				v, err := p.factory.Microservices().FirstOrCreateByName(as.Name)
				if err != nil {
					return nil, err
				}
				return v, nil
			}); err != nil {
				p.reject(pl, err)
				return
			}

			entity := new(model.Entity)
			entityCacheKey := model.EntityItemCacheKey(ae.Name, service)

			if err := p.remember(entityCacheKey, 5*time.Minute, entity, func() (interface{}, error) {
				v, err := p.factory.Entities().FirstOrCreateByNameAndService(ae.Name, service)
				if err != nil {
					return nil, err
				}
				return v, nil
			}); err != nil {
				p.reject(pl, err)
				return
			}

			pl.update(func(e *model.Event) {
				e.ActorEntity = *entity
				e.ActorService = *service
			})

			select {
				case <-ctx.Done():
					p.logger.Debugf("preAssignActors received done and now exiting")
					return
				case next <- pl:
					p.logger.Debugf("preAssignActors passed payload to next handler")
			}
		}(pl)
	}
}

func (p *persister) preAssignTargets(ctx context.Context, next chan<- *payload) {
	for pl := range p.targetAssigns {
		if pl.isRejected() {
			continue
		}

		go func (pl *payload) {
			te := pl.targetEntity()
			ts := pl.targetService()

			service := new(model.Microservice)
			serviceCacheKey := model.MicroserviceItemCacheKey(ts.Name)

			err := p.remember(serviceCacheKey, 3*time.Minute, service, func() (interface{}, error) {
				v, err := p.factory.Microservices().FirstOrCreateByName(ts.Name)
				if err != nil {
					return nil, err
				}

				return v, nil
			})

			if err != nil {
				p.reject(pl, err)
				return
			}

			targetEntity := new(model.Entity)
			entityCacheKey := model.EntityItemCacheKey(te.Name, service)

			if err := p.remember(entityCacheKey, 5*time.Minute, targetEntity, func() (interface{}, error) {
				v, err := p.factory.Entities().FirstOrCreateByNameAndService(te.Name, service)
				if err != nil {
					return nil, err
				}

				return v, nil
			}); err != nil {
				p.reject(pl, err)
				return
			}

			pl.update(func(e *model.Event) {
				e.TargetEntity = *targetEntity
				e.TargetService = *service
			})

			select {
				case <-ctx.Done():
					p.logger.Debugf("preAssignTargets received done amd exiting...")
					return
				case next <- pl:
					p.logger.Debugf("preAssignTargets passed payload to the next handler...")
			}
		}(pl)
	}
}

func (p *persister) preAssignProperties(ctx context.Context, next chan<- *payload) {
	for pl := range p.propertiesAssigns {
		if pl.isRejected() {
			continue
		}

		go func (pl *payload) {
			targetEntity := pl.targetEntity()
			propertyNames := pl.changingProperties()

			var wg sync.WaitGroup

			for _, name := range propertyNames {
				wg.Add(1)
				go func(name string) {
					defer wg.Done()

					id, err := p.factory.Properties().GetIDOrCreate(name, targetEntity.ID)
					if err != nil {
						p.reject(pl, err)
						return
					}

					if !validator.IsUUID4(id) {
						p.reject(pl, errors.Errorf("property id %s invalid", id))
						return
					}

					pl.update(func(e *model.Event) {
						for i := range e.Changes {
							if e.Changes[i].PropertyName == name {
								p.logger.Debugf("found match for name %s", name)
								e.Changes[i].PropertyID = id
							}
						}
					})
				}(name)
			}

			wg.Wait()

			select {
			case <-ctx.Done():
				p.logger.Debugf("preAssignProperties received done amd exiting...")
				return
			case next <- pl:
				p.logger.Debugf("preAssignProperties passed payload to the next handler...")
			}
		}(pl)
	}
}

func (p *persister) saveEvent(ctx context.Context, next chan *payload) {
	for pl := range p.eventSaves {
		if pl.isRejected() {
			continue
		}

		go func(pl *payload) {
			e := pl.event()

			if err := p.factory.Events().Create(e); err != nil {
				p.reject(pl, err)
				return
			}

			select {
			case <-ctx.Done():
				p.logger.Debugf("saveEvent received done amd exiting...")
				return
			case next <- pl:
				p.logger.Debugf("saveEvent passed payload to the next handler...")
			}
		}(pl)
	}
}

func (p *persister) reject(pl *payload, err error) {
	<-p.sem

	pl.markAsRejected()
	r := failedResult(pl.eventID(), err)

	p.omu.RLock()
	defer p.omu.RUnlock()
	for i := range p.observers {
		p.observers[i] <- r
	}
}

func (p *persister) success(pl *payload) {
	<-p.sem

	pl.markAsSucceeded()
	r := successResult(pl.eventID())

	p.omu.RLock()
	defer p.omu.RUnlock()
	for i := range p.observers {
		p.observers[i] <- r
	}
}

func (p *persister) Persist(e *model.Event) {
	if p.running == false {
		panic("Persister is not running... You forgot to call the Run method")
	}

	p.sem <- struct{}{}
	p.actorAssigns <- wrap(e)
}

func (p *persister) NotifyOnResult(observer chan<- model.EventPersistenceResult) {
	p.omu.Lock()
	defer p.omu.Unlock()
	p.observers = append(p.observers, observer)
}

