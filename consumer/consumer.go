package consumer

import (
	"context"
	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/queue"
	"github.com/denismitr/auditbase/utils/logger"
	"sync"
	"time"
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

	received map[string]flow.ReceivedEvent

	mu              sync.RWMutex
	persistedEvents int
	failedEvents    int
	statusOK        bool
	startedAt       time.Time
	failedAt        time.Time
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
	c.mu.Lock()
	c.statusOK = true
	c.startedAt = time.Now()
	c.mu.Unlock()

	go c.healthCheck()
	go c.processEvents(queueName, consumerName)

	return c.persister.Run(ctx)
}

func (c *Consumer) processEvents(queue, consumerName string) {
	c.mu.Lock()
	c.eventFlow.NotifyOnStateChange(c.eventFlowStateCh)
	c.mu.Unlock()

	events := c.eventFlow.Receive(queue, consumerName)

	for {
		select {
		case re := <-events:
			// event flow failed
			if re == nil {
				c.logger.Debugf("EMPTY EVENT RECEIVED")
				continue
			}

			e, err := re.Event()
			if err != nil {
				c.logger.Error(err)
				if err := c.eventFlow.Reject(re); err != nil {
					c.mu.Lock()
					c.failedEvents++
					c.mu.Unlock()
					c.logger.Error(err)
				}
			}

			c.mu.Lock()
			c.received[e.ID] = re
			c.mu.Unlock()
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
	c.mu.Lock()
	defer c.mu.Unlock()

	if r.Ok() {
		if re, ok := c.received[r.ID()]; ok {
			if err := c.eventFlow.Ack(re); err != nil {
				c.logger.Error(err)
			}

			delete(c.received, r.ID())
			c.persistedEvents++
			c.logger.Debugf("successfully processed event with ID %s", r.ID())
		}

		return
	}

	c.logger.Error(r.Err())
	c.failedEvents++

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

func (c *Consumer) markAsFailed() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.statusOK = false
	c.failedAt = time.Now()
}
