package rest

import (
	"fmt"

	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/clock"
	"github.com/denismitr/auditbase/utils/errbag"
	"github.com/denismitr/auditbase/utils/validator"
	"github.com/labstack/echo"
)

func extractIDParamFrom(ctx echo.Context) model.ID {
	id := ctx.Param("id")
	return model.ID(id)
}

type Change struct {
	PropertyName        string `json:"propertyName"`
	CurrentPropertyType string `json:"currentPropertyType"`
	From                string `json:"from"`
	To                  string `json:"to"`
}

func (c *Change) ToModel(eventID string) *model.PropertyChange {
	return &model.PropertyChange{
		PropertyName: c.PropertyName,
		EventID:      eventID,
		From:         interfaceToString(c.From),
		To:           interfaceToString(c.To),
	}
}

type CreateEvent struct {
	ID            string   `json:"id"`
	Operation     string   `json:"operation"`
	ActorID       string   `json:"actorId"`
	ActorEntity   string   `json:"actorEntity"`
	ActorService  string   `json:"actorService"`
	TargetID      string   `json:"targetId"`
	TargetEntity  string   `json:"targetEntity"`
	TargetService string   `json:"targetService"`
	EventName     string   `json:"eventName"`
	EmittedAt     int64    `json:"emittedAt"`
	RegisteredAt  int64    `json:"registeredAt"`
	Changes       []*Change `json:"changes"`
}

func (ce CreateEvent) Validate() *errbag.ErrorBag {
	eb := errbag.New()

	if !validator.IsEmptyString(ce.ID) && !validator.IsUUID4(ce.ID) {
		eb.Add("id", ErrInvalidUUID4)
	}

	if validator.IsEmptyString(ce.ActorID) {
		eb.Add("actorID", ErrActorIDEmpty)
	}

	if validator.IsEmptyString(ce.ActorEntity) {
		eb.Add("actorEntity", ErrActorEntityEmpty)
	}

	if validator.IsEmptyString(ce.TargetEntity) {
		eb.Add("targetEntity", ErrTargetEntityEmpty)
	}

	if validator.IsEmptyString(ce.ActorService) {
		eb.Add("actorService", ErrActorServiceEmpty)
	}

	if validator.IsEmptyString(ce.TargetService) {
		eb.Add("targetService", ErrTargetServiceEmpty)
	}

	if ce.EmittedAt == 0 {
		eb.Add("targetService", ErrTargetServiceEmpty)
	}

	for i := range ce.Changes {
		if validator.IsEmptyString(ce.Changes[i].PropertyName) {
			eb.Add(fmt.Sprintf("changes.%d", i), ErrEmptyPropertyName)
		}
	}

	return eb
}

func (ce CreateEvent) ToEvent() *model.Event {
	changes := make([]*model.PropertyChange, len(ce.Changes))

	for i := range ce.Changes {
		changes[i] = ce.Changes[i].ToModel(ce.ID)
	}

	return &model.Event{
		ID:      ce.ID,
		ActorID: ce.ActorID,
		ActorEntity: model.Entity{
			Name: ce.ActorEntity,
		},
		ActorService: model.Microservice{
			Name: ce.ActorService,
		},
		TargetID: ce.TargetID,
		TargetEntity: model.Entity{
			Name: ce.TargetEntity,
		},
		TargetService: model.Microservice{
			Name: ce.TargetService,
		},
		EventName:    ce.EventName,
		EmittedAt:    model.JSONTime{Time: clock.TimestampToTime(ce.EmittedAt)},
		Changes:      changes,
	}
}

func interfaceToString(value interface{}) *string {
	var out string

	switch raw := value.(type) {
	case string:
		out = raw
		return &out
	case int, int64:
		out = fmt.Sprintf("%d", raw)
		return &out
	case float64, float32:
		out = fmt.Sprintf("%0.4f", raw)
	default:
		return nil
	}

	return nil
}
