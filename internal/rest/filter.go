package rest

import (
	"github.com/denismitr/auditbase/internal/db"
	"net/url"
	"strconv"
)

func createFilter(q url.Values, allowedKeys []string) *db.Filter {
	f := db.NewFilter(allowedKeys)

	for k, v := range q {
		if f.Allows(k) && len(v) == 1 && v[0] != "" {
			f.Add(k, v[0])
		}
	}

	return f
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

