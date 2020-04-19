package model

import (
	"time"

	"github.com/denismitr/auditbase/utils/errbag"
	"github.com/denismitr/auditbase/utils/validator"
)

type Order string

const DESCOrder Order = "DESC"
const ASCOrder Order = "ASC"

type Sort struct {
	Items []map[string]Order
}

type Pagination struct {
	Page    int
	PerPage int
}

type EventFilter struct {
	ActorID          string
	ActorEntityID    string
	ActorEntityName  string
	TargetID         string
	TargetEntityID   string
	TargetEntityName string
	ActorServiceID   string
	TargetServiceID  string
	EventName        string
	EmittedAtGt      int64
	EmittedAtLt      int64
}

func (ef EventFilter) Empty() bool {
	return ef.ActorID == "" && ef.ActorEntityID == ""
}

type Event struct {
	ID            string       `json:"id"`
	ParentEventID string       `json:"parentEventId"`
	ActorID       string       `json:"actorId"`
	ActorEntity   Entity       `json:"actorEntity"`
	ActorService  Microservice `json:"actorService"`
	TargetID      string       `json:"targetId"`
	TargetEntity  Entity       `json:"targetType"`
	TargetService Microservice `json:"targetService"`
	EventName     string       `json:"eventName"`
	EmittedAt     time.Time    `json:"emittedAt"`
	RegisteredAt  time.Time    `json:"registeredAt"`
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

type EventRepository interface {
	Create(*Event) error
	Delete(ID) error
	Count() (int, error)
	FindOneByID(ID) (*Event, error)
	Select(EventFilter, Sort, Pagination) ([]*Event, error)
}
