package consumer

import (
	"sync"

	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/utils"
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
	ef        flow.EventFlow
	logger    utils.Logger
	mu        sync.Mutex
}

func newTasks(maxTasks int, logger utils.Logger, persister persister, ef flow.EventFlow) *tasks {
	return &tasks{
		sem:       make(chan semaphore, maxTasks),
		eventsCh:  make(chan flow.ReceivedEvent),
		logger:    logger,
		mu:        sync.Mutex{},
		ef:        ef,
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
			if err := t.persister.persist(event); err != nil {
				if err := t.ef.Requeue(event); err != nil {
					t.logger.Error(err)
				}
				// fixme emit success process event
			} else {
				// fixme emit success process event
				if err := t.ef.Ack(event); err != nil {
					t.logger.Error(err)
				}
			}

			<-t.sem
		}()
	}
}

func (t *tasks) stop() {
	close(t.eventsCh)
}
