package consumer

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/queue"
	"github.com/denismitr/auditbase/utils"
	"github.com/pkg/errors"
)

// Consumer - consumers from the event flow and
// persists events to the permanent storage
type Consumer struct {
	logger    utils.Logger
	eventFlow flow.EventFlow

	receiveCh           chan queue.ReceivedMessage
	stopCh              chan struct{}
	eventFlowStateCh    chan flow.State
	persistenceResultCh chan persistenceResult

	persistedEvents int
	failedEvents    int

	tasks *tasks

	mu       sync.RWMutex
	statusOK bool
}

// New consumer
func New(
	ef flow.EventFlow,
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

	tasks := newTasks(10, logger, persister, ef)

	return &Consumer{
		eventFlow:           ef,
		tasks:               tasks,
		logger:              logger,
		persistenceResultCh: resultCh,
		receiveCh:           make(chan queue.ReceivedMessage),
		stopCh:              make(chan struct{}),
		eventFlowStateCh:    make(chan flow.State),
		mu:                  sync.RWMutex{},
		statusOK:            true,
	}
}

// Start consumer
func (c *Consumer) Start(ctx context.Context, queueName, consumerName string) {
	go c.healthCheck()
	go c.tasks.run()
	go c.processEvents(queueName, consumerName)

	for {
		select {
		case <-ctx.Done():
			close(c.stopCh)
			time.Sleep(10 * time.Second)
			return
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

			go c.tasks.process(e)
		case efState := <-c.eventFlowStateCh:
			if efState == flow.Failed || efState == flow.Stopped {
				c.markAsFailed()
			}
		case result := <-c.persistenceResultCh:
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
