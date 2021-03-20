package db

import (
	"fmt"
	"github.com/denismitr/auditbase/model"
	"strconv"
)

type Filter struct {
	allowed []string
	ids     []model.ID
	items   map[string]string
}

func NewFilter(allowed []string) *Filter {
	return &Filter{
		items:   make(map[string]string),
		allowed: allowed,
	}
}

func (f *Filter) Empty() bool {
	return len(f.items) == 0
}

func (f *Filter) Allows(k string) bool {
	for i := range f.allowed {
		if k == f.allowed[i] {
			return true
		}
	}

	return false
}

func (f *Filter) Add(k, v string) *Filter {
	f.items[k] = v
	return f
}

func (f *Filter) ByIDs(ids []model.ID) *Filter {
	f.ids = ids
	return f
}

func (f *Filter) Has(k string) bool {
	if _, ok := f.items[k]; ok {
		return true
	}

	return false
}

func (f *Filter) HasIDs() bool {
	return len(f.ids) > 0
}

func (f *Filter) IDs() []model.ID {
	return f.ids
}

func (f *Filter) StringOrDefault(k, d string) string {
	if f.Has(k) {
		return f.items[k]
	}

	return d
}

func (f *Filter) MustString(k string) string {
	if !f.Has(k) {
		panic(fmt.Sprintf("no suchkey in filter %s", k))
	}

	return f.items[k]
}

func (f *Filter) IntOrDefault(k string, d int) int {
	if f.Has(k) {
		v := f.items[k]
		n, err := strconv.Atoi(v)
		if err != nil {
			return d
		}
		return n
	}

	return d
}
