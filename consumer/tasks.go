package consumer

import (
	"sync"

	"github.com/denismitr/auditbase/flow"
)

type semaphore int

type processor interface {
	process(flow.ReceivedEvent)
}

// tasts implements processor interface
type tasks struct {
	sem       chan semaphore
	eventsCh  chan flow.ReceivedEvent
	persister persister
	mu        sync.Mutex
}

func newTasks(maxTasks int, persister persister) *tasks {
	return &tasks{
		sem:       make(chan semaphore, maxTasks),
		eventsCh:  make(chan flow.ReceivedEvent),
		mu:        sync.Mutex{},
		persister: persister,
	}
}

func (t *tasks) process(e flow.ReceivedEvent) {
	t.eventsCh <- e
}

func (t *tasks) run() {
	for e := range t.eventsCh {
		if e == nil {
			return
		}

		event := e
		t.sem <- 1

		go func() {
			t.persister.persist(event)
			<-t.sem
		}()
	}
}

func (t *tasks) stop() {
	close(t.eventsCh)
}
