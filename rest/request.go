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

type createEvent struct {
	ID            string                   `json:"id"`
	ActorID       string                   `json:"actorId"`
	ActorEntity   string                   `json:"actorEntity"`
	ActorService  string                   `json:"actorService"`
	TargetID      string                   `json:"targetId"`
	TargetEntity  string                   `json:"targetEntity"`
	TargetService string                   `json:"targetService"`
	EventName     string                   `json:"eventName"`
	EmittedAt     int64                    `json:"emittedAt"`
	RegisteredAt  int64                    `json:"registeredAt"`
	Delta         map[string][]interface{} `json:"delta"`
}

func (ce createEvent) Validate() *errbag.ErrorBag {
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

	return eb
}

func (ce createEvent) ToEvent() *model.Event {
	props := make([]model.Property, len(ce.Delta))

	i := 0
	for k, values := range ce.Delta {
		if len(values) != 2 {
			continue
		}

		changedTo := interfaceToString(values[1])
		changedFrom := interfaceToString(values[0])

		props[i] = model.Property{
			Name:        k,
			EventID:     ce.ID,
			ChangedFrom: changedFrom,
			ChangedTo:   changedTo,
		}

		i++
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
		EmittedAt:    clock.TimestampToTime(ce.EmittedAt),
		RegisteredAt: clock.TimestampToTime(ce.RegisteredAt),
		Delta:        props,
	}
}

func interfaceToString(value interface{}) *string {
	var out string

	switch raw := value.(type) {
	case string:
		out = raw
		return &out
	case int:
		out = fmt.Sprintf("%d", raw)
		return &out
	case float64, float32:
		out = fmt.Sprintf("%0.4f", raw)
	default:
		return nil
	}

	return nil
}
