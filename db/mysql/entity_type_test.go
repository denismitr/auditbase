package mysql

import (
	"github.com/denismitr/auditbase/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_firstEntityTypeByNameAndServiceIDQuery(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		name := "foo"
		serviceID := model.ID(123)
		expected := "SELECT `et`.`id` AS `id`, `et`.`service_id` AS `service_id`, " +
			"`et`.`name` AS `name`, `et`.`description` AS `description`, `et`.`created_at` AS `created_at`, " +
			"`et`.`updated_at` AS `updated_at`, count(distinct `e`.`id`) AS `entities_cnt` " +
			"FROM `entity_types` AS `et` " +
			"LEFT JOIN `entities` AS `e` ON (`e`.`entity_type_id` = `et`.`id`) " +
			"WHERE (`et`.`service_id` = ? AND (`name` = ?)) " +
			"GROUP BY `et`.`id`, `et`.`service_id`, `et`.`name`, `et`.`description`, `et`.`created_at`, `et`.`updated_at` " +
			"LIMIT ?"

		queryStr, args, err := firstEntityTypeByNameAndServiceIDQuery(name, serviceID)

		assert.NoError(t, err)
		assert.Len(t, args, 3)
		assert.Equal(t, expected, queryStr)
	})
}

func Test_firstEntityTypeIDQuery(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		ID := model.ID(11)
		expected := "SELECT `et`.`id` AS `id`, `et`.`service_id` AS `service_id`, " +
			"`et`.`name` AS `name`, `et`.`description` AS `description`, `et`.`created_at` AS `created_at`, " +
			"`et`.`updated_at` AS `updated_at`, count(distinct `e`.`id`) AS `entities_cnt` " +
			"FROM `entity_types` AS `et` " +
			"LEFT JOIN `entities` AS `e` ON (`e`.`entity_type_id` = `et`.`id`) " +
			"WHERE `et`.`id` = ? " +
			"GROUP BY `et`.`id`, `et`.`service_id`, `et`.`name`, `et`.`description`, `et`.`created_at`, `et`.`updated_at` " +
			"LIMIT ?"

		queryStr, args, err := firstEntityTypeByIDQuery(ID)

		assert.NoError(t, err)
		assert.Len(t, args, 2)
		assert.Equal(t, expected, queryStr)
	})
}

func Test_create(t *testing.T) {
	validInputs := []struct {
		ID          model.ID
		serviceID   model.ID
		name        string
		description string
		isActor     bool
		expected    string
	}{
		{
			serviceID:   model.ID(124),
			name:        "foo",
			description: "bar",
			isActor:     true,
			expected:    "INSERT INTO `entity_types` (`service_id`, `name`, `description`, `is_actor`) VALUES (?, ?, ?, ?)",
		},
		{
			serviceID:   model.ID(124),
			name:        "foo",
			description: "",
			isActor:     false,
			expected:    "INSERT INTO `entity_types` (`service_id`, `name`, `description`, `is_actor`) VALUES (?, ?, ?, ?)",
		},
	}

	for _, tc := range validInputs {
		q, args, err := createEntityTypeQuery(tc.ID, tc.serviceID, tc.name, tc.description, tc.isActor)
		assert.NoError(t, err)
		assert.Equal(t, tc.expected, q)

		assert.Len(t, args, 4)
		assert.Equal(t, int64(tc.serviceID), args[0])
		assert.Equal(t, tc.name, args[1])
		assert.Equal(t, tc.description, args[2])
		assert.Equal(t, tc.isActor, args[3])
	}
}
