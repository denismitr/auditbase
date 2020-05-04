package model

import "strconv"

type Order string

func (o Order) String() string {
	return string(o)
}

const DESCOrder Order = "DESC"
const ASCOrder Order = "ASC"

type Sort struct {
	items map[string]Order
}

func NewSort() *Sort {
	return &Sort{
		items: make(map[string]Order),
	}
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

func (s *Sort) GetOrDefault(k string, d Order) Order {
	v, ok := s.items[k]
	if !ok {
		return d
	}

	return v
}

func (s *Sort) Empty() bool {
	return len(s.items) == 0
}

type Filter struct {
	allowed []string
	items map[string]string
}

func NewFilter(allowed []string) *Filter {
	return &Filter{
		items: make(map[string]string),
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

type Pagination struct {
	Page    int
	PerPage int
}

func (p *Pagination) Offset() int {
	return (p.Page - 1) * p.PerPage
}

type EventFilter struct {
	ActorID          string
	ActorEntityID    string
	ActorEntityName  string
	TargetID         string
	TargetEntityID   string
	TargetEntityName string
	ActorServiceID   string
	TargetServiceID  string
	EventName        string
	EmittedAtGt      int64
	EmittedAtLt      int64
}

func (f *EventFilter) Empty() bool {
	return f.ActorID == "" && f.ActorEntityID == "" && f.EventName == "" && f.TargetID == ""
}

type EntityFilter struct {
	ServiceID string
}

func (f *EntityFilter) Empty() bool {
	return f.ServiceID == ""
}