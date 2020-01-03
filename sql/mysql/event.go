package mysql

import (
	"encoding/json"
	"fmt"

	"github.com/denismitr/auditbase/model"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	"github.com/pkg/errors"
)

type event struct {
	ID              string         `db:"id"`
	ParentEventID   string         `db:"parent_event_id"`
	ActorID         string         `db:"actor_id"`
	ActorType       string         `db:"actor_type"`
	ActorServiceID  string         `db:"actor_service_id"`
	TargetID        string         `db:"target_id"`
	TargetType      string         `db:"target_type"`
	TargetServiceID string         `db:"target_service_id"`
	EventName       string         `db:"event_name"`
	EmittedAt       string         `db:"emitted_at"`
	RegisteredAt    string         `db:"registered_at"`
	Delta           types.JSONText `db:"delta"`
}

type EventRepository struct {
	Conn *sqlx.DB
}

func (r *EventRepository) Create(e model.Event) error {
	stmt := `
		INSERT INTO events (
			id, parent_event_id, actor_id, 
			actor_type, actor_service_id, target_id, 
			target_type, target_service_id, event_name,
			emitted_at, registered_at, delta
		) VALUES (
			UUID_TO_BIN(:id), UUID_TO_BIN(:parent_event_id), :actor_id, 
			:actor_type, UUID_TO_BIN(:actor_service_id), :target_id, :target_type, 
			UUID_TO_BIN(:target_service_id), :event_name, :emitted_at, 
			:registered_at, :delta
		)
	`

	jsBytes, _ := json.Marshal(e.Delta)
	// js := types.JSONText{}
	// js.MarshalJSON()

	dbEvent := event{
		ID:              e.ID,
		ParentEventID:   e.ParentEventID,
		ActorID:         e.ActorID,
		ActorType:       e.ActorType,
		ActorServiceID:  e.ActorServiceID,
		TargetID:        e.TargetID,
		TargetType:      e.TargetType,
		TargetServiceID: e.TargetServiceID,
		EventName:       e.EventName,
		EmittedAt:       e.EmittedAt,
		RegisteredAt:    e.RegisteredAt,
		Delta:           types.JSONText(jsBytes),
	}
	fmt.Printf("\n%#v", dbEvent)
	if _, err := r.Conn.NamedExec(stmt, &dbEvent); err != nil {
		return errors.Wrapf(err, "could not insert new event with ID %s", e.ID)
	}

	return nil
}

func (r *EventRepository) Update(int, model.Event) error {
	return nil
}

func (r *EventRepository) Delete(int) error {
	return nil
}

func (r *EventRepository) FindOneByID(int) (model.Event, error) {
	return model.Event{}, nil
}

func (r *EventRepository) SelectAll() ([]model.Event, error) {
	stmt := `
		SELECT 
			BIN_TO_UUID(id) as id, BIN_TO_UUID(parent_event_id) as parent_event_id,
			actor_id, actor_type, BIN_TO_UUID(actor_service_id), target_id,
			target_type, BIN_TO_UUID(target_service_id), event_name, 
			emitted_at, registered_at, delta 
		FROM events
	`
	events := []event{}

	if err := r.Conn.Select(&events, stmt); err != nil {
		return []model.Event{}, errors.Wrapf(err, "could not get a list of events from db")
	}

	result := make([]model.Event, len(events))

	for i := range events {
		var d map[string][]interface{}
		json.Unmarshal(events[i].Delta, &d)
		result[i] = model.Event{
			ID:              events[i].ID,
			ParentEventID:   events[i].ParentEventID,
			ActorID:         events[i].ActorID,
			ActorType:       events[i].ActorType,
			ActorServiceID:  events[i].ActorServiceID,
			TargetID:        events[i].TargetID,
			TargetType:      events[i].TargetType,
			TargetServiceID: events[i].TargetServiceID,
			EventName:       events[i].EventName,
			EmittedAt:       events[i].EmittedAt,
			RegisteredAt:    events[i].RegisteredAt,
			Delta:           d,
		}
	}

	return result, nil
}
