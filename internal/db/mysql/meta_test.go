package mysql

//import (
//	"github.com/denismitr/auditbase/model"
//	"github.com/stretchr/testify/assert"
//	"testing"
//)
//
//func TestItCanConvertItselfToModel(t *testing.T) {
//	tt := []struct{
//		name string
//		total int
//		page int
//		perPage int
//		hasNextPage bool
//		p *model.Pagination
//	}{
//		{
//			name: "1-result",
//			total: 1,
//			page: 1,
//			perPage: 25,
//			hasNextPage: false,
//			p: &model.Pagination{Page: 1, PerPage: 25},
//		},
//		{
//			name: "10-results",
//			total: 10,
//			page: 1,
//			perPage: 25,
//			hasNextPage: false,
//			p: &model.Pagination{Page: 1, PerPage: 25},
//		},
//		{
//			name: "80-results",
//			total: 80,
//			page: 1,
//			perPage: 25,
//			hasNextPage: true,
//			p: &model.Pagination{Page: 1, PerPage: 25},
//		},
//		{
//			name: "99-results",
//			total: 99,
//			page: 4,
//			perPage: 25,
//			hasNextPage: false,
//			p: &model.Pagination{Page: 4, PerPage: 25},
//		},
//		{
//			name: "101-results",
//			total: 101,
//			page: 4,
//			perPage: 25,
//			hasNextPage: true,
//			p: &model.Pagination{Page: 4, PerPage: 25},
//		},
//		{
//			name: "100-results",
//			total: 100,
//			page: 4,
//			perPage: 25,
//			hasNextPage: false,
//			p: &model.Pagination{Page: 4, PerPage: 25},
//		},
//		{
//			name: "200-results",
//			total: 200,
//			page: 3,
//			perPage: 20,
//			hasNextPage: true,
//			p: &model.Pagination{Page: 3, PerPage: 20},
//		},
//	}
//
//	for _, tc := range tt {
//		t.Run(tc.name, func(t *testing.T) {
//			m := meta{Total: tc.total}
//			mm := m.ToModel(tc.p)
//			assert.Equal(t, tc.total, mm.Total)
//			assert.Equal(t, tc.hasNextPage, mm.HasNextPage)
//			assert.Equal(t, tc.page, mm.Page)
//			assert.Equal(t, tc.perPage, mm.PerPage)
//		})
//	}
//}