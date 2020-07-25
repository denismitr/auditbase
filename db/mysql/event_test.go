package mysql

import (
	"fmt"
	"github.com/denismitr/auditbase/db"
	"github.com/denismitr/auditbase/model"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSelectOneEventQuery(t *testing.T) {
	tt := []struct {
		name string
		ID   string
		sql  string
		args []interface{}
		err  error
	}{
		{
			name: "valid ID",
			ID:   "34e1d82a-a065-436d-afd0-5fbcb752a4e1",
			sql: "SELECT BIN_TO_UUID(e.id) as id, BIN_TO_UUID(parent_event_id) as parent_event_id, BIN_TO_UUID(actor_entity_id) as actor_entity_id, " +
				"BIN_TO_UUID(actor_service_id) as actor_service_id, BIN_TO_UUID(target_entity_id) as target_entity_id, " +
				"BIN_TO_UUID(target_service_id) as target_service_id, hash, actor_id, target_id, event_name, emitted_at, " + "" +
				"registered_at, ams.name as actor_service_name, tms.name as target_service_name, ams.description as actor_service_description, " +
				"tms.description as target_service_description, ae.name as actor_entity_name, ae.description as actor_entity_description, " +
				"te.name as target_entity_name, te.description as target_entity_description FROM events as e JOIN microservices as ams ON ams.id = e.actor_service_id " +
				"JOIN microservices as tms ON tms.id = e.target_service_id JOIN entities as ae ON ae.id = e.actor_entity_id " +
				"JOIN entities as te ON te.id = e.target_entity_id " +
				"JOIN properties as tp ON te.id = tp.entity_id " +
				"WHERE e.id = UUID_TO_BIN(?) LIMIT 1",
			args: []interface{}{"34e1d82a-a065-436d-afd0-5fbcb752a4e1"},
			err:  nil,
		},
		{
			name: "invalid string",
			ID:   "foo",
			sql:  "",
			args: nil,
			err:  db.ErrInvalidUUID4,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			sql, args, err := selectOneEventQuery(tc.ID)

			if tc.err == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, tc.err, err)
			}

			assert.Equal(t, tc.sql, sql)
			assert.Equal(t, tc.args, args)
		})
	}
}

func TestSelectChangesByEventIDsQuery(t *testing.T) {
	tt := []struct {
		ids  []string
		sql  string
		args []interface{}
		err  error
	}{
		{
			ids:  []string{"34e1d82a-a065-436d-afd0-5fbcb752a4e1", "55e1d82a-a065-436d-bfd0-5fbcb752a4f2"},
			sql:  "SELECT BIN_TO_UUID(c.id) as id, BIN_TO_UUID(c.property_id) as property_id, BIN_TO_UUID(c.event_id) as event_id, BIN_TO_UUID(p.entity_id) as entity_id, p.name as property_name, c.current_data_type, c.from_value, c.to_value FROM changes as c JOIN properties as p ON p.id = c.property_id WHERE event_id IN (UUID_TO_BIN(?),UUID_TO_BIN(?))",
			args: []interface{}{"34e1d82a-a065-436d-afd0-5fbcb752a4e1", "55e1d82a-a065-436d-bfd0-5fbcb752a4f2"},
			err:  nil,
		},
		{
			ids:  []string{},
			sql:  "",
			args: nil,
			err:  db.ErrEmptyWhereInList,
		},
	}

	for _, tc := range tt {
		t.Run(fmt.Sprintf("%#v", tc.ids), func(t *testing.T) {
			sql, args, err := selectChangesByEventIDsQuery(tc.ids)

			if tc.err == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, tc.err, err)
			}

			assert.Equal(t, tc.sql, sql)
			assert.Equal(t, tc.args, args)
		})
	}
}

func TestSelectEventChanges(t *testing.T) {
	tt := []struct {
		name string
		id  string
		sql  string
		args []interface{}
		err  error
	}{
		{
			name: "default",
			id:  "34e1d82a-a065-436d-afd0-5fbcb752a4e1",
			sql:  "SELECT BIN_TO_UUID(c.id) as id, " +
				"BIN_TO_UUID(c.event_id) as event_id, " +
				"BIN_TO_UUID(c.property_id) as property_id, " +
				"BIN_TO_UUID(p.entity_id) as entity_id, " +
				"c.current_data_type, p.name as property_name, from_value, to_value " +
				"FROM changes as c JOIN properties as p ON p.id = c.property_id " +
				"WHERE event_id = UUID_TO_BIN(?)",
			args: []interface{}{"34e1d82a-a065-436d-afd0-5fbcb752a4e1"},
			err:  nil,
		},
		{
			id:  "no id",
			sql:  "",
			args: nil,
			err:  db.ErrInvalidUUID4,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			sql, args, err := selectChangesByEventIDQuery(tc.id)

			if tc.err == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, tc.err, err)
			}

			assert.Equal(t, tc.sql, sql)
			assert.Equal(t, tc.args, args)
		})
	}
}

