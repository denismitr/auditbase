package db

type Include struct {
	allowed []string
	items   []string
}

func NewInclude(allowed []string) *Include {
	return &Include{
		allowed: allowed,
	}
}

func (inc *Include) Allows(k string) bool {
	for _, allowed := range inc.allowed {
		if k == allowed {
			return true
		}
	}

	return false
}

func (inc *Include) Add(k string) *Include {
	inc.items = append(inc.items, k)
	return inc
}


func (inc *Include) Has(k string) bool {
	for _, item := range inc.items {
		if item == k {
			return true
		}
	}

	return false
}
