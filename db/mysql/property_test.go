package mysql

import (
	"github.com/denismitr/auditbase/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateSelectPropertiesQuery(t *testing.T) {
	tt := []struct{
		name string
		selectSQL string
		selectArgs []interface{}
		countSQL string
		countArgs []interface{}
		propertyName string
		sortBy string
		entityID string
		page int
		perPage int
		err error
	}{
		{
			name: "default",
			selectSQL: `SELECT BIN_TO_UUID(p.id) as id, BIN_TO_UUID(p.entity_id) as entity_id, p.name FROM properties as p ORDER BY id ASC LIMIT 30 OFFSET 0`,
			countSQL: `SELECT COUNT(*) as total FROM properties as p`,
			countArgs: nil,
			selectArgs: nil,
			page: 1,
			perPage: 30,
			err: nil,
		},
		{
			name: "default",
			selectSQL: `SELECT BIN_TO_UUID(p.id) as id, BIN_TO_UUID(p.entity_id) as entity_id, p.name FROM properties as p WHERE p.name = name ORDER BY id ASC LIMIT 30 OFFSET 0`,
			countSQL: `SELECT COUNT(*) as total FROM properties as p WHERE p.name = name`,
			countArgs: []interface{}{"foo"},
			selectArgs: []interface{}{"foo"},
			propertyName: "foo",
			entityID: "eadd1efe-2430-4c9c-a7fc-04d1a8e82e96",
			page: 1,
			perPage: 30,
			err: nil,
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
