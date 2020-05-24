package model

import (
	"github.com/denismitr/auditbase/utils/errbag"
	"github.com/denismitr/auditbase/utils/validator"
)

type Event struct {
	ID            string       `json:"id"`
	ParentEventID string       `json:"parentEventId"`
	Hash          string       `json:"hash"`
	ActorID       string       `json:"actorId"`
	ActorEntity   Entity       `json:"actorEntity"`
	ActorService  Microservice `json:"actorService"`
	TargetID      string       `json:"targetId"`
	TargetEntity  Entity       `json:"targetEntity"`
	TargetService Microservice `json:"targetService"`
	EventName     string       `json:"eventName"`
	EmittedAt     JSONTime     `json:"emittedAt"`
	RegisteredAt  JSONTime     `json:"registeredAt"`
	Delta         []Property   `json:"delta"`
}

func (e *Event) Validate() *errbag.ErrorBag {
	bag := errbag.New()

	if !validator.IsEmptyString(e.ID) {
		bag.Add("id", ErrMissingEventID)
	}

	if !validator.IsUUID4(e.ID) {
		bag.Add("id", ErrInvalidUUID4)
	}

	if validator.IsEmptyString(e.ActorID) {
		bag.Add("actorID", ErrActorIDEmpty)
	}

	if validator.IsEmptyString(e.ActorEntity.Name) {
		bag.Add("actorEntity.Name", ErrActorEntityNameEmpty)
	}

	if validator.IsEmptyString(e.TargetEntity.Name) {
		bag.Add("targetEntity.Name", ErrTargetEntityNameEmpty)
	}

	if validator.IsEmptyString(e.ActorService.Name) {
		bag.Add("actorService.Name", ErrActorServiceNameEmpty)
	}

	if validator.IsEmptyString(e.TargetService.Name) {
		bag.Add("targetService.Name", ErrTargetServiceNameEmpty)
	}

	return bag
}

// EventRepository - provides an abstraction over persistent storage
type EventRepository interface {
	Create(*Event) error
	Delete(ID) error
	Count() (int, error)
	FindOneByID(ID) (*Event, error)
	Select(*Filter, *Sort, *Pagination) ([]*Event, *Meta, error)
}
