package mysql

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	"github.com/pkg/errors"
)

type event struct {
	ID                       string         `db:"id"`
	ParentEventID            sql.NullString `db:"parent_event_id"`
	ActorID                  string         `db:"actor_id"`
	ActorTypeID              string         `db:"actor_type_id"`
	ActorServiceID           string         `db:"actor_service_id"`
	ActorServiceName         string         `db:"actor_service_name"`
	ActorServiceDescription  string         `db:"actor_service_description"`
	ActorTypeName            string         `db:"actor_type_name"`
	ActorTypeDescription     string         `db:"actor_type_description"`
	TargetID                 string         `db:"target_id"`
	TargetTypeID             string         `db:"target_type_id"`
	TargetTypeName           string         `db:"target_type_name"`
	TargetTypeDescription    string         `db:"target_type_description"`
	TargetServiceID          string         `db:"target_service_id"`
	TargetServiceName        string         `db:"target_service_name"`
	TargetServiceDescription string         `db:"target_service_description"`
	EventName                string         `db:"event_name"`
	EmittedAt                time.Time      `db:"emitted_at"`
	RegisteredAt             time.Time      `db:"registered_at"`
	Delta                    types.JSONText `db:"delta"`
}

type EventRepository struct {
	conn  *sqlx.DB
	uuid4 utils.UUID4Generatgor
}

func NewEventRepository(conn *sqlx.DB, uuid4 utils.UUID4Generatgor) *EventRepository {
	return &EventRepository{
		conn:  conn,
		uuid4: uuid4,
	}
}

func (r *EventRepository) Create(e *model.Event) error {
	jsBytes, err := json.Marshal(e.Delta)
	if err != nil {
		return errors.Wrap(err, "could not serialize DELTA")
	}

	dbEvent := event{
		ID:              e.ID,
		ActorID:         e.ActorID,
		ActorTypeID:     e.ActorType.ID,
		ActorServiceID:  e.ActorService.ID,
		TargetID:        e.TargetID,
		TargetTypeID:    e.TargetType.ID,
		TargetServiceID: e.TargetService.ID,
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

	if _, err := r.conn.NamedExec(createEvent, &dbEvent); err != nil {
		return errors.Wrapf(err, "could not insert new event with ID %s", e.ID)
	}

	return nil
}

func (r *EventRepository) Delete(ID string) error {
	stmt := `DELETE FROM events WHERE id = UUID_TO_BIN(?)`

	if _, err := r.conn.Exec(stmt, ID); err != nil {
		return errors.Wrapf(err, "could not delete event with ID %s", ID)
	}

	return nil
}

func (r *EventRepository) Count() (int, error) {
	stmt := `SELECT COUNT(*) FROM events`
	var count int

	if err := r.conn.Get(&count, stmt); err != nil {
		return 0, errors.Wrap(err, "could not count events")
	}

	return count, nil
}

func (r *EventRepository) FindOneByID(ID model.ID) (model.Event, error) {
	stmt := selectEvents + " WHERE e.id = UUID_TO_BIN(?)"

	e := event{}

	if err := r.conn.Get(&e, stmt, ID.String()); err != nil {
		return model.Event{}, errors.Wrapf(err, "could not get a list of events from db")
	}

	var d map[string][]interface{}
	json.Unmarshal(e.Delta, &d)

	// TODO: inner join name and description
	at := model.ActorType{
		ID:          e.ActorTypeID,
		Name:        e.ActorTypeName,
		Description: e.ActorTypeDescription,
	}

	// TODO: inner join name and description
	tt := model.TargetType{
		ID:          e.TargetTypeID,
		Name:        e.TargetTypeName,
		Description: e.TargetTypeDescription,
	}

	as := model.Microservice{
		ID:          e.ActorServiceID,
		Name:        e.ActorServiceName,
		Description: e.ActorServiceDescription,
	}

	ts := model.Microservice{
		ID:          e.TargetServiceID,
		Name:        e.TargetServiceName,
		Description: e.TargetServiceDescription,
	}

	return model.Event{
		ID:            e.ID,
		ParentEventID: e.ParentEventID.String,
		ActorID:       e.ActorID,
		ActorType:     at,
		ActorService:  as,
		TargetID:      e.TargetID,
		TargetType:    tt,
		TargetService: ts,
		EventName:     e.EventName,
		EmittedAt:     e.EmittedAt.Unix(),
		RegisteredAt:  e.RegisteredAt.Unix(),
		Delta:         d,
	}, nil
}

// Select events using filter, sort, and pagination
func (r *EventRepository) Select(
	filter model.EventFilter,
	sort model.Sort,
	pagination model.Pagination,
) ([]model.Event, error) {
	q := selectEvents
	args := make(map[string]interface{})

	if filter.ActorTypeID != "" {
		q += ` where actor_type_id = UUID_TO_BIN(:actor_type_id)`
		args["actor_type_id"] = filter.ActorTypeID
	}

	if filter.ActorID != "" {
		q += ` where actor_id = :actor_id`
		args["actor_id"] = filter.ActorID
	}

	if filter.ActorServiceID != "" {
		q += ` where actor_service_id = UUID_TO_BIN(:actor_service_id)`
		args["actor_service_id"] = filter.ActorServiceID
	}

	if filter.TargetID != "" {
		q += ` where target_id = :target_id`
		args["target_id"] = filter.TargetID
	}

	if filter.TargetTypeID != "" {
		q += ` where target_type_id = UUID_TO_BIN(:target_type_id)`
		args["target_type_id"] = filter.TargetTypeID
	}

	if filter.TargetServiceID != "" {
		q += ` where target_service_id = UUID_TO_BIN(:target_service_id)`
		args["target_service_id"] = filter.TargetServiceID
	}

	events := []event{}
	result := []model.Event{}

	stmt, err := r.conn.PrepareNamed(q)
	if err != nil {
		return result, errors.Wrapf(err, "could not prepare select events stmt")
	}

	if err := stmt.Select(&events, args); err != nil {
		return result, errors.Wrapf(err, "could not get a list of events from db")
	}

	for i := range events {
		var d map[string][]interface{}
		json.Unmarshal(events[i].Delta, &d)

		at := model.ActorType{
			ID:          events[i].ActorTypeID,
			Name:        events[i].ActorTypeName,
			Description: events[i].ActorTypeDescription,
		}

		tt := model.TargetType{
			ID:          events[i].TargetTypeID,
			Name:        events[i].TargetTypeName,
			Description: events[i].TargetTypeDescription,
		}

		as := model.Microservice{
			ID:          events[i].ActorServiceID,
			Name:        events[i].ActorServiceName,
			Description: events[i].ActorServiceDescription,
		}

		ts := model.Microservice{
			ID:          events[i].TargetServiceID,
			Name:        events[i].TargetServiceName,
			Description: events[i].TargetServiceDescription,
		}

		result = append(result, model.Event{
			ID:            events[i].ID,
			ParentEventID: events[i].ParentEventID.String,
			ActorID:       events[i].ActorID,
			ActorType:     at,
			ActorService:  as,
			TargetID:      events[i].TargetID,
			TargetType:    tt,
			TargetService: ts,
			EventName:     events[i].EventName,
			EmittedAt:     events[i].EmittedAt.Unix(),
			RegisteredAt:  events[i].RegisteredAt.Unix(),
			Delta:         d,
		})
	}

	return result, nil
}
