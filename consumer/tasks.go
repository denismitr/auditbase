package consumer

import (
	"sync"

	"github.com/denismitr/auditbase/db"
	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/errtype"
	"github.com/denismitr/auditbase/utils/logger"
)

const ErrInvalidReceivedEvent = errtype.StringError("invalid received event")

type semaphore struct{}

type processor interface {
	process(flow.ReceivedEvent)
}

// tasts implements processor interface
type tasks struct {
	sem       chan semaphore
	eventsCh  chan flow.ReceivedEvent
	persister db.Persister
	ef        flow.EventFlow
	logger    logger.Logger
	mu        sync.Mutex
}

func newTasks(maxTasks int, logger logger.Logger, persister db.Persister, ef flow.EventFlow) *tasks {
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
	for re := range t.eventsCh {
		if re == nil {
			t.logger.Error(ErrInvalidReceivedEvent)
			return
		}

		event, err := re.Event()
		if err != nil {
			t.logger.Error(err)
			t.ef.Requeue(re)
			return
		}

		t.sem <- struct{}{}

		go func(re flow.ReceivedEvent, event model.Event) {
			if err := t.persister.Persist(&event); err != nil {
				t.logger.Error(err)

				if err := t.ef.Requeue(re); err != nil {
					t.logger.Error(err)
				}
				// fixme emit success process event
			} else {
				// fixme emit success process event
				if err := t.ef.Ack(re); err != nil {
					t.logger.Error(err)
				}
			}

			<-t.sem
		}(re, event)
	}
}

func (t *tasks) stop() {
	close(t.eventsCh)
}
