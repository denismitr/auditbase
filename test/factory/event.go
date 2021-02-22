package factory

import (
	"fmt"
	"github.com/denismitr/auditbase/db"
	"github.com/denismitr/auditbase/model"
	"github.com/golang/mock/gomock"
	"time"
)

type incomingEventState string
type eventState string

type Matcher interface {
	gomock.Matcher
	ID() string
}

const (
	EventWithID incomingEventState = "event_with_id"
	EventWithoutID incomingEventState = "event_without_id"
	EventWithoutEmittedAt incomingEventState = "event_without_emitted_at"
)

const (
	DefaultEvent eventState = "default_event"
)

type IncomingEventState struct {
	State incomingEventState
	Now time.Time
}

type EventState struct {
	State eventState
	Now time.Time
}

type incomingEventMatcher struct {
	evt *model.Action
	state incomingEventState
}


type EventMatcher struct {
	Evt *model.Action
	state eventState
}

func (i incomingEventMatcher) Matches(x interface{}) bool {
	if i.evt == nil && x == nil {
		return true
	}

	e, ok := x.(*model.Action)
	if !ok {
		return false
	}

	expectedChangesCount := len(i.evt.Changes)
	actualChangesCount := len(e.Changes)

	if actualChangesCount != expectedChangesCount {
		return false
	}

	for j := range i.evt.Changes {
		expChange := i.evt.Changes[j]
		actChange := e.Changes[j]

		if expChange.PropertyName != actChange.PropertyName {
			return false
		}

		if expChange.From != nil && actChange.From != nil && *expChange.From != *actChange.From {
			return false
		}

		if expChange.To != nil && actChange.To != nil && *expChange.To != *actChange.To {
			return false
		}
	}

	return i.evt.ID == e.ID &&
		i.evt.Hash == e.Hash &&
		i.evt.ActorID == e.ActorID &&
		i.evt.TargetID == e.TargetID &&
		i.evt.TargetService.Name == e.TargetService.Name &&
		i.evt.ActorService.Name == e.ActorService.Name &&
		i.evt.ActorEntity.Name == e.ActorEntity.Name &&
		i.evt.TargetEntity.Name == e.TargetEntity.Name &&
		i.evt.EmittedAt.Equal(e.EmittedAt.Time) &&
		i.evt.RegisteredAt.Equal(e.RegisteredAt.Time)
}

func (i incomingEventMatcher) String() string {
	return fmt.Sprintf("is similar to %#v given state %s", i.evt, i.state)
}

func (i incomingEventMatcher) ID() string {
	return i.evt.ID
}

func (evt EventMatcher) Matches(x interface{}) bool {
	if evt.Evt == nil && x == nil {
		return true
	}

	e, ok := x.(*model.Action)
	if !ok {
		return false
	}

	expectedChangesCount := len(evt.Evt.Changes)
	actualChangesCount := len(e.Changes)

	if actualChangesCount != expectedChangesCount {
		return false
	}

	for j := range evt.Evt.Changes {
		expChange := evt.Evt.Changes[j]
		actChange := e.Changes[j]

		if expChange.PropertyName != actChange.PropertyName {
			return false
		}

		if expChange.From != nil && actChange.From != nil && *expChange.From != *actChange.From {
			return false
		}

		if expChange.To != nil && actChange.To != nil && *expChange.To != *actChange.To {
			return false
		}
	}

	return evt.Evt.ID == e.ID &&
		evt.Evt.Hash == e.Hash &&
		evt.Evt.ActorID == e.ActorID &&
		evt.Evt.TargetID == e.TargetID &&
		evt.Evt.TargetService.Name == e.TargetService.Name &&
		evt.Evt.ActorService.Name == e.ActorService.Name &&
		evt.Evt.ActorEntity.Name == e.ActorEntity.Name &&
		evt.Evt.TargetEntity.Name == e.TargetEntity.Name &&
		evt.Evt.EmittedAt.Equal(e.EmittedAt.Time) &&
		evt.Evt.RegisteredAt.Equal(e.RegisteredAt.Time)
}

func (evt EventMatcher) String() string {
	return fmt.Sprintf("is similar to %#v given state %s", evt.Evt, evt.state)
}

func (evt EventMatcher) ID() string {
	return evt.Evt.ID
}

func MatchingIncomingEvent(state IncomingEventState) ([]byte, Matcher) {
	var raw string
	var e *model.Action

	switch state.State {
	case EventWithID:
		raw, e = createIncomingEventWithId(state.Now)
	case EventWithoutID:
		raw, e = createIncomingEventWithoutId(state.Now)
	case EventWithoutEmittedAt:
		raw, e = createIncomingEventWithoutEmittedAt(state.Now)
	default:
		panic(fmt.Sprintf("%s state is not supported", state.State))
	}

	return []byte(raw), incomingEventMatcher{evt: e, state: state.State}
}

func MatchingEvent(state EventState) (*EventMatcher) {
	var e *model.Action

	switch state.State {
	case DefaultEvent:
		e = createDefaultEvent(state.Now)
	default:
		panic(fmt.Sprintf("%s state is not supported", state.State))
	}

	return &EventMatcher{Evt: e, state: state.State}
}

