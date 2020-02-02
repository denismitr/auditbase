package consumer

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"sync"

	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/queue"
	"github.com/denismitr/auditbase/utils"
	"github.com/pkg/errors"
)

type StopFunc func(ctx context.Context) error

// Consumer - consumers from the event flow and
// persists events to the permanent storage
type Consumer struct {
	logger utils.Logger
	f      flow.EventFlow

	receiveCh           chan queue.ReceivedMessage
	stopCh              chan struct{}
	efStateCh           chan flow.State
	persistenceResultCh chan persistenceResult

	persistedEvents int
	failedEvents    int

	tasks *tasks

	mu       sync.RWMutex
	statusOK bool
}

// New consumer
func New(
	eventFlow flow.EventFlow,
	logger utils.Logger,
	mq queue.MQ,
	microservices model.MicroserviceRepository,
	events model.EventRepository,
	targetTypes model.TargetTypeRepository,
	actorTypes model.ActorTypeRepository,
) *Consumer {
	resultCh := make(chan persistenceResult)

	persister := newDBPersister(
		microservices,
		events,
		targetTypes,
		actorTypes,
		logger,
		resultCh,
	)

	tasks := newTasks(10, persister)

	return &Consumer{
		f:                   eventFlow,
		tasks:               tasks,
		logger:              logger,
		persistenceResultCh: resultCh,
		receiveCh:           make(chan queue.ReceivedMessage),
		stopCh:              make(chan struct{}),
		efStateCh:           make(chan flow.State),
		mu:                  sync.RWMutex{},
		statusOK:            true,
	}
}

// Start consumer
func (c *Consumer) Start(consumerName string) StopFunc {
	go c.healthCheck()
	go c.tasks.run()
	go c.processEvents(consumerName)

	return func(ctx context.Context) error {
		close(c.stopCh)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.receiveCh:
			return nil
		}
	}
}

func (c *Consumer) processEvents(consumerName string) {
	c.f.NotifyOnStateChange(c.efStateCh)

	events := c.f.Receive(consumerName)

	for {
		select {
		case e := <-events:
			// event flow failed
			if e == nil {
				c.markAsFailed()
				continue
			}

			c.tasks.process(e)
		case efState := <-c.efStateCh:
			if efState == flow.Failed || efState == flow.Stopped {
				c.markAsFailed()
			}
		case result := <-c.persistenceResultCh:
			c.registerResult(result)
		case <-c.stopCh:
			c.logger.Debugf("Received on stop channel")
			c.markAsFailed()
			c.f.Stop()
			return
		}
	}
}

func (c *Consumer) registerResult(r persistenceResult) {
	switch r {
	case eventFlowFailed:
		c.incrementFailedEvents()
		c.markAsFailed()
	case success:
		c.incrementPersistedEvents()
	case databaseFailed:
		c.incrementFailedEvents()
	case eventCouldNotBeProcessed:
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
