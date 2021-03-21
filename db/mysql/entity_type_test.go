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
			ID:          model.ID(123),
			serviceID:   model.ID(124),
			name:        "foo",
			description: "bar",
			isActor:     true,
			expected:    "INSERT INTO `entity_types` (`id`, `service_id`, `name`, `description`, `is_actor`) VALUES (?, ?, ?, ?, ?)",
		},
		{
			ID:          model.ID(123),
			serviceID:   model.ID(124),
			name:        "foo",
			description: "",
			isActor:     false,
			expected:    "INSERT INTO `entity_types` (`id`, `service_id`, `name`, `description`, `is_actor`) VALUES (?, ?, ?, ?, ?)",
		},
	}

	for _, tc := range validInputs {
		q, args, err := createEntityTypeQuery(tc.ID, tc.serviceID, tc.name, tc.description, tc.isActor)
		assert.NoError(t, err)
		assert.Len(t, args, 5)
		assert.Equal(t, args[0], tc.ID)
		assert.Equal(t, args[1], tc.serviceID)
		assert.Equal(t, args[2], tc.name)
		assert.Equal(t, args[3], tc.description)
		assert.Equal(t, args[4], tc.isActor)
		assert.Equal(t, tc.expected, q)
	}
}
