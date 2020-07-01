package consumer

import (
	"context"
	"encoding/json"
	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/model"
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
	persister model.EventPersister

	receiveCh        chan queue.ReceivedMessage
	stopCh           chan struct{}
	eventFlowStateCh chan flow.State
	resultCh         chan model.EventPersistenceResult

	persistedEvents int
	failedEvents    int

	received map[string]flow.ReceivedEvent

	mu       sync.RWMutex
	statusOK bool
}

// New consumer
func New(
	ef flow.EventFlow,
	logger logger.Logger,
	persister model.EventPersister,
) *Consumer {
	resultCh := make(chan model.EventPersistenceResult)
	persister.NotifyOnResult(resultCh)

	return &Consumer{
		eventFlow:        ef,
		persister:        persister,
		logger:           logger,
		resultCh:         resultCh,
		receiveCh:        make(chan queue.ReceivedMessage),
		stopCh:           make(chan struct{}),
		eventFlowStateCh: make(chan flow.State),
		received:         make(map[string]flow.ReceivedEvent),
		mu:               sync.RWMutex{},
		statusOK:         true,
	}
}

// Start consumer
func (c *Consumer) Start(ctx context.Context, queueName, consumerName string) error {
	go c.healthCheck()
	go c.processEvents(queueName, consumerName)

	return c.persister.Run(ctx)
}

func (c *Consumer) processEvents(queue, consumerName string) {
	c.eventFlow.NotifyOnStateChange(c.eventFlowStateCh)

	events := c.eventFlow.Receive(queue, consumerName)

	for {
		select {
		case re := <-events:
			// event flow failed
			if re == nil {
				c.markAsFailed()
				continue
			}

			e, err := re.Event()
			if err != nil {
				c.logger.Error(err)
				if err := c.eventFlow.Reject(re); err != nil {
					c.logger.Error(err)
				}
			}

			c.received[e.ID] = re
			c.persister.Persist(e)
		case efState := <-c.eventFlowStateCh:
			if efState == flow.Failed || efState == flow.Stopped {
				c.markAsFailed()
			}
		case result := <-c.resultCh:
			c.handleResult(result)
		case <-c.stopCh:
			c.logger.Debugf("Received on stop channel")
			c.markAsFailed()
			_ = c.eventFlow.Stop()
			return
		}
	}
}

func (c *Consumer) handleResult(r model.EventPersistenceResult) {
	if r.Ok() {
		if re, ok := c.received[r.ID()]; ok {
			if err := c.eventFlow.Ack(re); err != nil {
				c.logger.Error(err)
			}

			delete(c.received, r.ID())
			c.logger.Debugf("successfully processed event with ID %s", r.ID())
		}

		return
	}

	c.logger.Error(r.Err())

	if re, ok := c.received[r.ID()]; ok {
		if err := c.eventFlow.Requeue(re); err != nil {
			c.logger.Error(err)
		} else {
			c.logger.Debugf("requeued event with ID %s", r.ID())
		}

		delete(c.received, r.ID())
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
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		} else {
			c.logger.Debugf("Health check failed")
			w.WriteHeader(500)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error"})
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
