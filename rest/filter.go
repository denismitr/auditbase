package rest

import (
	"github.com/denismitr/auditbase/model"
	"net/url"
	"strconv"
	"strings"
)

func createFilter(q url.Values, allowedKeys []string) *model.Filter {
	f := model.NewFilter(allowedKeys)

	for k, v := range convertQueryArrayToSimpleMap("filter", q) {
		if f.Allows(k) {
			f.Add(k, v)
		}
	}

	return f
}

func convertQueryArrayToSimpleMap(key string, q url.Values) map[string]string {
	result := make(map[string]string)

	prefix := key + "["
	suffix := "]"
	for k, values := range q {
		if strings.HasPrefix(k, prefix) && strings.HasSuffix(k, suffix) {
			if len(values) > 0 {
				i := strings.Index(k, prefix) + len(prefix)
				j := strings.Index(k, suffix)
				param := k[i:j]
				result[param] = values[0]
			}
		}
	}

	return result
}

func createSort(q url.Values) *model.Sort {
	s := model.NewSort()

	for k, value := range convertQueryArrayToSimpleMap("sort", q) {
		v := strings.ToUpper(value)
		if v == string(model.ASCOrder) || v == string(model.DESCOrder) {
			s.Add(k, model.Order(v))
		}
	}

	return s
}

func createPagination(q url.Values, maxPerPage int) *model.Pagination {
	pagination := new(model.Pagination)

	if v, ok := q["page"]; ok && len(v) > 0 {
		if p, err := strconv.Atoi(v[0]); err == nil {
			pagination.Page = p
		}
	}

	if v, ok := q["perPage"]; ok && len(v) > 0 {
		if pp, err := strconv.Atoi(v[0]); err == nil {
			pagination.PerPage = pp
		}
	}

	if pagination.Page == 0 {
		pagination.Page = 1
	}

	if pagination.PerPage == 0 {
		pagination.PerPage = maxPerPage
	}

	return pagination
}

