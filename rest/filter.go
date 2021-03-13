package rest

import (
	"github.com/denismitr/auditbase/db"
	"net/url"
	"strconv"
	"strings"
)

func createFilter(q url.Values, allowedKeys []string) *db.Filter {
	f := db.NewFilter(allowedKeys)

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

func createCursor(q url.Values, maxPerPage int, allowedSortColumns []string) *db.Cursor {
	cursor := new(db.Cursor)

	if v, ok := q["page"]; ok && len(v) > 0 {
		if p, err := strconv.Atoi(v[0]); err == nil {
			cursor.Page = uint(p)
		}
	}

	if v, ok := q["perPage"]; ok && len(v) > 0 {
		if pp, err := strconv.Atoi(v[0]); err == nil {
			cursor.PerPage = uint(pp)
		}
	}

	if cursor.Page == 0 {
		cursor.Page = 1
	}

	if cursor.PerPage == 0 {
		cursor.PerPage = uint(maxPerPage)
	}

	s := db.NewSort(allowedSortColumns)
	cursor.Sort = s

	return cursor
}

