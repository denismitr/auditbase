package errbag

type ErrorBag struct {
	bag map[string][]error
}

func (eb *ErrorBag) NotEmpty() bool {
	return len(eb.bag) > 0
}

func (eb *ErrorBag) IsEmpty() bool {
	return len(eb.bag) == 0
}

// All errors as flat slice
func (eb *ErrorBag) All() []error {
	var errors []error
	for _, bag := range eb.bag {
		for i := range bag {
			errors = append(errors, bag[i])
		}
	}
	return errors
}

func (eb *ErrorBag) Add(key string, err error) {
	eb.bag[key] = append(eb.bag[key], err)
}

func New() *ErrorBag {
	return &ErrorBag{bag: make(map[string][]error)}
}
