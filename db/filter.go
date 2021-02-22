package db

import (
	"strconv"
	"fmt"
)

type Filter struct {
	allowed []string
	items map[string]string
	includeCount []string
}

func NewFilter(allowed []string) *Filter {
	return &Filter{
		items: make(map[string]string),
		allowed: allowed,
		includeCount: make([]string, 0),
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

func (f *Filter) IncludeCount(includes ...string) *Filter {
	f.includeCount = append(f.includeCount, includes...)
	return f
}

func (f *Filter) Add(k, v string) *Filter {
	f.items[k] = v
	return f
}

func (f *Filter) Has(k string) bool {
	if _, ok := f.items[k]; ok {
		return true
	}

	return false
}

func (f *Filter) StringOrDefault(k, d string) string {
	if f.Has(k) {
		return f.items[k]
	}

	return d
}

func (f *Filter) MustString(k string) string {
	if ! f.Has(k) {
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
