package mysql

import (
	"github.com/denismitr/auditbase/db"
	"github.com/denismitr/auditbase/model"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestSelectEntitiesQuery(t *testing.T) {
	tt := []struct {
		name         string
		entityTypeID model.ID
		perPage      uint
		page         uint
		selectSql    string
		sort         *db.Sort
		args         []interface{}
	}{
		{
			name:         "default",
			entityTypeID: 0,
			perPage:      10,
			page:         0,
			selectSql: "SELECT `e`.`id`, `e`.`entity_type_id`, `e`.`external_id`, `e`.`created_at`, `e`.`updated_at` " +
				"FROM `entities` AS `e` ORDER BY `e`.`updated_at` DESC LIMIT ?",
			args: []interface{}{int64(10)},
		},
		{
			name:         "entityTypeID",
			entityTypeID: model.ID(124),
			perPage:      10,
			page:         0,
			selectSql: "SELECT `e`.`id`, `e`.`entity_type_id`, `e`.`external_id`, `e`.`created_at`, `e`.`updated_at` " +
				"FROM `entities` AS `e` WHERE (`e`.`entity_type_id` = ?) " +
				"ORDER BY `e`.`updated_at` DESC LIMIT ?",
			args: []interface{}{"124", int64(10)},
		},
		{
			name:         "order-external-id",
			entityTypeID: model.ID(124),
			perPage:      18,
			page:         0,
			sort:         db.NewSort([]string{"externalId"}).Add("externalId", db.DESCOrder),
			selectSql:    "SELECT `e`.`id`, `e`.`entity_type_id`, `e`.`external_id`, `e`.`created_at`, `e`.`updated_at` " +
				"FROM `entities` AS `e` WHERE (`e`.`entity_type_id` = ?) " +
				"ORDER BY `e`.`external_id` DESC LIMIT ?",
			args:         []interface{}{"124", int64(18)},
		},
		{
			name:         "order-name-pagination",
			entityTypeID: model.ID(0),
			perPage:      25,
			page:         3,
			selectSql:    "SELECT `e`.`id`, `e`.`entity_type_id`, `e`.`external_id`, `e`.`created_at`, `e`.`updated_at` " +
				"FROM `entities` AS `e` ORDER BY `e`.`updated_at` DESC LIMIT ? OFFSET ?",
			args:         []interface{}{int64(25), int64(50)},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testting %s", tc.name)

			c := db.NewCursor(tc.page, tc.perPage, nil, []string{"name"})
			if tc.sort != nil {
				c.Sort = tc.sort
			}

			f := db.NewFilter([]string{"entityTypeId"})
			if tc.entityTypeID != 0 {
				f.Add("entityTypeId", strconv.Itoa(int(tc.entityTypeID)))
			}

			selectQuery, err := selectEntitiesQuery(c, f)
			if !assert.NoError(t, err) {
				t.Fatalf("unexpected error %s", err.Error())
			}

			assert.Equal(t, tc.selectSql, selectQuery.selectSQL)

			if !assert.Equal(t, len(tc.args), len(selectQuery.selectArgs)) {
				t.Fatalf("expected query arguments count %d != actual %d", len(tc.args), len(selectQuery.selectArgs))
			}

			for i := range tc.args {
				assert.Equal(t, tc.args[i], selectQuery.selectArgs[i])
			}
		})
	}
}

func Test_firstEntityByIDQuery(t *testing.T) {
	t.Run("valid ID", func(t *testing.T) {
		expected := "SELECT `e`.`id` AS `entity_id`, `e`.`entity_type_id` AS `entity_type_id`, `et`.`service_id` AS `service_id`, " +
			"`e`.`external_id` AS `entity_external_id`, `et`.`name` AS `entity_type_name`, `et`.`description` AS `entity_type_description`, " +
			"`ms`.`name` AS `service_name`, `ms`.`description` AS `service_description`, `e`.`created_at` AS `entity_created_at`, " +
			"`et`.`created_at` AS `entity_type_created_at`, `ms`.`created_at` AS `service_created_at`, `e`.`updated_at` AS `entity_updated_at`, " +
			"`et`.`updated_at` AS `entity_type_updated_at`, `ms`.`updated_at` AS `service_updated_at` FROM `entities` AS `e` " +
			"INNER JOIN `entity_types` AS `et` ON (`e`.`entity_type_id` = `et`.`id`) " +
			"INNER JOIN `microservices` AS `ms` ON (`et`.`service_id` = `ms`.`id`) WHERE `e`.`id` = ? LIMIT ?"

		ID := 123

		sql, args, err := firstEntityByIDQuery(model.ID(ID))
		assert.NoError(t, err)
		assert.Len(t, args, 2)
		assert.Equal(t, int64(ID), args[0])
		assert.Equal(t, int64(1), args[1]) // Limit
		assert.Equal(t, expected, sql)
	})
}
