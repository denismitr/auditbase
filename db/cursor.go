package db

import "time"

type Order string

func (o Order) String() string {
	return string(o)
}

const (
	DESCOrder Order = "DESC"
	ASCOrder Order = "ASC"
	DefaultPerPage uint = 50
)

type Sort struct {
	allowedSortColumns []string
	items map[string]Order
}

func NewSort(allowedSortColumns []string) *Sort {
	return &Sort{
		allowedSortColumns: allowedSortColumns,
		items: make(map[string]Order),
	}
}

func (s *Sort) Allows(k string) bool {
	for _, column := range s.allowedSortColumns {
		if column == k {
			return true
		}
	}

	return false
}

func (s *Sort) Add(k string, o Order) *Sort {
	s.items[k] = o
	return s
}

func (s *Sort) Has(k string) bool {
	if _, ok := s.items[k]; ok {
		return true
	}

	return false
}

func (s *Sort) All()  map[string]Order {
	return s.items
}

func (s *Sort) GetOrDefault(k string, d Order) Order {
	v, ok := s.items[k]
	if !ok {
		return d
	}

	return v
}

type Cursor struct {
	Sort *Sort
	LastCreatedAt *time.Time
	Page uint
	PerPage uint
}

func (c *Cursor) Offset() uint {
	if c.Page <= 1 {
		return 0
	}

	return (c.Page - 1) * c.PerPage
}

func NewCursor(page, perPage uint, lastCreatedAt *time.Time, allowedSortColumns []string) *Cursor {
	return &Cursor{
		Sort: NewSort(allowedSortColumns),
		Page: page,
		PerPage: perPage,
		LastCreatedAt: lastCreatedAt,
	}
}
