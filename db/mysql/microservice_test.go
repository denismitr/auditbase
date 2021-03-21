package mysql

import (
	"github.com/denismitr/auditbase/model"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_deleteMicroserviceQuery(t *testing.T) {
	t.Run("valid id", func(t *testing.T) {
		var id int64 = 12
		q, args, err := deleteMicroserviceQuery(model.ID(id))
		assert.NoError(t, err)
		assert.Equal(t, "DELETE FROM `microservices` WHERE `id` = ?", q)
		assert.Len(t, args, 1)
		assert.Equal(t, id, args[0])
	})

	t.Run("invalid ID", func(t *testing.T) {
		id := 0
		_, _, err := deleteMicroserviceQuery(model.ID(id))
		assert.Error(t, err)
	})
}

func Test_updateMicroserviceQuery(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		var id int64 = 10
		now := time.Now()
		m := &model.Microservice{
			Name: "foo-service",
			Description: "Foo service",
			CreatedAt: model.JSONTime{Time: now},
			UpdatedAt: model.JSONTime{Time: now.Add(1 * time.Hour)},
		}

		q, args, err := updateMicroserviceQuery(model.ID(id), m)
		assert.NoError(t, err)
		assert.Equal(t, "UPDATE `microservices` SET `description`=?,`name`=?,`updated_at`=? WHERE `id`=?", q)
		assert.Len(t, args, 4)
		assert.Equal(t, "Foo service", args[0])
		assert.Equal(t, "foo-service", args[1])
		assert.Equal(t, now.Add(1 * time.Hour).Unix(), args[2])
		assert.Equal(t, id, args[3])
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
	t.Run("valid input", func(t *testing.T) {
		id := model.ID(12)
		expectedSQL := "SELECT `id`, `name`, `description`, `created_at`, `updated_at` FROM `microservices` "+
			"WHERE (`id` = ?)"

		q, args, err := firstMicroserviceByIDQuery(id)
		assert.NoError(t, err)
		assert.Len(t, args, 1)
		assert.Equal(t, int64(id), args[0])
		assert.Equal(t, expectedSQL, q)
	})

	t.Run("invalid input", func(t *testing.T) {
		var id model.ID = 0
		q, args, err := firstMicroserviceByIDQuery(id)
		assert.Error(t, err)
		assert.Len(t, args, 0)
		assert.Equal(t, "", q)
	})
}

func TestCreateMicroserviceQuery(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		m := model.Microservice{
			Name: "normal-name",
			Description: "foo",
		}

		expectedSQL := "INSERT INTO `microservices` (`description`, `name`) VALUES (?, ?)"
		q, args, err := createMicroserviceQuery(&m)
		assert.NoError(t, err)
		assert.Equal(t, expectedSQL, q)
		assert.Len(t, args, 2)
		assert.Equal(t, m.Description, args[0])
		assert.Equal(t, m.Name, args[1])
	})
}
