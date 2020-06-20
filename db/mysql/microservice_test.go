package mysql

import (
	"github.com/denismitr/auditbase/db"
	"github.com/denismitr/auditbase/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFirstMicroserviceByIDQuery(t *testing.T) {
	tt := []struct {
		name    string
		SQL     string
		args    []interface{}
		ID      string
		err     error
	}{
		{
			name: "normal-id",
			SQL:  "SELECT BIN_TO_UUID(id) as id, name, description, created_at, updated_at FROM microservices "+
				  "WHERE id = UUID_TO_BIN(?)",
			args: []interface{}{"bbdd1efe-2430-4c9c-a7fc-04d1a8e82e11"},
			ID:   "bbdd1efe-2430-4c9c-a7fc-04d1a8e82e11",
			err:  nil,
		},
		{
			name: "empty-id",
			SQL:  "",
			args: nil,
			ID:   "",
			err:  db.ErrEmptyUUID4,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			q, args, err := firstMicroserviceByIDQuery(tc.ID)

			if err == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.err, err)
			}

			assert.Equal(t, tc.args, args)
			assert.Equal(t, tc.SQL, q)
		})
	}
}

func TestCreateMicroserviceQuery(t *testing.T) {
	tt := []struct {
		name    string
		SQL     string
		args    []interface{}
		m      *model.Microservice
		err     error
	}{
		{
			name: "normal-id",
			SQL:  "INSERT INTO microservices (id,name,description) VALUES (UUID_TO_BIN(?),?,?)",
			args: []interface{}{"bbdd1efe-2430-4c9c-a7fc-04d1a8e82e11", "foo", "bar"},
			m:   &model.Microservice{
				ID: "bbdd1efe-2430-4c9c-a7fc-04d1a8e82e11",
				Name: "foo",
				Description: "bar",
			},
			err:  nil,
		},
		{
			name: "empty-id",
			SQL:  "",
			args: nil,
			m:   &model.Microservice{
				ID: "",
				Name: "foo",
				Description: "bar",
			},
			err:  db.ErrEmptyUUID4,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			q, args, err := createMicroserviceQuery(tc.m)

			if err == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.err, err)
			}

			assert.Equal(t, tc.args, args)
			assert.Equal(t, tc.SQL, q)
		})
	}
}
