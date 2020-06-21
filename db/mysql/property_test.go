package mysql

import (
	"github.com/denismitr/auditbase/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateSelectPropertiesQuery(t *testing.T) {
	tt := []struct {
		name         string
		selectSQL    string
		selectArgs   []interface{}
		countSQL     string
		countArgs    []interface{}
		propertyName string
		sortBy       string
		entityID     string
		page         int
		perPage      int
		err          error
	}{
		{
			name:       "default",
			selectSQL:  "SELECT BIN_TO_UUID(p.id) as id, BIN_TO_UUID(p.entity_id) as entity_id, p.name, max(e.emitted_at) as last_event_at, count(c.id) as change_count " +
						"FROM properties as p JOIN changes c ON c.property_id = p.id JOIN events e ON c.event_id = e.id GROUP BY p.id, p.entity_id, p.name " +
						"ORDER BY max(e.emitted_at) DESC LIMIT 30 OFFSET 0",
			countSQL:   `SELECT COUNT(*) as total FROM properties as p`,
			countArgs:  nil,
			selectArgs: nil,
			page:       1,
			perPage:    30,
			err:        nil,
		},
		{
			name:         "default",
			selectSQL:    "SELECT BIN_TO_UUID(p.id) as id, BIN_TO_UUID(p.entity_id) as entity_id, p.name, max(e.emitted_at) as last_event_at, count(c.id) as change_count " +
						  "FROM properties as p JOIN changes c ON c.property_id = p.id JOIN events e ON c.event_id = e.id WHERE p.name = name " +
						  "GROUP BY p.id, p.entity_id, p.name ORDER BY max(e.emitted_at) DESC LIMIT 30 OFFSET 0",
			countSQL:     `SELECT COUNT(*) as total FROM properties as p WHERE p.name = name`,
			countArgs:    []interface{}{"foo"},
			selectArgs:   []interface{}{"foo"},
			propertyName: "foo",
			entityID:     "eadd1efe-2430-4c9c-a7fc-04d1a8e82e96",
			page:         1,
			perPage:      30,
			err:          nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			f := model.NewFilter([]string{"entityId", "name", "includeChanges", "includeChangesCount"})
			s := model.NewSort()
			pg := &model.Pagination{Page: tc.page, PerPage: tc.perPage}

			if tc.propertyName != "" {
				f.Add("name", tc.propertyName)
			}

			q, err := createSelectPropertiesQuery(f, s, pg)

			assert.NoError(t, err)
			assert.NotNil(t, q)
			assert.Equal(t, tc.selectSQL, q.selectSQL)
			assert.Equal(t, tc.countSQL, q.countSQL)

			assert.Equal(t, tc.selectArgs, q.selectArgs)
			assert.Equal(t, tc.countArgs, q.countArgs)
		})
	}
}

func TestCreateFirstByIDQuery(t *testing.T) {
	tt := []struct {
		name string
		sql  string
		args []interface{}
		ID   string
		err  error
	}{
		{
			name: "default",
			ID:   "eadd1efe-2430-4c9c-a7fc-04d1a8e82e96",
			sql:  `SELECT BIN_TO_UUID(p.id) as id, BIN_TO_UUID(p.entity_id) as entity_id, p.name FROM properties as p WHERE p.id = UUID_TO_BIN(?)`,
			args: []interface{}{"eadd1efe-2430-4c9c-a7fc-04d1a8e82e96"},
			err:  nil,
		},
		{
			name: "invalid-id",
			ID:   "foo",
			sql:  ``,
			args: nil,
			err:  errors.Errorf("%s is not a valid uuid4", "foo"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			sql, args, err := createFirstByIDQuery(tc.ID)

			if tc.err == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.err.Error(), err.Error())
			}

			assert.Equal(t, tc.sql, sql)
			assert.Equal(t, tc.args, args)
		})
	}
}

func TestCreateGetPropertyIDQuery(t *testing.T) {
	tt := []struct {
		name         string
		sql          string
		args         []interface{}
		entityID     string
		propertyName string
		err          error
	}{
		{
			name:         "default",
			entityID:     "eadd1efe-2430-4c9c-a7fc-04d1a8e82e96",
			sql:          `SELECT BIN_TO_UUID(id) as id FROM properties WHERE name = ? AND entity_id = UUID_TO_BIN(?) LIMIT 1`,
			args:         []interface{}{"foo", "eadd1efe-2430-4c9c-a7fc-04d1a8e82e96"},
			propertyName: "foo",
			err:          nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			sql, args, err := createGetPropertyIDQuery(tc.propertyName, tc.entityID)

			if tc.err == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.err.Error(), err.Error())
			}

			assert.Equal(t, tc.args, args)
			assert.Equal(t, tc.sql, sql)
		})
	}
}

func TestCreateInsertPropertyQuery(t *testing.T) {
	tt := []struct {
		name         string
		sql          string
		args         []interface{}
		entityID     string
		ID           string
		propertyName string
		err          error
	}{
		{
			name:         "default",
			ID:           "ccdd1efe-2430-4c9c-a7fc-04d1a8e82f11",
			entityID:     "eadd1efe-2430-4c9c-a7fc-04d1a8e82e96",
			sql:          `INSERT INTO properties (id,name,entity_id) VALUES (UUID_TO_BIN(?),?,UUID_TO_BIN(?))`,
			args:         []interface{}{"ccdd1efe-2430-4c9c-a7fc-04d1a8e82f11", "foo", "eadd1efe-2430-4c9c-a7fc-04d1a8e82e96"},
			propertyName: "foo",
			err:          nil,
		},
		{
			name:         "invalid-id",
			ID:           "bar",
			entityID:     "eadd1efe-2430-4c9c-a7fc-04d1a8e82e96",
			sql:          ``,
			args:         nil,
			propertyName: "foo",
			err:          errors.New("bar is not a valid uuid4"),
		},
		{
			name:         "invalid-entities-id",
			ID:           "eadd1efe-2430-4c9c-a7fc-04d1a8e82e96",
			entityID:     "baz",
			sql:          ``,
			args:         nil,
			propertyName: "foo",
			err:          errors.New("baz is not a valid uuid4"),
		},
		{
			name:         "no-properties-name",
			ID:           "ccdd1efe-2430-4c9c-a7fc-04d1a8e82f11",
			entityID:     "eadd1efe-2430-4c9c-a7fc-04d1a8e82e96",
			sql:          ``,
			args:         nil,
			propertyName: "",
			err:          errors.New("properties name is empty"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			sql, args, err := createInsertPropertyQuery(tc.ID, tc.propertyName, tc.entityID)

			if tc.err == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.err.Error(), err.Error())
			}

			assert.Equal(t, tc.args, args)
			assert.Equal(t, tc.sql, sql)
		})
	}
}
