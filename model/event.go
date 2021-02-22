package model

import (
	"context"
	"github.com/denismitr/auditbase/utils/errbag"
	"github.com/denismitr/auditbase/utils/validator"
)

type Crud int

const (
	Unknown Crud = iota
	Create
	Update
	Delete
)

type Action struct {
	ID             string      `json:"id"`
	ParentID       ID          `json:"parentId"`
	Hash           string      `json:"hash"`
	ActorEntityID  ID          `json:"actorEntityId"`
	ActorEntity    *Entity     `json:"actorEntity"`
	TargetEntityID ID          `json:"targetId"`
	TargetEntity   *Entity     `json:"targetEntity"`
	Name           string      `json:"name"`
	EmittedAt      JSONTime    `json:"emittedAt"`
	RegisteredAt   JSONTime    `json:"registeredAt"`
	Details        interface{} `json:"details"`
}

func (a *Action) Validate() *errbag.ErrorBag {
	bag := errbag.New()

	if !validator.IsEmptyString(a.ID) {
		bag.Add("id", ErrMissingEventID)
	}

	if !validator.IsUUID4(a.ID) {
		bag.Add("id", ErrInvalidUUID4)
	}

	if validator.IsEmptyString(a.ActorEntityID.String()) {
		bag.Add("actorID", ErrActorIDEmpty)
	}

	return bag
}

// EventRepository - provides an abstraction over persistent storage
type EventRepository interface {
	Create(context.Context, *Action) error
	Delete(ID) error
	Count() (int, error)
	FindOneByID(ID) (*Action, error)
	Select(*Filter, *Sort, *Pagination) ([]*Action, *Meta, error)
}

// EventPersister persists event to DB
type EventPersister interface {
	Persist(*Action)
	Run(ctx context.Context) error
	NotifyOnResult(chan<- EventPersistenceResult)
}

type EventPersistenceResult interface {
	ID() string
	Err() error
	Ok() bool
}
