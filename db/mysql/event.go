package mysql

import (
	"bytes"
	"database/sql"
	"github.com/denismitr/auditbase/utils/errtype"
	"github.com/denismitr/auditbase/utils/logger"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

const ErrEmptyWhereInList = errtype.StringError("WHERE IN clause cannot be empty, no values provided")

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

const selectEventListProperties = `
	SELECT
		BIN_TO_UUID(id) as id, BIN_TO_UUID(event_id) as event_id,
		BIN_TO_UUID(entity_id) as entity_id, name, changed_from, changed_to 
	FROM properties
		WHERE event_id IN (:eventIds)
`

const insertProperties = `
	INSERT INTO properties 
		(id, event_id, entity_id, name, changed_from, changed_to)
	VALUES 
		(UUID_TO_BIN(?), UUID_TO_BIN(?), UUID_TO_BIN(?), ?, ?, ?)
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
	log logger.Logger
	uuid4 uuid.UUID4Generator
}

func NewEventRepository(conn *sqlx.DB, uuid4 uuid.UUID4Generator, log logger.Logger) *EventRepository {
	return &EventRepository{
		conn:  conn,
		log: log,
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
		EmittedAt:       e.EmittedAt.Time,
		RegisteredAt:    e.RegisteredAt.Time,
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
		_ = tx.Rollback()
		return errors.Wrapf(err, "could not insert new events with ID %s", e.ID)
	}

	for i := range e.Changes {
		id := r.uuid4.Generate()

		_, err := sq.Insert("changes").
			Columns("id", "property_id", "event_id", "from_value", "to_value").
			Values(
				sq.Expr("UUID_TO_BIN(?)", id),
				sq.Expr("UUID_TO_BIN(?)", e.Changes[i].PropertyID),
				sq.Expr("UUID_TO_BIN(?)", e.ID),
				e.Changes[i].From,
				e.Changes[i].To,
			).
			RunWith(tx).Query()

		if err != nil {
			_ = tx.Rollback()
			return errors.Wrapf(err, "could not insert properties for events with ID %s", e.ID)
		}
	}

	return tx.Commit()
}

func (r *EventRepository) Delete(ID model.ID) error {
	stmt := `DELETE FROM events WHERE id = UUID_TO_BIN(?)`

	if _, err := r.conn.Exec(stmt, ID.String()); err != nil {
		return errors.Wrapf(err, "could not delete events with ID %s", ID.String())
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
	const selectChangeSet = `
		SELECT
			BIN_TO_UUID(id) as id, BIN_TO_UUID(event_id) as event_id,
			BIN_TO_UUID(property_id) as property_id, name, from_value, to_value
		FROM changes
			WHERE event_id = UUID_TO_BIN(?)
	`

	selectEvents := createBaseSelectEventsQuery().Where("e.id = UUID_TO_BIN(?)")
	stmt, _, err := selectEvents.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "could not build query")
	}

	e := event{}
	changes := make([]propertyChange, 0)

	if err := r.conn.Get(&e, stmt, ID.String()); err != nil {
		return nil, errors.Wrapf(err, "could not get a list of events from db")
	}

	if err := r.conn.Select(&changes, selectChangeSet, ID.String()); err != nil {
		return nil, errors.Wrapf(err, "could not get a list of properties for eve from db")
	}

	delta := make([]*model.PropertyChange, len(changes))

	for i := range changes {
		delta[i] = &model.PropertyChange{
			ID:      changes[i].ID,
			EventID: changes[i].EventID,
			PropertyName:    changes[i].PropertyName,
		}

		if changes[i].FromValue.Valid == true {
			delta[i].From = &changes[i].FromValue.String
		}

		if changes[i].ToValue.Valid == true {
			delta[i].To = &changes[i].ToValue.String
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
		EmittedAt:     model.JSONTime{Time: e.EmittedAt},
		RegisteredAt:  model.JSONTime{Time: e.RegisteredAt},
		Changes:         delta,
	}, nil
}

// Select events using filter, sort, and pagination
func (r *EventRepository) Select(
	filter *model.Filter,
	sort *model.Sort,
	pagination *model.Pagination,
) ([]*model.Event, *model.Meta, error) {
	q, err := prepareSelectEventsQueryWithArgs(filter, sort, pagination)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not create sql query")
	}

	r.log.SQL(q.query, q.queryArgs)
	r.log.SQL(q.count, q.countArgs)

	var events []event
	var meta meta
	result := make([]*model.Event, 0)

	ss, err := r.conn.PrepareNamed(q.query)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "could not prepare select events stmt")
	}

	cs, err := r.conn.PrepareNamed(q.count)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "could not prepare count events stmt")
	}

	if err := ss.Select(&events, q.queryArgs); err != nil {
		return nil, nil, errors.Wrapf(err, "could not get a list of events from db")
	}

	if err := cs.Get(&meta, q.countArgs); err != nil {
		return nil, nil, errors.Wrapf(err, "could not count events")
	}

	changes, err := r.joinChangesWithEvents(events)
	if err != nil {
		return nil, nil, err
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

		var changeSet []*model.PropertyChange

		if _, ok := changes[events[i].ID]; ok {
			for j := range changes[events[i].ID] {
				p := changes[events[i].ID][j].ToModel()
				changeSet = append(changeSet, p)
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
			EmittedAt:     model.JSONTime{Time: events[i].EmittedAt},
			RegisteredAt:  model.JSONTime{Time: events[i].RegisteredAt},
			Changes:       changeSet,
		})
	}

	return result, meta.ToModel(pagination), nil
}

func (r *EventRepository) joinChangesWithEvents(events []event) (map[string][]propertyChange, error) {
	var eventIds []string
	changes := make(map[string][]propertyChange)

	if len(events) == 0 {
		return changes, nil
	}

	for i := range events {
		eventIds = append(eventIds, events[i].ID)
	}

	q, args, err := createSelectChangesByIDsQuery(eventIds)
	if err != nil {
		return nil, err
	}

	r.log.Debugf("%s -- %#v", q, args)

	rows, err := r.conn.Queryx(q, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get a list of properties for events with stmt %s", q)
	}

	for rows.Next() {
		var p propertyChange
		_ = rows.StructScan(&p)

		if _, ok := changes[p.EventID]; !ok {
			changes[p.EventID] = make([]propertyChange, 0)
		}

		changes[p.EventID] = append(changes[p.EventID], p)
	}

	return changes, nil
}

func prepareSelectEventsQueryWithArgs(
	filter *model.Filter,
	sort *model.Sort,
	pagination *model.Pagination,
) (*selectWithMetaQuery, error) {
	selectEvents := createBaseSelectEventsQuery()
	countEvents := createBaseCountEventsQuery()
	args := make(map[string]interface{})

	if filter.Has("actorEntityId") {
		w := `actor_entity_id = UUID_TO_BIN(:actor_entity_id)`
		selectEvents = selectEvents.Where(w)
		countEvents = countEvents.Where(w)
		args["actor_entity_id"] = filter.MustString("actorEntityId")
	}

	if filter.Has("actorId") {
		w := `actor_id = :actor_id`
		selectEvents = selectEvents.Where(w)
		countEvents = countEvents.Where(w)
		args["actor_id"] = filter.MustString("actorId")
	}

	if filter.Has("actorServiceId") {
		w := `actor_service_id = UUID_TO_BIN(:actor_service_id)`
		selectEvents = selectEvents.Where(w)
		countEvents = countEvents.Where(w)
		args["actor_service_id"] = filter.MustString("actorServiceId")
	}

	if filter.Has("targetId") {
		w:= `actor_service_id = UUID_TO_BIN(:actor_service_id)`
		selectEvents = selectEvents.Where(w)
		countEvents = countEvents.Where(w)
		args["target_id"] = filter.MustString("targetId")
	}

	if filter.Has("targetEntityId") {
		w := `target_entity_id = UUID_TO_BIN(:target_entity_id)`
		selectEvents = selectEvents.Where(w)
		countEvents = countEvents.Where(w)
		args["target_entity_id"] = filter.MustString("targetEntityId")
	}

	if filter.Has("targetServiceId") {
		w := `target_service_id = UUID_TO_BIN(:target_service_id)`
		selectEvents = selectEvents.Where(w)
		countEvents = countEvents.Where(w)
		args["target_service_id"] = filter.MustString("targetServiceId")
	}

	if filter.Has("eventName") {
		w := `event_name = :event_name`
		selectEvents = selectEvents.Where(w)
		countEvents = countEvents.Where(w)
		args["event_name"] = filter.MustString("eventName")
	}

	if sort.Empty() {
		selectEvents = selectEvents.OrderBy("emitted_at DESC")
	}

	if pagination.Page > 0 && pagination.PerPage > 0 {
		selectEvents = selectEvents.Limit(uint64(pagination.PerPage)).Offset(uint64(pagination.Offset()))
	}

	selectEventsQuery, _, err := selectEvents.ToSql()
	if err != nil {
		return nil, err
	}

	countEventsQuery, _, err := countEvents.ToSql()
	if err != nil {
		return nil, err
	}

	return &selectWithMetaQuery{
		query:     selectEventsQuery,
		count:     countEventsQuery,
		queryArgs: args,
		countArgs: args,
	}, nil
}

func createBaseSelectEventsQuery() sq.SelectBuilder {
	query := sq.Select(
		"BIN_TO_UUID(e.id) as id",
		"BIN_TO_UUID(parent_event_id) as parent_event_id",
		"BIN_TO_UUID(actor_entity_id) as actor_entity_id",
		"BIN_TO_UUID(actor_service_id) as actor_service_id",
		"BIN_TO_UUID(target_entity_id) as target_entity_id",
		"BIN_TO_UUID(target_service_id) as target_service_id",
		"hash", "actor_id", "target_id", "event_name",
		"emitted_at", "registered_at",
		"ams.name as actor_service_name",
		"tms.name as target_service_name",
		"ams.description as actor_service_description",
		"tms.description as target_service_description",
		"ae.name as actor_entity_name",
		"ae.description as actor_entity_description",
		"te.name as target_entity_name",
		"te.description as target_entity_description",
	)

	query = query.From("events as e")
	query = query.Join("microservices as ams ON ams.id = e.actor_service_id")
	query = query.Join("microservices as tms ON tms.id = e.target_service_id")
	query = query.Join("entities as ae ON ae.id = e.actor_entity_id")
	query = query.Join("entities as te ON te.id = e.target_entity_id")
	return query
}

func createBaseCountEventsQuery() sq.SelectBuilder {
	return sq.Select("COUNT(*) as total FROM events e")
}

func createSelectChangesByIDsQuery(ids []string) (string, []interface{}, error) {
	if len(ids) == 0 {
		return "", nil, ErrEmptyWhereInList
	}

	q := sq.Select(
		"BIN_TO_UUID(c.id) as id",
		"BIN_TO_UUID(c.property_id) as property_id",
		"BIN_TO_UUID(c.event_id) as event_id",
		"BIN_TO_UUID(p.entity_id) as entity_id",
		"p.name as property_name",
		"from_value", "to_value",
	)

	q = q.From("changes as c")
	q = q.Join("properties as p ON p.id = c.property_id")

	var expr bytes.Buffer
	var args []interface{}
	expr.WriteString("event_id IN (")
	for i := range ids {
		expr.WriteString("UUID_TO_BIN(?)")
		if i + 1 < len(ids) {
			expr.WriteString(",")
		}

		args = append(args, ids[i])
	}
	expr.WriteString(")")

	return q.Where(expr.String(), args...).ToSql()
}
