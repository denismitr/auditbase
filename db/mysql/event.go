package mysql

import (
	"database/sql"
	"github.com/denismitr/auditbase/db"
	"github.com/denismitr/auditbase/utils/errtype"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/denismitr/auditbase/utils/validator"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

const ErrEmptyWhereInList = errtype.StringError("WHERE IN clause cannot be empty, no values provided")

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

type insertEvent struct {
	ID              string         `db:"id"`
	ParentEventID   sql.NullString `db:"parent_event_id"`
	Hash            string         `db:"hash"`
	ActorID         string         `db:"actor_id"`
	ActorEntityID   string         `db:"actor_entity_id"`
	ActorServiceID  string         `db:"actor_service_id"`
	TargetID        string         `db:"target_id"`
	TargetEntityID  string         `db:"target_entity_id"`
	TargetServiceID string         `db:"target_service_id"`
	EventName       string         `db:"event_name"`
	EmittedAt       time.Time      `db:"emitted_at"`
	RegisteredAt    time.Time      `db:"registered_at"`
}

type EventRepository struct {
	conn  *sqlx.DB
	log   logger.Logger
	uuid4 uuid.UUID4Generator
}

func NewEventRepository(conn *sqlx.DB, uuid4 uuid.UUID4Generator, log logger.Logger) *EventRepository {
	return &EventRepository{
		conn:  conn,
		log:   log,
		uuid4: uuid4,
	}
}

func (r *EventRepository) Create(e *model.Event) error {
	ie := insertEvent{
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

	ie.ParentEventID = db.NullStringFromStringPointer(e.ParentEventID)

	createSQL, args, err := createEventQuery(&ie)
	if err != nil {
		return errors.Wrap(err, "could not build create event query")
	}

	tx, err := r.conn.Beginx()
	if err != nil {
		return errors.Wrap(err, "could not begin transaction")
	}

	createStmt, err := tx.Preparex(createSQL)
	if err != nil {
		_ = tx.Rollback()
		return errors.Wrapf(err, "could not prepare create event query %s", createSQL)
	}

	if _, err := createStmt.Exec(args...); err != nil {
		_ = tx.Rollback()
		return errors.Wrapf(err, "could not create new event with ID %s", e.ID)
	}

	for i := range e.Changes {
		c := change{
			ID:              r.uuid4.Generate(),
			PropertyID:      e.Changes[i].PropertyID,
			EventID:         e.ID,
			FromValue:       db.NullStringFromStringPointer(e.Changes[i].From),
			ToValue:         db.NullStringFromStringPointer(e.Changes[i].To),
			CurrentDataType: int(e.Changes[i].CurrentDataType),
		}

		changeSQL, args, err := createChangeQuery(&c)
		if err != nil {
			_ = tx.Rollback()
			return errors.Wrapf(err, "could not create insert change query for event with ID %s", e.ID)
		}

		if _, err := tx.Exec(changeSQL, args...); err != nil {
			_ = tx.Rollback()
			return errors.Wrapf(err, "could not insert property change for event with ID %s", e.ID)
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
	q, args, err := countEventsQuery()
	if err != nil {
		return 0, errors.Wrap(err, "could not build count events query")
	}

	stmt, err := r.conn.Preparex(q)
	if err != nil {
		return 0, errors.Wrapf(err, "could not prepare count events query %s", q)
	}

	r.log.SQL(q, args)
	var count int

	if err := stmt.Get(&count, args...); err != nil {
		return 0, errors.Wrap(err, "could not count events")
	}

	return count, nil
}

func (r *EventRepository) FindOneByID(ID model.ID) (*model.Event, error) {
	q, args, err := selectOneEventQuery(ID.String())
	if err != nil {
		return nil, errors.Wrap(err, "could not build select one event query")
	}

	cq, cqArgs, err := selectChangesByEventIDQuery(ID.String())
	if err != nil {
		return nil, errors.Wrap(err, "could not build select event changes query")
	}

	r.log.SQL(q, args)
	r.log.SQL(cq, cqArgs)

	tx, err := r.conn.Beginx()
	if err != nil {
		return nil, errors.Wrap(err, "could not begin transaction")
	}

	stmt, err := tx.Preparex(q)
	if err != nil {
		_ = tx.Rollback()
		return nil, errors.Wrap(err, "could not prepare  select one event query")
	}

	cqStmt, err := tx.Preparex(cq)
	if err != nil {
		_ = tx.Rollback()
		return nil, errors.Wrap(err, "could not prepare select event changes query")
	}

	e := event{}
	var changes []propertyChange

	if err := stmt.Get(&e, args...); err != nil {
		_ = tx.Rollback()
		return nil, errors.Wrapf(err, "could not get event with ID %s from db", ID.String())
	}

	if err := cqStmt.Select(&changes, cqArgs...); err != nil {
		_ = tx.Rollback()
		return nil, errors.Wrapf(err, "could not get a list of changes for event from db")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrapf(err, db.ErrCouldNotCommit.Error())
	}

	delta := make([]*model.PropertyChange, len(changes))

	for i := range changes {
		delta[i] = changes[i].ToModel()
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
		ParentEventID: &e.ParentEventID.String,
		ActorID:       e.ActorID,
		ActorEntity:   at,
		ActorService:  as,
		TargetID:      e.TargetID,
		TargetEntity:  tt,
		TargetService: ts,
		EventName:     e.EventName,
		EmittedAt:     model.JSONTime{Time: e.EmittedAt},
		RegisteredAt:  model.JSONTime{Time: e.RegisteredAt},
		Changes:       delta,
	}, nil
}

// Select events using filter, sort, and pagination
func (r *EventRepository) Select(
	filter *model.Filter,
	sort *model.Sort,
	pagination *model.Pagination,
) ([]*model.Event, *model.Meta, error) {
	q, err := selectEventsQuery(filter, sort, pagination)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not create sql query")
	}

	r.log.SQL(q.selectSQL, q.selectArgs)
	r.log.SQL(q.countSQL, q.countArgs)

	var events []event
	var meta meta
	result := make([]*model.Event, 0)

	selectStmt, err := r.conn.Preparex(q.selectSQL)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "could not prepare select events stmt")
	}

	cs, err := r.conn.Preparex(q.countSQL)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "could not prepare count events stmt")
	}

	if err := selectStmt.Select(&events, q.selectArgs...); err != nil {
		return nil, nil, errors.Wrapf(err, "could not get a list of events from db")
	}

	if err := cs.Get(&meta, q.countArgs...); err != nil {
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
			ParentEventID: &events[i].ParentEventID.String,
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

	q, args, err := selectChangesByEventIDsQuery(eventIds)
	if err != nil {
		return nil, err
	}

	r.log.SQL(q, args)

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

func selectEventsQuery(
	filter *model.Filter,
	sort *model.Sort,
	pagination *model.Pagination,
) (*selectQuery, error) {
	selectEvents := baseSelectEventsQuery()
	countEvents := createBaseCountEventsQuery()

	if filter.Has("actorEntityId") {
		w := `actor_entity_id = UUID_TO_BIN(?)`
		selectEvents = selectEvents.Where(w, filter.MustString("actorEntityId"))
		countEvents = countEvents.Where(w, filter.MustString("actorEntityId"))
	}

	if filter.Has("actorId") {
		w := `actor_id = ?`
		selectEvents = selectEvents.Where(w, filter.MustString("actorId"))
		countEvents = countEvents.Where(w, filter.MustString("actorId"))
	}

	if filter.Has("actorServiceId") {
		w := `actor_service_id = UUID_TO_BIN(?)`
		selectEvents = selectEvents.Where(w, filter.MustString("actorServiceId"))
		countEvents = countEvents.Where(w, filter.MustString("actorServiceId"))
	}

	if filter.Has("targetId") {
		w := `actor_service_id = UUID_TO_BIN(?)`
		selectEvents = selectEvents.Where(w, filter.MustString("targetId"))
		countEvents = countEvents.Where(w, filter.MustString("targetId"))
	}

	if filter.Has("targetEntityId") {
		w := `target_entity_id = UUID_TO_BIN(?)`
		selectEvents = selectEvents.Where(w, filter.MustString("targetEntityId"))
		countEvents = countEvents.Where(w, filter.MustString("targetEntityId"))
	}

	if filter.Has("targetServiceId") {
		w := `target_service_id = UUID_TO_BIN(?)`
		selectEvents = selectEvents.Where(w, filter.MustString("targetServiceId"))
		countEvents = countEvents.Where(w, filter.MustString("targetServiceId"))
	}

	if filter.Has("propertyId") {
		w := `tp.id = UUID_TO_BIN(?)`
		selectEvents = selectEvents.Where(w, filter.MustString("propertyId"))
		countEvents = countEvents.Where(w, filter.MustString("propertyId"))
	}

	if filter.Has("emittedAfter") {
		w := `emitted_at > ?`
		selectEvents = selectEvents.Where(w, filter.MustString("emittedAfter"))
		countEvents = countEvents.Where(w, filter.MustString("emittedAfter"))
	}

	if filter.Has("emittedBefore") {
		w := `emitted_at < ?`
		selectEvents = selectEvents.Where(w, filter.MustString("emittedBefore"))
		countEvents = countEvents.Where(w, filter.MustString("emittedBefore"))
	}

	if filter.Has("eventName") {
		w := `event_name = ?`
		selectEvents = selectEvents.Where(w, filter.MustString("eventName"))
		countEvents = countEvents.Where(w, filter.MustString("eventName"))
	}

	if sort.Empty() {
		selectEvents = selectEvents.OrderBy("emitted_at DESC")
	}

	if pagination.Page > 0 && pagination.PerPage > 0 {
		selectEvents = selectEvents.Limit(uint64(pagination.PerPage)).Offset(uint64(pagination.Offset()))
	}

	selectSQL, selectArgs, err := selectEvents.ToSql()
	if err != nil {
		return nil, err
	}

	countSQL, countArgs, err := countEvents.ToSql()
	if err != nil {
		return nil, err
	}

	return &selectQuery{
		selectSQL:  selectSQL,
		selectArgs: selectArgs,
		countSQL:   countSQL,
		countArgs:  countArgs,
	}, nil
}

func selectOneEventQuery(ID string) (string, []interface{}, error) {
	if !validator.IsUUID4(ID) {
		return "", nil, db.ErrInvalidUUID4
	}

	return baseSelectEventsQuery().
		Where("e.id = UUID_TO_BIN(?)", ID).
		Limit(1).
		ToSql()
}

func baseSelectEventsQuery() sq.SelectBuilder {
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
	query = query.Join("properties as tp ON te.id = tp.entity_id")
	return query
}

func createBaseCountEventsQuery() sq.SelectBuilder {
	query := sq.Select("COUNT(*) as total")

	query = query.From("events as e")
	query = query.Join("microservices as ams ON ams.id = e.actor_service_id")
	query = query.Join("microservices as tms ON tms.id = e.target_service_id")
	query = query.Join("entities as ae ON ae.id = e.actor_entity_id")
	query = query.Join("entities as te ON te.id = e.target_entity_id")
	query = query.Join("properties as tp ON te.id = tp.entity_id")

	return query
}

func countEventsQuery() (string, []interface{}, error) {
	return sq.Select("COUNT(*)").From("events").ToSql()
}

func createEventQuery(e *insertEvent) (string, []interface{}, error) {
	var pex sq.Sqlizer
	if e.ParentEventID.Valid {
		pex = sq.Expr("UUID_TO_BIN(?)", e.ParentEventID.String)
	} else {
		pex = sq.Expr("?", nil)
	}

	return sq.Insert("events").Columns(
		"id", "hash", "actor_id",
		"actor_entity_id", "actor_service_id", "target_id",
		"target_entity_id", "target_service_id", "event_name",
		"emitted_at", "registered_at", "parent_event_id",
	).Values(
		sq.Expr("UUID_TO_BIN(?)", e.ID),
		e.Hash,
		e.ActorID,
		sq.Expr("UUID_TO_BIN(?)", e.ActorEntityID),
		sq.Expr("UUID_TO_BIN(?)", e.ActorServiceID),
		e.TargetID,
		sq.Expr("UUID_TO_BIN(?)", e.TargetEntityID),
		sq.Expr("UUID_TO_BIN(?)", e.TargetServiceID),
		e.EventName, e.EmittedAt, e.RegisteredAt,
		pex,
	).ToSql()
}
