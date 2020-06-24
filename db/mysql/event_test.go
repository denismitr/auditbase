package mysql

import (
	"fmt"
	"github.com/denismitr/auditbase/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateSelectChangesByIDsQuery(t *testing.T) {
	tt := []struct{
		ids []string
		sql string
		args []interface{}
		err error
	}{
		{
			ids: []string{"34e1d82a-a065-436d-afd0-5fbcb752a4e1", "55e1d82a-a065-436d-bfd0-5fbcb752a4f2"},
			sql: "SELECT BIN_TO_UUID(c.id) as id, BIN_TO_UUID(c.property_id) as property_id, BIN_TO_UUID(c.event_id) as event_id, BIN_TO_UUID(p.entity_id) as entity_id, p.name as property_name, from_value, to_value FROM changes as c JOIN properties as p ON p.id = c.property_id WHERE event_id IN (UUID_TO_BIN(?),UUID_TO_BIN(?))",
			args: []interface{}{"34e1d82a-a065-436d-afd0-5fbcb752a4e1", "55e1d82a-a065-436d-bfd0-5fbcb752a4f2"},
			err: nil,
		},
		{
			ids: []string{},
			sql: "",
			args: nil,
			err: ErrEmptyWhereInList,
		},
	}

	for _, tc := range tt {
		t.Run(fmt.Sprintf("%#v", tc.ids), func(t *testing.T) {
			sql, args, err := createSelectChangesByIDsQuery(tc.ids)

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
		name       string
		selectSQL  string
		selectArgs []interface{}
		countSQL   string
		countArgs  []interface{}
		eventName  string
		emittedAfter string
		emittedBefore string
		sortBy     string
		page       int
		perPage    int
		err        error
	}{
		{
			name:       "default",
			selectSQL:  "SELECT BIN_TO_UUID(e.id) as id, BIN_TO_UUID(parent_event_id) as parent_event_id, BIN_TO_UUID(actor_entity_id) as actor_entity_id, " +
				        "BIN_TO_UUID(actor_service_id) as actor_service_id, BIN_TO_UUID(target_entity_id) as target_entity_id, " +
				        "BIN_TO_UUID(target_service_id) as target_service_id, hash, actor_id, target_id, event_name, emitted_at, registered_at, " +
				        "ams.name as actor_service_name, tms.name as target_service_name, ams.description as actor_service_description, " +
				        "tms.description as target_service_description, ae.name as actor_entity_name, ae.description as actor_entity_description, " +
				        "te.name as target_entity_name, te.description as target_entity_description FROM events as e " +
				        "JOIN microservices as ams ON ams.id = e.actor_service_id JOIN microservices as tms ON tms.id = e.target_service_id " +
				        "JOIN entities as ae ON ae.id = e.actor_entity_id JOIN entities as te ON te.id = e.target_entity_id ORDER BY emitted_at DESC LIMIT 30 OFFSET 0",
			countSQL:   `SELECT COUNT(*) as total FROM events e`,
			countArgs:  nil,
			selectArgs: nil,
			page:       1,
			perPage:    30,
			err:        nil,
		},
		{
			name:       "with-name-and-emitted-at-limit",
			selectSQL:  "SELECT BIN_TO_UUID(e.id) as id, BIN_TO_UUID(parent_event_id) as parent_event_id, BIN_TO_UUID(actor_entity_id) as actor_entity_id, " +
				        "BIN_TO_UUID(actor_service_id) as actor_service_id, BIN_TO_UUID(target_entity_id) as target_entity_id, " +
				        "BIN_TO_UUID(target_service_id) as target_service_id, hash, actor_id, target_id, event_name, emitted_at, registered_at, " +
				        "ams.name as actor_service_name, tms.name as target_service_name, ams.description as actor_service_description, " +
				        "tms.description as target_service_description, ae.name as actor_entity_name, ae.description as actor_entity_description, " +
				        "te.name as target_entity_name, te.description as target_entity_description FROM events as e " +
				        "JOIN microservices as ams ON ams.id = e.actor_service_id JOIN microservices as tms ON tms.id = e.target_service_id " +
				        "JOIN entities as ae ON ae.id = e.actor_entity_id JOIN entities as te ON te.id = e.target_entity_id " +
				        "WHERE emitted_at > ? AND emitted_at < ? AND event_name = ? ORDER BY emitted_at DESC LIMIT 30 OFFSET 0",
			countSQL:   `SELECT COUNT(*) as total FROM events e WHERE emitted_at > ? AND emitted_at < ? AND event_name = ?`,
			countArgs:  []interface{}{"12345678901", "12345688901", "subscriptionCanceled"},
			selectArgs: []interface{}{"12345678901", "12345688901", "subscriptionCanceled"},
			eventName: "subscriptionCanceled",
			emittedAfter:    "12345678901",
			emittedBefore:    "12345688901",
			page:       1,
			perPage:    30,
			err:        nil,
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
