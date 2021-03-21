package mysql

import (
	"github.com/denismitr/auditbase/db"
	"github.com/denismitr/auditbase/model"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_deleteMicroserviceQuery(t *testing.T) {
	t.Run("valid id", func(t *testing.T) {
		id := 12
		q, args, err := deleteMicroserviceQuery(model.ID(id))
		assert.NoError(t, err)
		assert.Equal(t, "DELETE FROM `microservices` WHERE id = uuid_to_bin(?)", q)
		assert.Len(t, args, 1)
		assert.Equal(t, args[0], "bbdd1efe-2430-4c9c-a7fc-04d1a8e82e11")
	})

	t.Run("invalid ID", func(t *testing.T) {
		id := -12
		_, _, err := deleteMicroserviceQuery(model.ID(id))
		assert.Error(t, err)
	})
}

func Test_updateMicroserviceQuery(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		id := 10
		now := time.Now()
		m := &model.Microservice{
			Name: "foo-service",
			Description: "Foo service",
			CreatedAt: model.JSONTime{Time: now},
			UpdatedAt: model.JSONTime{Time: now.Add(1 * time.Hour)},
		}

		q, args, err := updateMicroserviceQuery(model.ID(id), m)
		assert.NoError(t, err)
		assert.Equal(t, "UPDATE `microservices` SET `description`=?,`name`=?,`updated_at`=? WHERE id = uuid_to_bin(?)", q)
		assert.Len(t, args, 4)
		assert.Equal(t, args[0], "Foo service")
		assert.Equal(t, args[1], "foo-service")
		assert.Equal(t, args[2], now.Add(1 * time.Hour).Unix())
		assert.Equal(t, args[3], 10)
	})

	t.Run("invalid ID", func(t *testing.T) {
		id := -111
		_, _, err := updateMicroserviceQuery(model.ID(id), new(model.Microservice))
		assert.Error(t, err)
	})

	t.Run("no name", func(t *testing.T) {
		id := 10
		_, _, err := updateMicroserviceQuery(model.ID(id), new(model.Microservice))
		assert.Error(t, err)
		assert.Equal(t, "how can microservice name be empty on update?", err.Error())
	})

	t.Run("no updated at", func(t *testing.T) {
		id := 12
		m := &model.Microservice{Name: "foo"}
		_, _, err := updateMicroserviceQuery(model.ID(id), m)
		assert.Error(t, err)
		assert.Equal(t, "how can microservice updated at time be zero?", err.Error())
	})
}

func TestFirstMicroserviceByIDQuery(t *testing.T) {
	tt := []struct {
		name    string
		SQL     string
		args    []interface{}
		ID      model.ID
		err     error
	}{
		{
			name: "normal-id",
			SQL:  "SELECT BIN_TO_UUID(id) as id, name, description, created_at, updated_at FROM microservices "+
				  "WHERE id = UUID_TO_BIN(?)",
			args: []interface{}{"bbdd1efe-2430-4c9c-a7fc-04d1a8e82e11"},
			ID:   12,
			err:  nil,
		},
		{
			name: "empty-id",
			SQL:  "",
			args: nil,
			ID:   0,
			err:  db.ErrInvalidID,
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
				ID: 12,
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
				ID: 0,
				Name: "foo",
				Description: "bar",
			},
			err:  db.ErrInvalidID,
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
