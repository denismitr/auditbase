package db

import "sync"

type propertySink struct {
	mu      sync.RWMutex
	results map[string]string
	errs    []error
}

func (s *propertySink) add(name, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id != "" {
		s.results[name] = id
	}
}

func (s *propertySink) err(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errs = append(s.errs, err)
}

func (s *propertySink) hasErrors() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.errs) > 0
}

func (s *propertySink) firstError() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.errs[0]
}

func (s *propertySink) all() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.results
}

func (s *propertySink) count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.results)
}

func newPropertySink() *propertySink {
	return &propertySink{
		results: make(map[string]string),
		errs:    make([]error, 0),
	}
}
