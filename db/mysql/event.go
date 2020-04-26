package mysql

import (
	"database/sql"
	"strings"
	"time"

	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

const createEvent = `
	INSERT INTO events (
		id, parent_event_id, hash, actor_id, 
		actor_entity_id, actor_service_id, target_id, 
		target_entity_id, target_service_id, event_name,
		emitted_at, registered_at
	) VALUES (
		UUID_TO_BIN(:id), UUID_TO_BIN(:parent_event_id), :hash, :actor_id, 
		UUID_TO_BIN(:actor_entity_id), UUID_TO_BIN(:actor_service_id), :target_id, 
		UUID_TO_BIN(:target_entity_id), UUID_TO_BIN(:target_service_id), :event_name, 
		:emitted_at, :registered_at
	)
`

const selectEvents = `
	SELECT 
		BIN_TO_UUID(e.id) as id, BIN_TO_UUID(parent_event_id) as parent_event_id,
		hash, actor_id, BIN_TO_UUID(actor_entity_id) as actor_entity_id, 
		BIN_TO_UUID(actor_service_id) as actor_service_id, 
		target_id, BIN_TO_UUID(target_entity_id) as target_entity_id, 
		BIN_TO_UUID(target_service_id) as target_service_id, 
		event_name, emitted_at, registered_at,
		ams.name as actor_service_name, tms.name as target_service_name,
		ams.description as actor_service_description, tms.description as target_service_description,
		ae.name as actor_entity_name, ae.description as actor_entity_description,
		te.name as target_entity_name, te.description as target_entity_description 
	FROM events as e
		INNER JOIN microservices as ams
	ON ams.id = e.actor_service_id
		INNER JOIN microservices as tms
	ON tms.id = e.target_service_id
		INNER JOIN entities as ae
	ON ae.id = e.actor_entity_id
		INNER JOIN entities as te
	ON te.id = e.target_entity_id
`

const selectEventProperties = `
	SELECT
		BIN_TO_UUID(id) as id, BIN_TO_UUID(event_id) as event_id,
		name, changed_from, changed_to
	FROM properties
		WHERE event_id = UUID_TO_BIN(?)
`

const selectEventListProperties = `
	SELECT
		BIN_TO_UUID(id) as id, BIN_TO_UUID(event_id) as event_id,
		name, changed_from, changed_to 
	FROM properties
		WHERE event_id IN (:eventIds)
`

const insertProperties = `
	INSERT INTO properties 
		(id, event_id, name, changed_from, changed_to)
	VALUES 
		(UUID_TO_BIN(?), UUID_TO_BIN(?), ?, ?, ?)
`

type event struct {
	ID                       string         `db:"id"`
	ParentEventID            sql.NullString `db:"parent_event_id"`
	Hash                     string         `db:"hash"`
	ActorID                  string         `db:"actor_id"`
	ActorEntityID            string         `db:"actor_entity_id"`
	ActorServiceID           string         `db:"actor_service_id"`
	ActorServiceName         string         `db:"actor_service_name"`
	ActorServiceDescription  string         `db:"actor_service_description"`
	ActorEntityName          string         `db:"actor_entity_name"`
	ActorEntityDescription   string         `db:"actor_entity_description"`
	TargetID                 string         `db:"target_id"`
	TargetEntityID           string         `db:"target_entity_id"`
	TargetEntityName         string         `db:"target_entity_name"`
	TargetEntityDescription  string         `db:"target_entity_description"`
	TargetServiceID          string         `db:"target_service_id"`
	TargetServiceName        string         `db:"target_service_name"`
	TargetServiceDescription string         `db:"target_service_description"`
	EventName                string         `db:"event_name"`
	EmittedAt                time.Time      `db:"emitted_at"`
	RegisteredAt             time.Time      `db:"registered_at"`
}

type EventRepository struct {
	conn  *sqlx.DB
	uuid4 uuid.UUID4Generator
}

func NewEventRepository(conn *sqlx.DB, uuid4 uuid.UUID4Generator) *EventRepository {
	return &EventRepository{
		conn:  conn,
		uuid4: uuid4,
	}
}

func (r *EventRepository) Create(e *model.Event) error {
	dbEvent := event{
		ID:              e.ID,
		Hash:            e.Hash,
		ActorID:         e.ActorID,
		ActorEntityID:   e.ActorEntity.ID,
		ActorServiceID:  e.ActorService.ID,
		TargetID:        e.TargetID,
		TargetEntityID:  e.TargetEntity.ID,
		TargetServiceID: e.TargetService.ID,
		EventName:       e.EventName,
		EmittedAt:       e.EmittedAt,
		RegisteredAt:    e.RegisteredAt,
	}

	if e.ParentEventID == "" {
		dbEvent.ParentEventID = sql.NullString{"", false}
	} else {
		dbEvent.ParentEventID = sql.NullString{e.ParentEventID, true}
	}

	tx, err := r.conn.Beginx()
	if err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	if _, err := tx.NamedExec(createEvent, &dbEvent); err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "could not insert new event with ID %s", e.ID)
	}

	for i := range e.Delta {
		id := r.uuid4.Generate()
		if _, err := tx.Exec(
			insertProperties,
			id,
			e.ID,
			e.Delta[i].Name,
			e.Delta[i].ChangedFrom,
			e.Delta[i].ChangedTo,
		); err != nil {
			tx.Rollback()
			return errors.Wrapf(err, "could not insert property for event with ID %s", e.ID)
		}
	}

	return tx.Commit()
}

