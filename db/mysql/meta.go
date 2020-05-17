package mysql

import "github.com/denismitr/auditbase/model"

type selectWithMetaQuery struct {
	query string
	count string
	queryArgs map[string]interface{}
	countArgs map[string]interface{}
}

type meta struct {
	Total int `db:"total"`
}

func (m *meta) ToModel(p *model.Pagination) *model.Meta {
	var hasNext bool
	n := float64(m.Total) / (float64(p.Page) * float64(p.PerPage))
	if n <= 1 {
		hasNext = false
	} else {
		hasNext = true
	}

	return &model.Meta{
		Total:       m.Total,
		Page:        p.Page,
		PerPage:     p.PerPage,
		HasNextPage: hasNext,
	}
}
