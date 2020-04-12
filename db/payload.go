package db

import (
	"sync"

	"github.com/denismitr/auditbase/model"
)

type updater func(*model.Event)

type payload struct {
	mu       sync.RWMutex
	e        *model.Event
	errorBus []error
}

func wrap(e *model.Event) *payload {
	return &payload{
		e:        e,
		errorBus: make([]error, 0),
	}
}

func (p *payload) hasErrors() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.errorBus) > 0
}

func (p *payload) errors() []error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.errorBus
}

func (p *payload) appendError(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.errorBus = append(p.errorBus, err)
}

func (p *payload) targetType() model.TargetType {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.e.TargetType
}

func (p *payload) targetService() model.Microservice {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.e.TargetService
}

func (p *payload) ActorType() model.ActorType {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.e.ActorType
}

func (p *payload) ActorService() model.Microservice {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.e.ActorService
}

func (p *payload) event() *model.Event {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.e
}

func (p *payload) update(f updater) {
	p.mu.Lock()
	defer p.mu.Unlock()
	f(p.e)
}
