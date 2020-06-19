package mysql

import (
	"github.com/denismitr/auditbase/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateSelectEntitiesQuery(t *testing.T) {
	nameAsc := model.ASCOrder
	_ = model.DESCOrder

	tt := []struct{
		name string
		serviceId string
		nameOrder *model.Order
		perPage int
		page int
		sql string
		args []interface{}
		err error
	}{
		{
			name: "default",
			serviceId: "",
			nameOrder: nil,
			perPage: 10,
			page: 0,
			sql: "SELECT BIN_TO_UUID(id) as id, BIN_TO_UUID(service_id) as service_id, name, description, created_at, updated_at FROM entities ORDER BY created_at DESC LIMIT 10 OFFSET 0",
			args: nil,
		},
		{
			name: "serviceId",
			serviceId: "18420701-7852-4361-8e91-a660385dd0c5",
			nameOrder: nil,
			perPage: 10,
			page: 0,
			sql: "SELECT BIN_TO_UUID(id) as id, BIN_TO_UUID(service_id) as service_id, name, description, created_at, updated_at FROM entities WHERE service_id = uuid_to_bin(?) ORDER BY service_id DESC LIMIT 10 OFFSET 0",
			args:[]interface{}{"18420701-7852-4361-8e91-a660385dd0c5"},
		},
		{
			name: "order-name",
			serviceId: "18420701-7852-4361-8e91-a660385dd0c5",
			nameOrder: &nameAsc,
			perPage: 10,
			page: 0,
			sql: "SELECT BIN_TO_UUID(id) as id, BIN_TO_UUID(service_id) as service_id, name, description, created_at, updated_at FROM entities WHERE service_id = uuid_to_bin(?) ORDER BY name ? LIMIT 10 OFFSET 0",
			args:[]interface{}{"18420701-7852-4361-8e91-a660385dd0c5", "ASC"},
		},
		{
			name: "order-name-pagination",
			serviceId: "",
			nameOrder: &nameAsc,
			perPage: 25,
			page: 3,
			sql: "SELECT BIN_TO_UUID(id) as id, BIN_TO_UUID(service_id) as service_id, name, description, created_at, updated_at FROM entities ORDER BY name ? LIMIT 25 OFFSET 50",
			args:[]interface{}{"ASC"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			f := model.NewFilter([]string{"serviceId"})
			if tc.serviceId != "" {
				f.Add("serviceId", tc.serviceId)
			}

			s := model.NewSort()
			if tc.nameOrder != nil {
				s.Add("name", *tc.nameOrder)
			}

			p := &model.Pagination{Page: tc.page, PerPage: tc.perPage}

			sql, args, err := createSelectEntitiesQuery(f, s, p)

			if tc.err == nil {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, tc.err, err)
			}

			assert.Equal(t, tc.sql, sql)
			assert.Equal(t, tc.args, args)
		})
	}
}

func TestCreateFirstEntityByIDQuery(t *testing.T) {
	tt := []struct{
		name string
		sql string
		args []interface{}
		ID string
		err error
	}{
		{
			name: "default",
			sql: `SELECT BIN_TO_UUID(id) as id, BIN_TO_UUID(service_id) as service_id, name, description, created_at, updated_at FROM entities WHERE id = UUID_TO_BIN(?) LIMIT 1`,
			ID: "eadd1efe-2430-4c9c-a7fc-04d1a8e82e96",
			args: []interface{}{"eadd1efe-2430-4c9c-a7fc-04d1a8e82e96"},
			err: nil,
		},
		{
			name: "invalid-uuid-4",
			sql: "",
			ID: "foo-123",
			args: nil,
			err: errors.New("foo-123 is not a valid UUID4"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			sql, args, err := createFirstEntityByIDQuery(tc.ID)

			if tc.err == nil {
				assert.Nil(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.err.Error(), err.Error())
			}

			assert.Equal(t, tc.sql, sql)
			assert.Equal(t, tc.args, args)
		})
	}
}
