package mysql

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/denismitr/auditbase/model"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	"github.com/pkg/errors"
)

type event struct {
	ID              string         `db:"id"`
	ParentEventID   sql.NullString `db:"parent_event_id"`
	ActorID         string         `db:"actor_id"`
	ActorTypeID     string         `db:"actor_type_id"`
	ActorServiceID  string         `db:"actor_service_id"`
	TargetID        string         `db:"target_id"`
	TargetTypeID    string         `db:"target_type_id"`
	TargetServiceID string         `db:"target_service_id"`
	EventName       string         `db:"event_name"`
	EmittedAt       time.Time      `db:"emitted_at"`
	RegisteredAt    time.Time      `db:"registered_at"`
	Delta           types.JSONText `db:"delta"`
}

type EventRepository struct {
	Conn *sqlx.DB
}

func (r *EventRepository) Create(e model.Event) error {
	stmt := `
		INSERT INTO events (
			id, parent_event_id, actor_id, 
			actor_type_id, actor_service_id, target_id, 
			target_type_id, target_service_id, event_name,
			emitted_at, registered_at, delta
		) VALUES (
			UUID_TO_BIN(:id), UUID_TO_BIN(:parent_event_id), :actor_id, 
			UUID_TO_BIN(:actor_type_id), :actor_service_id, :target_id, 
			UUID_TO_BIN(:target_type_id), :target_service_id, :event_name, 
			:emitted_at, :registered_at, :delta
		)
	`

	jsBytes, err := json.Marshal(e.Delta)
	if err != nil {
		return errors.Wrap(err, "could not serialize DELTA")
	}

	dbEvent := event{
		ID:              e.ID,
		ActorID:         e.ActorID,
		ActorTypeID:     e.ActorType.ID,
		ActorServiceID:  e.ActorServiceID,
		TargetID:        e.TargetID,
		TargetTypeID:    e.TargetType.ID,
		TargetServiceID: e.TargetServiceID,
		EventName:       e.EventName,
		EmittedAt:       time.Unix(e.EmittedAt, 0),
		RegisteredAt:    time.Unix(e.RegisteredAt, 0),
		Delta:           types.JSONText(jsBytes),
	}

	if e.ParentEventID == "" {
		dbEvent.ParentEventID = sql.NullString{"", false}
	} else {
		dbEvent.ParentEventID = sql.NullString{e.ParentEventID, true}
	}

	if _, err := r.Conn.NamedExec(stmt, &dbEvent); err != nil {
		return errors.Wrapf(err, "could not insert new event with ID %s", e.ID)
	}

	return nil
}

func (r *EventRepository) Delete(ID string) error {
	stmt := `DELETE FROM events WHERE id = UUID_TO_BIN(?)`

	if _, err := r.Conn.Exec(stmt, ID); err != nil {
		return errors.Wrapf(err, "could not delete event with ID %s", ID)
	}

	return nil
}

func (r *EventRepository) FindOneByID(ID string) (model.Event, error) {
	stmt := `
		SELECT 
			BIN_TO_UUID(id) as id, BIN_TO_UUID(parent_event_id) as parent_event_id,
			actor_id, BIN_TO_UUID(actor_type_id) as actor_type_id, actor_service_id, 
			target_id, BIN_TO_UUID(target_type_id) as target_type_id, target_service_id,
			event_name, emitted_at, registered_at, delta 
		FROM events WHERE id = ?
	`

	e := event{}

	if err := r.Conn.Select(&e, stmt); err != nil {
		return model.Event{}, errors.Wrapf(err, "could not get a list of events from db")
	}

	var d map[string][]interface{}
	json.Unmarshal(e.Delta, &d)

	// TODO: inner join name and description
	at := model.ActorType{
		ID: e.ActorTypeID,
	}

	// TODO: inner join name and description
	tt := model.TargetType{
		ID: e.TargetTypeID,
	}

	return model.Event{
		ID:              e.ID,
		ParentEventID:   e.ParentEventID.String,
		ActorID:         e.ActorID,
		ActorType:       at,
		ActorServiceID:  e.ActorServiceID,
		TargetID:        e.TargetID,
		TargetType:      tt,
		TargetServiceID: e.TargetServiceID,
		EventName:       e.EventName,
		EmittedAt:       e.EmittedAt.Unix(),
		RegisteredAt:    e.RegisteredAt.Unix(),
		Delta:           d,
	}, nil
}

func (r *EventRepository) SelectAll() ([]model.Event, error) {
	stmt := `
		SELECT 
			BIN_TO_UUID(id) as id, BIN_TO_UUID(parent_event_id) as parent_event_id,
			actor_id, BIN_TO_UUID(actor_type_id) as actor_type_id, actor_service_id, 
			target_id, BIN_TO_UUID(target_type_id) as target_type_id, target_service_id, 
			event_name, emitted_at, registered_at, delta 
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

		// TODO: inner join name and description
		at := model.ActorType{
			ID: events[i].ActorTypeID,
		}

		// TODO: inner join name and description
		tt := model.TargetType{
			ID: events[i].TargetTypeID,
		}

		result[i] = model.Event{
			ID:              events[i].ID,
			ParentEventID:   events[i].ParentEventID.String,
			ActorID:         events[i].ActorID,
			ActorType:       at,
			ActorServiceID:  events[i].ActorServiceID,
			TargetID:        events[i].TargetID,
			TargetType:      tt,
			TargetServiceID: events[i].TargetServiceID,
			EventName:       events[i].EventName,
			EmittedAt:       events[i].EmittedAt.Unix(),
			RegisteredAt:    events[i].RegisteredAt.Unix(),
			Delta:           d,
		}
	}

	return result, nil
}
