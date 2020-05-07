package consumer

import (
	"context"
	"encoding/json"
	"github.com/denismitr/auditbase/db"
	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/queue"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/pkg/errors"
	"net/http"
	"os"
	"sync"
)

// Consumer - consumers from the event flow and
// persists events to the permanent storage
type Consumer struct {
	logger    logger.Logger
	eventFlow flow.EventFlow
	persister db.Persister

	receiveCh        chan queue.ReceivedMessage
	stopCh           chan struct{}
	eventFlowStateCh chan flow.State
	pResultCh        chan db.PersistenceResult

	persistedEvents int
	failedEvents    int

	tasks *tasks

	mu       sync.RWMutex
	statusOK bool
}

// New consumer
func New(
	ef flow.EventFlow,
	logger logger.Logger,
	persister db.Persister,
) *Consumer {
	pResultCh := make(chan db.PersistenceResult)
	persister.NotifyOnResult(pResultCh)
	tasks := newTasks(10, logger, persister, ef)

	return &Consumer{
		eventFlow:        ef,
		tasks:            tasks,
		persister:        persister,
		logger:           logger,
		pResultCh:        pResultCh,
		receiveCh:        make(chan queue.ReceivedMessage),
		stopCh:           make(chan struct{}),
		eventFlowStateCh: make(chan flow.State),
		mu:               sync.RWMutex{},
		statusOK:         true,
	}
}

type StopFunc func(ctx context.Context) error

// Start consumer
func (c *Consumer) Start(queueName, consumerName string) StopFunc {
	go c.healthCheck()
	go c.tasks.run()
	go c.processEvents(queueName, consumerName)

	return c.stop
}

func (c *Consumer) stop(ctx context.Context) error {
	close(c.stopCh)

	for {
		select {
		case <-ctx.Done():
			return nil
		}
	}
}

func (c *Consumer) processEvents(queue, consumerName string) {
	c.eventFlow.NotifyOnStateChange(c.eventFlowStateCh)

	events := c.eventFlow.Receive(queue, consumerName)

	for {
		select {
		case e := <-events:
			// event flow failed
			if e == nil {
				c.markAsFailed()
				continue
			}

			c.tasks.process(e)
		case efState := <-c.eventFlowStateCh:
			if efState == flow.Failed || efState == flow.Stopped {
				c.markAsFailed()
			}
		case result := <-c.pResultCh:
			c.registerResult(result)
		case <-c.stopCh:
			c.logger.Debugf("Received on stop channel")
			c.markAsFailed()
			c.eventFlow.Stop()
			c.tasks.stop()
			return
		}
	}
}

func (c *Consumer) registerResult(r db.PersistenceResult) {
	switch r {
	case db.EventFlowFailed:
		c.incrementFailedEvents()
		c.markAsFailed()
	case db.Success:
		c.incrementPersistedEvents()
	case db.CriticalDatabaseFailure, db.LogicalError, db.EventCouldNotBeProcessed:
		c.incrementFailedEvents()
	}
}

func (c *Consumer) statusIsOK() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.statusOK
}

func (c *Consumer) healthCheck() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if c.statusIsOK() {
			w.WriteHeader(200)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		} else {
			c.logger.Debugf("Health check failed")
			w.WriteHeader(500)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "error"})
		}
	})

	c.logger.Debugf("\nStarting healthcheck on port %s", os.Getenv("HEALTH_PORT"))
	err := http.ListenAndServe(":"+os.Getenv("HEALTH_PORT"), nil)
	if err != nil {
		c.logger.Error(errors.Wrap(err, "helthcheck endpoint failed"))
	}
}

func (c *Consumer) markAsFailed() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.statusOK = false
}

func (c *Consumer) incrementFailedEvents() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failedEvents++
}

func (c *Consumer) incrementPersistedEvents() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.persistedEvents++
}
