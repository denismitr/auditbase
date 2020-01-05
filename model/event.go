package model

import (
	"time"

	"github.com/denismitr/auditbase/utils"
)

type Event struct {
	ID              string                   `json:"id"`
	ParentEventID   string                   `json:"parentEventId"`
	ActorID         string                   `json:"actorId"`
	ActorType       ActorType                `json:"actorType"`
	ActorServiceID  string                   `json:"actorServiceId"`
	TargetID        string                   `json:"targetId"`
	TargetType      TargetType               `json:"targetType"`
	TargetServiceID string                   `json:"targetServiceId"`
	EventName       string                   `json:"eventName"`
	EmittedAt       int64                    `json:"emittedAt"`
	RegisteredAt    int64                    `json:"registeredAt"`
	Delta           map[string][]interface{} `json:"delta"`
}

func (e *Event) Validate(v Validator) ValidationErrors {
	if v.IsEmptyString(e.ID) {
		e.ID = utils.UUID4()
	}

	if !v.IsUUID4(e.ID) {
		v.Add("id", ":id is not a valid UUID4")
	}

	if v.IsEmptyString(e.ActorID) {
		v.Add("actorID", ":actorID must not be empty")
	}

	if v.IsEmptyString(e.ActorType.Name) {
		v.Add("actorType.ID", ":actorType.ID must not be empty")
	}

	if v.IsEmptyString(e.TargetType.Name) {
		v.Add("targetType.ID", ":targetType.ID must not be empty")
	}

	// TODO: add validation, should not be empty
	if e.EmittedAt == 0 {
		e.EmittedAt = time.Now().Unix()
	}

	e.RegisteredAt = time.Now().Unix()

	return v.Errors()
}

type EventRepository interface {
	Create(Event) error
	Delete(string) error
	FindOneByID(string) (Event, error)
	SelectAll() ([]Event, error)
}

type EventExchange interface {
	Publish(Event) error
	Consume() <-chan Event
}