func createDefaultEvent(now time.Time) *model.Action {
	return &model.Action{
		ID: "34e1d82a-a065-436d-afd0-5fbcb752a4e1",
		Hash: "1790e8a793ecd7f0b3e46c5dc5f71d18fc24c45a",
		TargetService: model.Microservice{
			ID: "355e1d82a-a065-436d-afd0-5fbcb752a4b4",
			Name: "billing",
		},
		TargetEntity: model.Entity{
			ID: "444e1d82a-a065-436d-afd0-5fbcb752ae5",
			Name: "subscription",
		},
		TargetID: "122242120",
		ActorID: "88999",
		ActorService: model.Microservice{
			Name: "web",
		},
		ActorEntity: model.Entity{
			Name: "user",
		},
		EventName: "subscriptionCanceled",
		EmittedAt: model.JSONTime{Time: time.Unix(1578173212, 0)},
		RegisteredAt: model.JSONTime{Time: now},
		Changes: []*model.PropertyChange{
			{
				PropertyName: "status",
				CurrentDataType: model.StringDataType,
				From: db.PointerFromString("active"),
				To: db.PointerFromString("canceled"),
			},
			{
				PropertyName: "rating",
				CurrentDataType: model.IntegerDataType,
				From: db.PointerFromString("500"),
				To: nil,
			},
		},
	}
}

func createIncomingEventWithId(now time.Time) (string, *model.Action) {
	raw := `{
		"id": "34e1d82a-a065-436d-afd0-5fbcb752a4e1",
		"targetId": "122242120",
		"targetEntity": "subscription",
		"targetService": "billing",
		"actorId": "88999",
		"actorEntity": "user",
		"actorService": "web",
		"eventName": "subscriptionCanceled",
		"emittedAt": 1578173212,
		"changes": [
			{
				"propertyName": "status",
				"from": "active",
				"to": "canceled"
			},
			{
				"propertyName": "rating",
				"from": "500",
				"to": null
			}
		]
	}`

	m := &model.Action{
		ID: "34e1d82a-a065-436d-afd0-5fbcb752a4e1",
		Hash: "1790e8a793ecd7f0b3e46c5dc5f71d18fc24c45a",
		TargetService: model.Microservice{
			Name: "billing",
		},
		TargetEntity: model.Entity{
			Name: "subscription",
		},
		TargetID: "122242120",
		ActorID: "88999",
		ActorService: model.Microservice{
			Name: "web",
		},
		ActorEntity: model.Entity{
			Name: "user",
		},
		EventName: "subscriptionCanceled",
		EmittedAt: model.JSONTime{Time: time.Unix(1578173212, 0)},
		RegisteredAt: model.JSONTime{Time: now},
		Changes: []*model.PropertyChange{
			{
				PropertyName: "status",
				CurrentDataType: model.StringDataType,
				From: db.PointerFromString("active"),
				To: db.PointerFromString("canceled"),
			},
			{
				PropertyName: "rating",
				CurrentDataType: model.IntegerDataType,
				From: db.PointerFromString("500"),
				To: nil,
			},
		},
	}

	return raw, m
}

func createIncomingEventWithoutId(now time.Time) (string, *model.Action) {
	raw := `{
		"targetId": "122242120",
		"targetEntity": "subscription",
		"targetService": "billing",
		"actorId": "88999",
		"actorEntity": "user",
		"actorService": "web",
		"eventName": "subscriptionCanceled",
		"emittedAt": 1578173212,
		"changes": [
			{
				"propertyName": "status",
				"currentPropertyType": "string",
				"from": "active",
				"to": "canceled"
			},
			{
				"propertyName": "rating",
				"currentPropertyType": "integer",
				"from": "500",
				"to": null
			}
		]
	}`

	m := &model.Action{
		ID: "22e1d82a-a065-436d-afd0-5fbcb752a4f3",
		Hash: "fb01901eb94091e8dd6c38c81f7d2576ff4ec735",
		TargetService: model.Microservice{
			Name: "billing",
		},
		TargetEntity: model.Entity{
			Name: "subscription",
		},
		TargetID: "122242120",
		ActorID: "88999",
		ActorService: model.Microservice{
			Name: "web",
		},
		ActorEntity: model.Entity{
			Name: "user",
		},
		EventName: "subscriptionCanceled",
		EmittedAt: model.JSONTime{Time: time.Unix(1578173212, 0)},
		RegisteredAt: model.JSONTime{Time: now},
		Changes: []*model.PropertyChange{
			{
				PropertyName: "status",
				CurrentDataType: model.StringDataType,
				From: db.PointerFromString("active"),
				To: db.PointerFromString("canceled"),
			},
			{
				PropertyName: "rating",
				CurrentDataType: model.IntegerDataType,
				From: db.PointerFromString("500"),
				To: nil,
			},
		},
	}

	return raw, m
}

func createIncomingEventWithoutEmittedAt(now time.Time) (string, *model.Action) {
	raw := `{
		"targetId": "122242120",
		"targetEntity": "subscription",
		"targetService": "billing",
		"actorId": "88999",
		"actorEntity": "user",
		"actorService": "web",
		"eventName": "subscriptionCanceled",
		"changes": [
			{
				"propertyName": "status",
				"currentPropertyType": "string",
				"from": "active",
				"to": "canceled"
			},
			{
				"propertyName": "rating",
				"currentPropertyType": "integer",
				"from": "500",
				"to": null
			}
		]
	}`

	return raw, nil
}