func TestSelectEventsQuery(t *testing.T) {
	tt := []struct {
		name          string
		selectSQL     string
		selectArgs    []interface{}
		countSQL      string
		countArgs     []interface{}
		eventName     string
		emittedAfter  string
		emittedBefore string
		sortBy        string
		page          int
		perPage       int
		err           error
	}{
		{
			name: "default",
			selectSQL: "SELECT BIN_TO_UUID(e.id) as id, BIN_TO_UUID(parent_event_id) as parent_event_id, BIN_TO_UUID(actor_entity_id) as actor_entity_id, " +
				"BIN_TO_UUID(actor_service_id) as actor_service_id, BIN_TO_UUID(target_entity_id) as target_entity_id, " +
				"BIN_TO_UUID(target_service_id) as target_service_id, hash, actor_id, target_id, event_name, emitted_at, registered_at, " +
				"ams.name as actor_service_name, tms.name as target_service_name, ams.description as actor_service_description, " +
				"tms.description as target_service_description, ae.name as actor_entity_name, ae.description as actor_entity_description, " +
				"te.name as target_entity_name, te.description as target_entity_description FROM events as e " +
				"JOIN microservices as ams ON ams.id = e.actor_service_id JOIN microservices as tms ON tms.id = e.target_service_id " +
				"JOIN entities as ae ON ae.id = e.actor_entity_id JOIN entities as te ON te.id = e.target_entity_id " +
				"JOIN properties as tp ON te.id = tp.entity_id " +
				"ORDER BY emitted_at DESC LIMIT 30 OFFSET 0",
			countSQL:   `SELECT COUNT(*) as total FROM events as e JOIN microservices as ams ON ams.id = e.actor_service_id JOIN microservices as tms ON tms.id = e.target_service_id JOIN entities as ae ON ae.id = e.actor_entity_id JOIN entities as te ON te.id = e.target_entity_id JOIN properties as tp ON te.id = tp.entity_id`,
			countArgs:  nil,
			selectArgs: nil,
			page:       1,
			perPage:    30,
			err:        nil,
		},
		{
			name: "with-name-and-emitted-at-limit",
			selectSQL: "SELECT BIN_TO_UUID(e.id) as id, BIN_TO_UUID(parent_event_id) as parent_event_id, BIN_TO_UUID(actor_entity_id) as actor_entity_id, " +
				"BIN_TO_UUID(actor_service_id) as actor_service_id, BIN_TO_UUID(target_entity_id) as target_entity_id, " +
				"BIN_TO_UUID(target_service_id) as target_service_id, hash, actor_id, target_id, event_name, emitted_at, registered_at, " +
				"ams.name as actor_service_name, tms.name as target_service_name, ams.description as actor_service_description, " +
				"tms.description as target_service_description, ae.name as actor_entity_name, ae.description as actor_entity_description, " +
				"te.name as target_entity_name, te.description as target_entity_description FROM events as e " +
				"JOIN microservices as ams ON ams.id = e.actor_service_id JOIN microservices as tms ON tms.id = e.target_service_id " +
				"JOIN entities as ae ON ae.id = e.actor_entity_id JOIN entities as te ON te.id = e.target_entity_id " +
				"JOIN properties as tp ON te.id = tp.entity_id " +
				"WHERE emitted_at > ? AND emitted_at < ? AND event_name = ? ORDER BY emitted_at DESC LIMIT 30 OFFSET 0",
			countSQL:      `SELECT COUNT(*) as total FROM events as e JOIN microservices as ams ON ams.id = e.actor_service_id JOIN microservices as tms ON tms.id = e.target_service_id JOIN entities as ae ON ae.id = e.actor_entity_id JOIN entities as te ON te.id = e.target_entity_id JOIN properties as tp ON te.id = tp.entity_id WHERE emitted_at > ? AND emitted_at < ? AND event_name = ?`,
			countArgs:     []interface{}{"12345678901", "12345688901", "subscriptionCanceled"},
			selectArgs:    []interface{}{"12345678901", "12345688901", "subscriptionCanceled"},
			eventName:     "subscriptionCanceled",
			emittedAfter:  "12345678901",
			emittedBefore: "12345688901",
			page:          1,
			perPage:       30,
			err:           nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			f := model.NewFilter([]string{"eventName", "emittedAfter", "emittedBefore"})
			s := model.NewSort()
			pg := &model.Pagination{Page: tc.page, PerPage: tc.perPage}

			if tc.eventName != "" {
				f.Add("eventName", tc.eventName)
			}

			if tc.emittedAfter != "" {
				f.Add("emittedAfter", tc.emittedAfter)
			}

			if tc.emittedBefore != "" {
				f.Add("emittedBefore", tc.emittedBefore)
			}

			q, err := selectEventsQuery(f, s, pg)

			assert.NoError(t, err)
			assert.NotNil(t, q)
			assert.Equal(t, tc.selectSQL, q.selectSQL)
			assert.Equal(t, tc.countSQL, q.countSQL)

			assert.Equal(t, tc.selectArgs, q.selectArgs)
			assert.Equal(t, tc.countArgs, q.countArgs)
		})
	}
}

