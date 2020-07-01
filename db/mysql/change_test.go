package mysql

import (
	"github.com/denismitr/auditbase/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateSelectChangesQuery(t *testing.T) {
	tt := []struct {
		name       string
		selectSQL  string
		selectArgs []interface{}
		countSQL   string
		countArgs  []interface{}
		propertyID string
		sortBy     string
		eventID    string
		page       int
		perPage    int
		err        error
	}{
		{
			name:       "default",
			selectSQL:  "SELECT BIN_TO_UUID(c.id) as id, BIN_TO_UUID(c.property_id) as property_id, " +
				"BIN_TO_UUID(c.event_id) as event_id, e.emitted_at as created_at, c.current_data_type, " +
				"c.from_value, c.to_value " +
				"FROM changes as c " +
				"JOIN events as e on e.id = c.event_id " +
				"ORDER BY e.emitted_at DESC LIMIT 30 OFFSET 0",
			countSQL:   "SELECT count(*) as total FROM changes as c",
			countArgs:  nil,
			selectArgs: nil,
			page:       1,
			perPage:    30,
			err:        nil,
		},
		{
			name: "with-property-id-and-event-id",
			selectSQL: "SELECT BIN_TO_UUID(c.id) as id, BIN_TO_UUID(c.property_id) as property_id, " +
				"BIN_TO_UUID(c.event_id) as event_id, e.emitted_at as created_at, c.current_data_type, " +
				"c.from_value, c.to_value " +
				"FROM changes as c " +
				"JOIN events as e on e.id = c.event_id " +
				"WHERE c.property_id = UUID_TO_BIN(?) AND c.event_id = UUID_TO_BIN(?) " +
				"ORDER BY e.emitted_at DESC LIMIT 30 OFFSET 0",
			countSQL: "SELECT count(*) as total FROM changes as c " +
				"WHERE c.property_id = UUID_TO_BIN(?) AND c.event_id = UUID_TO_BIN(?)",
			countArgs:  []interface{}{"bbdd1efe-2430-4c9c-a7fc-04d1a8e82e11", "bbdd1efe-2430-4c9c-a7fc-04d1a8e82e11"},
			selectArgs: []interface{}{"bbdd1efe-2430-4c9c-a7fc-04d1a8e82e11", "bbdd1efe-2430-4c9c-a7fc-04d1a8e82e11"},
			propertyID: "bbdd1efe-2430-4c9c-a7fc-04d1a8e82e11",
			eventID:    "eadd1efe-2430-4c9c-a7fc-04d1a8e82e96",
			page:       1,
			perPage:    30,
			err:        nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			f := model.NewFilter([]string{"eventId", "propertyId"})
			s := model.NewSort()
			pg := &model.Pagination{Page: tc.page, PerPage: tc.perPage}

			if tc.propertyID != "" {
				f.Add("propertyId", tc.propertyID)
			}

			if tc.eventID != "" {
				f.Add("eventId", tc.propertyID)
			}

			q, err := selectChangesQuery(f, s, pg)

			assert.NoError(t, err)
			assert.NotNil(t, q)
			assert.Equal(t, tc.selectSQL, q.selectSQL)
			assert.Equal(t, tc.countSQL, q.countSQL)

			assert.Equal(t, tc.selectArgs, q.selectArgs)
			assert.Equal(t, tc.countArgs, q.countArgs)
		})
	}
}

func TestCreateFirstChangeByIDQuery(t *testing.T) {
	tt := []struct {
		name    string
		SQL     string
		args    []interface{}
		ID      string
		eventID string
		err     error
	}{
		{
			name: "default",
			SQL: "SELECT BIN_TO_UUID(c.id) as id, BIN_TO_UUID(c.property_id) as property_id, " +
				"BIN_TO_UUID(c.event_id) as event_id, e.emitted_at as created_at, c.current_data_type, " +
				"c.from_value, c.to_value " +
				"FROM changes as c " +
				"JOIN events as e on e.id = c.event_id " +
				"WHERE c.id = UUID_TO_BIN(?) LIMIT 1",
			args: []interface{}{"bbdd1efe-2430-4c9c-a7fc-04d1a8e82e11"},
			ID:   "bbdd1efe-2430-4c9c-a7fc-04d1a8e82e11",
			err:  nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			q, args, err := firstChangeByIDQuery(tc.ID)

			assert.NoError(t, err)
			assert.Equal(t, tc.args, args)
			assert.Equal(t, tc.SQL, q)
		})
	}
}
