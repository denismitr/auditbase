package mysql

import (
	"github.com/denismitr/auditbase/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_firstEntityTypeByNameAndServiceIDQuery(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		name := "foo"
		serviceID := model.ID("015dc021-e6c0-4e25-a1a1-50e6c637524c")
		expected := "SELECT bin_to_uuid(`et`.`id`) AS `id`, bin_to_uuid(`et`.`service_id`) AS `service_id`, " +
			"`et`.`name` AS `name`, `et`.`description` AS `description`, `et`.`created_at` AS `created_at`, " +
			"`et`.`updated_at` AS `updated_at`, count(distinct `e`.`id`) AS `entities_cnt` " +
			"FROM `entity_types` AS `et` " +
			"LEFT JOIN `entities` AS `e` ON (`e`.`entity_type_id` = `et`.`id`) " +
			"WHERE (`et`.`service_id` = uuid_to_bin(?) AND (`name` = ?)) " +
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
		ID := "015dc021-e6c0-4e25-a1a1-50e6c637524c"
		expected := "SELECT bin_to_uuid(`et`.`id`) AS `id`, bin_to_uuid(`et`.`service_id`) AS `service_id`, " +
			"`et`.`name` AS `name`, `et`.`description` AS `description`, `et`.`created_at` AS `created_at`, " +
			"`et`.`updated_at` AS `updated_at`, count(distinct `e`.`id`) AS `entities_cnt` " +
			"FROM `entity_types` AS `et` " +
			"LEFT JOIN `entities` AS `e` ON (`e`.`entity_type_id` = `et`.`id`) " +
			"WHERE `et`.`id` = uuid_to_bin(?) " +
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
		ID          string
		serviceID   string
		name        string
		description string
		isActor     bool
		expected    string
	}{
		{
			ID:          "eadd1efe-2430-4c9c-a7fc-04d1a8e82e96",
			serviceID:   "bb1d1efe-2430-4c9c-a7fc-04d1a8e82e22",
			name:        "foo",
			description: "bar",
			isActor:     true,
			expected:    "INSERT INTO `entity_types` (`id`, `service_id`, `name`, `description`, `is_actor`) VALUES (uuid_to_bin(?), uuid_to_bin(?), ?, ?, ?)",
		},
		{
			ID:          "eadd1efe-2430-4c9c-a7fc-04d1a8e82e96",
			serviceID:   "bb1d1efe-2430-4c9c-a7fc-04d1a8e82e22",
			name:        "foo",
			description: "",
			isActor:     false,
			expected:    "INSERT INTO `entity_types` (`id`, `service_id`, `name`, `description`, `is_actor`) VALUES (uuid_to_bin(?), uuid_to_bin(?), ?, ?, ?)",
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