func (r *EventRepository) Delete(ID model.ID) error {
	stmt := `DELETE FROM events WHERE id = UUID_TO_BIN(?)`

	if _, err := r.conn.Exec(stmt, ID.String()); err != nil {
		return errors.Wrapf(err, "could not delete event with ID %s", ID.String())
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

func (r *EventRepository) FindOneByID(ID model.ID) (*model.Event, error) {
	stmt := selectEvents + " WHERE e.id = UUID_TO_BIN(?)"

	e := event{}
	props := make([]property, 0)

	if err := r.conn.Get(&e, stmt, ID.String()); err != nil {
		return nil, errors.Wrapf(err, "could not get a list of events from db")
	}

	if err := r.conn.Select(&props, selectEventProperties, ID.String()); err != nil {
		return nil, errors.Wrapf(err, "could not get a list of properties for eve from db")
	}

	delta := make([]model.Property, len(props))

	for i := range props {
		delta[i] = model.Property{
			ID:      props[i].ID,
			EventID: props[i].EventID,
			Name:    props[i].Name,
		}

		if props[i].ChangedFrom.Valid == true {
			delta[i].ChangedFrom = &props[i].ChangedFrom.String
		}

		if props[i].ChangedTo.Valid == true {
			delta[i].ChangedTo = &props[i].ChangedTo.String
		}
	}

	// TODO: inner join name and description
	at := model.Entity{
		ID:          e.ActorEntityID,
		Name:        e.ActorEntityName,
		Description: e.ActorEntityDescription,
	}

	// TODO: inner join name and description
	tt := model.Entity{
		ID:          e.TargetEntityID,
		Name:        e.TargetEntityName,
		Description: e.TargetEntityDescription,
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

	return &model.Event{
		ID:            e.ID,
		ParentEventID: e.ParentEventID.String,
		ActorID:       e.ActorID,
		ActorEntity:   at,
		ActorService:  as,
		TargetID:      e.TargetID,
		TargetEntity:  tt,
		TargetService: ts,
		EventName:     e.EventName,
		EmittedAt:     e.EmittedAt,
		RegisteredAt:  e.RegisteredAt,
		Delta:         delta,
	}, nil
}

// Select events using filter, sort, and pagination
func (r *EventRepository) Select(
	filter model.EventFilter,
	sort model.Sort,
	pagination model.Pagination,
) ([]*model.Event, error) {
	q, args := prepareSelectEventsQueryWithArgs(filter, sort, pagination)

	events := []event{}
	result := make([]*model.Event, 0)

	stmt, err := r.conn.PrepareNamed(q)
	if err != nil {
		return nil, errors.Wrapf(err, "could not prepare select events stmt")
	}

	if err := stmt.Select(&events, args); err != nil {
		return nil, errors.Wrapf(err, "could not get a list of events from db")
	}

	props, err := r.joinPropertiesToEvents(events)
	if err != nil {
		return nil, errors.Wrap(err, "could not propertis to events")
	}

	for i := range events {
		at := model.Entity{
			ID:          events[i].ActorEntityID,
			Name:        events[i].ActorEntityName,
			Description: events[i].ActorEntityDescription,
		}

		tt := model.Entity{
			ID:          events[i].TargetEntityID,
			Name:        events[i].TargetEntityName,
			Description: events[i].TargetEntityDescription,
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

		var delta []model.Property

		if _, ok := props[events[i].ID]; ok {
			for j := range props[events[i].ID] {
				p := model.Property{
					ID:   props[events[i].ID][j].ID,
					Name: props[events[i].ID][j].Name,
				}

				if props[events[i].ID][j].ChangedFrom.Valid {
					p.ChangedFrom = &props[events[i].ID][j].ChangedFrom.String
				}

				if props[events[i].ID][j].ChangedTo.Valid {
					p.ChangedTo = &props[events[i].ID][j].ChangedTo.String
				}

				delta = append(delta, p)
			}
		}

		result = append(result, &model.Event{
			ID:            events[i].ID,
			ParentEventID: events[i].ParentEventID.String,
			ActorID:       events[i].ActorID,
			Hash:          events[i].Hash,
			ActorEntity:   at,
			ActorService:  as,
			TargetID:      events[i].TargetID,
			TargetEntity:  tt,
			TargetService: ts,
			EventName:     events[i].EventName,
			EmittedAt:     events[i].EmittedAt,
			RegisteredAt:  events[i].RegisteredAt,
			Delta:         delta,
		})
	}

	return result, nil
}

func (r *EventRepository) joinPropertiesToEvents(events []event) (map[string][]property, error) {
	var eventIds []string
	props := make(map[string][]property)

	if len(events) > 0 {
		for i := range events {
			eventIds = append(eventIds, "UUID_TO_BIN('"+events[i].ID+"')")
		}

		propStmt := strings.Replace(selectEventListProperties, ":eventIds", strings.Join(eventIds, ","), 1)

		rows, err := r.conn.Queryx(propStmt)
		if err != nil {
			return nil, errors.Wrapf(err, "could not get a list of properties for events with stmt %s", propStmt)
		}

		for rows.Next() {
			var p property
			rows.StructScan(&p)

			if _, ok := props[p.EventID]; !ok {
				props[p.EventID] = make([]property, 0)
			}

			props[p.EventID] = append(props[p.EventID], p)
		}
	}

	return props, nil
}

func prepareSelectEventsQueryWithArgs(
	filter model.EventFilter,
	sort model.Sort,
	pagination model.Pagination,
) (string, map[string]interface{}) {
	q := selectEvents
	args := make(map[string]interface{})

	if filter.ActorEntityID != "" {
		q += ` where actor_entity_id = UUID_TO_BIN(:actor_entity_id)`
		args["actor_entity_id"] = filter.ActorEntityID
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

	if filter.TargetEntityID != "" {
		q += ` where target_entity_id = UUID_TO_BIN(:target_entity_id)`
		args["target_entity_id"] = filter.TargetEntityID
	}

	if filter.TargetServiceID != "" {
		q += ` where target_service_id = UUID_TO_BIN(:target_service_id)`
		args["target_service_id"] = filter.TargetServiceID
	}

	return q, args
}
