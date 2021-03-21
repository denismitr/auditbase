package mysql

import (
	"github.com/denismitr/auditbase/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

//func TestCreateSelectEntitiesQuery(t *testing.T) {
//	nameAsc := model.ASCOrder
//	_ = model.DESCOrder
//
//	tt := []struct{
//		name string
//		serviceId string
//		nameOrder *model.Order
//		perPage int
//		page int
//		sql string
//		args []interface{}
//		err error
//	}{
//		{
//			name: "default",
//			serviceId: "",
//			nameOrder: nil,
//			perPage: 10,
//			page: 0,
//			sql: "SELECT BIN_TO_UUID(id) as id, BIN_TO_UUID(service_id) as service_id, name, description, created_at, updated_at FROM entities ORDER BY created_at DESC LIMIT 10 OFFSET 0",
//			args: nil,
//		},
//		{
//			name: "serviceId",
//			serviceId: "18420701-7852-4361-8e91-a660385dd0c5",
//			nameOrder: nil,
//			perPage: 10,
//			page: 0,
//			sql: "SELECT BIN_TO_UUID(id) as id, BIN_TO_UUID(service_id) as service_id, name, description, created_at, updated_at FROM entities WHERE service_id = uuid_to_bin(?) ORDER BY service_id DESC LIMIT 10 OFFSET 0",
//			args:[]interface{}{"18420701-7852-4361-8e91-a660385dd0c5"},
//		},
//		{
//			name: "order-name",
//			serviceId: "18420701-7852-4361-8e91-a660385dd0c5",
//			nameOrder: &nameAsc,
//			perPage: 10,
//			page: 0,
//			sql: "SELECT BIN_TO_UUID(id) as id, BIN_TO_UUID(service_id) as service_id, name, description, created_at, updated_at FROM entities WHERE service_id = uuid_to_bin(?) ORDER BY name ? LIMIT 10 OFFSET 0",
//			args:[]interface{}{"18420701-7852-4361-8e91-a660385dd0c5", "ASC"},
//		},
//		{
//			name: "order-name-pagination",
//			serviceId: "",
//			nameOrder: &nameAsc,
//			perPage: 25,
//			page: 3,
//			sql: "SELECT BIN_TO_UUID(id) as id, BIN_TO_UUID(service_id) as service_id, name, description, created_at, updated_at FROM entities ORDER BY name ? LIMIT 25 OFFSET 50",
//			args:[]interface{}{"ASC"},
//		},
//	}
//
//	for _, tc := range tt {
//		t.Run(tc.name, func(t *testing.T) {
//			f := model.NewFilter([]string{"serviceId"})
//			if tc.serviceId != "" {
//				f.Add("serviceId", tc.serviceId)
//			}
//
//			s := model.NewSort()
//			if tc.nameOrder != nil {
//				s.Add("name", *tc.nameOrder)
//			}
//
//			p := &model.Pagination{Page: tc.page, PerPage: tc.perPage}
//
//			sql, args, err := selectEntitiesQuery(f, s, p)
//
//			if tc.err == nil {
//				assert.NoError(t, err)
//			} else {
//				assert.Equal(t, tc.err, err)
//			}
//
//			assert.Equal(t, tc.sql, sql)
//			assert.Equal(t, tc.args, args)
//		})
//	}
//}

func Test_firstEntityByIDQuery(t *testing.T) {
	t.Run("valid ID", func(t *testing.T) {
		expected := "SELECT bin_to_uuid(`e`.`id`) AS `entity_id`, bin_to_uuid(`e`.`entity_type_id`) AS `entity_type_id`, bin_to_uuid(`et`.`service_id`) AS `service_id`, `e`.`external_id` AS `entity_external_id`, `et`.`name` AS `entity_type_name`, `et`.`description` AS `entity_type_description`, `ms`.`name` AS `service_name`, `ms`.`description` AS `service_description`, `e`.`created_at` AS `entity_created_at`, `et`.`created_at` AS `entity_type_created_at`, `ms`.`created_at` AS `service_created_at`, `e`.`updated_at` AS `entity_updated_at`, `et`.`updated_at` AS `entity_type_updated_at`, `ms`.`updated_at` AS `service_updated_at` FROM `entities` AS `e` INNER JOIN `entity_types` AS `et` ON (`e`.`entity_type_id` = `et`.`id`) INNER JOIN `microservices` AS `ms` ON (`et`.`service_id` = `ms`.`id`) WHERE `e`.`id` = uuid_to_bin(?) LIMIT ?"
		ID := 123

		sql, args, err := firstEntityByIDQuery(model.ID(ID))
		assert.NoError(t, err)
		assert.Len(t, args, 2)
		assert.Equal(t, args[0], ID)
		assert.Equal(t, int64(1), args[1]) // Limit
		assert.Equal(t, expected, sql)
	})
}
