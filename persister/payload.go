package persister

import (
	"sync"

	"github.com/denismitr/auditbase/model"
)

type payloadStatus uint8

const (
	pending = iota
	succeeded
	rejected
)

type updater func(*model.Event) error
type mapper func(model.Event) interface{}

type payload struct {
	mu       sync.RWMutex
	e        *model.Event
	status payloadStatus
}

func wrap(e *model.Event) *payload {
	return &payload{
		e:        e,
		status: pending,
	}
}

func (p *payload) markAsRejected() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.status = rejected
}

func (p *payload) isRejected() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.status == rejected
}

func (p *payload) markAsSucceeded() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.status = succeeded
}

func (p *payload) targetEntity() model.Entity {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.e.TargetEntity
}

func (p *payload) targetService() model.Microservice {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.e.TargetService
}

func (p *payload) changingProperties() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	var propertyNames []string
	for i := range p.e.Changes {
		propertyNames = append(propertyNames, p.e.Changes[i].PropertyName)
	}
	return propertyNames
}

func (p *payload) actorEntity() model.Entity {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.e.ActorEntity
}

func (p *payload) actorService() model.Microservice {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.e.ActorService
}

func (p *payload) event() *model.Event {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.e
}

func (p *payload) eventID() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.e.ID
}


func (p *payload) update(f updater) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return f(p.e)
}

func (p *payload) mapper(f mapper) interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return f(*p.e)
}
