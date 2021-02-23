package rest

import (
	"fmt"
	"strings"

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
	PropertyName string      `json:"propertyName"`
	From         interface{} `json:"from"`
	To           interface{} `json:"to"`
}

func (c *Change) ToModel(eventID string) *model.PropertyChange {
	from := interfaceToStringPointer(c.From)
	to := interfaceToStringPointer(c.To)
	return &model.PropertyChange{
		PropertyName:    c.PropertyName,
		EventID:         eventID,
		From:            from,
		To:              to,
		CurrentDataType: guessPairDataType(from, to),
	}
}

type CreateEvent struct {
	ID            string    `json:"id"`
	Crud          int       `json:"operation"`
	ActorID       string    `json:"actorId"`
	ActorEntity   string    `json:"actorEntity"`
	ActorService  string    `json:"actorService"`
	TargetID      string    `json:"targetId"`
	TargetEntity  string    `json:"targetEntity"`
	TargetService string    `json:"targetService"`
	EventName     string    `json:"eventName"`
	EmittedAt     int64     `json:"emittedAt"`
	RegisteredAt  int64     `json:"registeredAt"`
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
		eb.Add("emittedAt", ErrEmittedAtEmpty)
	}

	for i := range ce.Changes {
		if validator.IsEmptyString(ce.Changes[i].PropertyName) {
			eb.Add(fmt.Sprintf("changes.%d", i), ErrEmptyPropertyName)
		}
	}

	return eb
}

func (ce CreateEvent) ToEvent() *model.Action {
	changes := make([]*model.PropertyChange, len(ce.Changes))

	for i := range ce.Changes {
		changes[i] = ce.Changes[i].ToModel(ce.ID)
	}

	return &model.Action{
		ID:            ce.ID,
		ActorEntityID: ce.ActorID,
		ActorEntity: model.Entity{
			Name: ce.ActorEntity,
		},
		ActorService: model.Microservice{
			Name: ce.ActorService,
		},
		TargetEntityID: ce.TargetID,
		TargetEntity: model.Entity{
			Name: ce.TargetEntity,
		},
		TargetService: model.Microservice{
			Name: ce.TargetService,
		},
		EventName: ce.EventName,
		EmittedAt: model.JSONTime{Time: clock.TimestampToTime(ce.EmittedAt)},
		Changes:   changes,
	}
}

func interfaceToStringPointer(value interface{}) *string {
	var out string

	switch raw := value.(type) {
	case string:
		out = raw
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		out = fmt.Sprintf("%d", raw)
	case float64, float32:
		out = strings.TrimRight(fmt.Sprintf("%0.4f", raw), "0")
	case bool:
		trueOrFalse := value.(bool)
		if !trueOrFalse {
			out = "0"
		} else {
			out = "1"
		}
	default:
		return nil
	}

	if out != "" {
		return &out
	}

	return nil
}

func guessPairDataType(from, to *string) model.DataType {
	if from == nil {
		return guessDataType(to)
	}

	if to == nil {
		return guessDataType(from)
	}

	// if length is great lets not waist time on regex - it's most probably a string
	if len(*from) > 150 && len(*to) > 150 {
		return model.StringDataType
	}

	if validator.IsFloat(*from) || validator.IsFloat(*to) {
		return model.FloatDataType
	}

	if validator.IsInteger(*from) && validator.IsInteger(*to) {
		return model.IntegerDataType
	}

	return model.StringDataType
}

func guessDataType(s *string) model.DataType {
	if s == nil {
		return model.NullDataType
	}

	// if length is great lets not waist time on regex - it's most probably a string
	if len(*s) > 150 {
		return model.StringDataType
	}

	if validator.IsFloat(*s) {
		return model.FloatDataType
	}

	if validator.IsInteger(*s) {
		return model.IntegerDataType
	}

	return model.StringDataType
}