func TestCreateEventQuery(t *testing.T) {
	tt := []struct {
		name string
		e    insertEvent
		sql  string
		err  error
	}{
		{
			name: "default",
			e: insertEvent{
				ID:              "34e1d82a-a065-436d-afd0-5fbcb752a4e1",
				ParentEventID:   db.NullStringFromStringPointer(nil),
				Hash:            "dasdasgdfgvrr432534fsafdsfdsa",
				ActorID:         "123",
				TargetID:        "367",
				ActorServiceID:  "55e1d82a-a077-436d-afd0-5fbcb752a4ff",
				TargetServiceID: "66e1d82a-a087-436d-afd0-5fbcb752a333",
				ActorEntityID:   "22e1d82a-a087-436d-afd0-5fbcb752a222",
				TargetEntityID:  "22e1d82a-a087-436d-afd0-5fbcb752a222",
				EventName:       "fooBarEvent",
				EmittedAt:       time.Unix(1578174512, 0),
				RegisteredAt:    time.Unix(1578174513, 0),
			},
			sql: "INSERT INTO events " +
				"(id,hash,actor_id,actor_entity_id,actor_service_id,target_id,target_entity_id,target_service_id,event_name,emitted_at,registered_at,parent_event_id) " +
				"VALUES " +
				"(UUID_TO_BIN(?),?,?,UUID_TO_BIN(?),UUID_TO_BIN(?),?,UUID_TO_BIN(?),UUID_TO_BIN(?),?,?,?,?)",
			err: nil,
		},
		{
			name: "with-parent-id",
			e: insertEvent{
				ID:              "34e1d82a-a065-436d-afd0-5fbcb752a4e1",
				ParentEventID:   db.NullStringFromString("88e1d82a-b065-436d-afd0-5fbcb752a4f6"),
				Hash:            "dasdasgdfgvrr432534fsafdsfdsa",
				ActorID:         "123",
				TargetID:        "367",
				ActorServiceID:  "55e1d82a-a077-436d-afd0-5fbcb752a4ff",
				TargetServiceID: "66e1d82a-a087-436d-afd0-5fbcb752a333",
				ActorEntityID:   "22e1d82a-a087-436d-afd0-5fbcb752a222",
				TargetEntityID:  "22e1d82a-a087-436d-afd0-5fbcb752a222",
				EventName:       "fooBarEvent",
				EmittedAt:       time.Unix(1578174512, 0),
				RegisteredAt:    time.Unix(1578174513, 0),
			},
			sql: "INSERT INTO events " +
				"(id,hash,actor_id,actor_entity_id,actor_service_id,target_id,target_entity_id,target_service_id,event_name,emitted_at,registered_at,parent_event_id) " +
				"VALUES " +
				"(UUID_TO_BIN(?),?,?,UUID_TO_BIN(?),UUID_TO_BIN(?),?,UUID_TO_BIN(?),UUID_TO_BIN(?),?,?,?,UUID_TO_BIN(?))",
			err: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			q, args, err := createEventQuery(&tc.e)

			if tc.err == nil {
				assert.NoError(t, err)
				assert.Equal(t, tc.sql, q)
				assert.Len(t, args, 12)

				if tc.e.ParentEventID.Valid {
					assert.Equal(t, tc.e.ParentEventID.String, args[11])
				} else {
					assert.Equal(t, nil, args[11])
				}

				assert.Equal(t, tc.e.ID, args[0])
				assert.Equal(t, tc.e.Hash, args[1])
				assert.Equal(t, tc.e.ActorID, args[2])
				assert.Equal(t, tc.e.ActorEntityID, args[3])
				assert.Equal(t, tc.e.ActorServiceID, args[4])
				assert.Equal(t, tc.e.TargetID, args[5])
				assert.Equal(t, tc.e.TargetEntityID, args[6])
				assert.Equal(t, tc.e.TargetServiceID, args[7])
				assert.Equal(t, tc.e.EventName, args[8])
				assert.Equal(t, tc.e.EmittedAt, args[9])
				assert.Equal(t, tc.e.RegisteredAt, args[10])
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.err.Error(), err.Error())
			}
		})
	}
}
