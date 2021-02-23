package consumer

import (
	"context"
	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/queue"
	"github.com/denismitr/auditbase/service"
	"github.com/denismitr/auditbase/utils/logger"
	"os"
	"sync"
	"time"
)

// Consumer - consumers from the event flow and
// persists events to the permanent storage
type Consumer struct {
	lg            logger.Logger
	actionFlow    flow.EventFlow
	actionService service.ActionService

	receiveCh        chan queue.ReceivedMessage
	eventFlowStateCh chan flow.State

	mu                    sync.RWMutex
	persistedEvents       int
	failedActionCreations int
	statusOK              bool
	startedAt             time.Time
	failedAt              time.Time
}

// New consumer
func New(
	ef flow.EventFlow,
	lg logger.Logger,
	actionService service.ActionService,
) *Consumer {
	return &Consumer{
		actionFlow:       ef,
		actionService:    actionService,
		lg:               lg,
		receiveCh:        make(chan queue.ReceivedMessage),
		eventFlowStateCh: make(chan flow.State),
		mu:               sync.RWMutex{},
		statusOK:         true,
	}
}

// Start consumer
func (c *Consumer) Start(stopCh <-chan os.Signal, queueName, consumerName string) error {
	c.mu.Lock()
	c.statusOK = true
	c.startedAt = time.Now()
	c.mu.Unlock()

	go c.healthCheck(stopCh)

	return c.processEvents(stopCh, queueName, consumerName)
}

func (c *Consumer) processEvents(stopCh <-chan os.Signal, queue, consumerName string) error {
	c.mu.Lock()
	c.actionFlow.NotifyOnStateChange(c.eventFlowStateCh)
	c.mu.Unlock()

	events := c.actionFlow.Receive(queue, consumerName)
	sem := make(chan struct{}, 4) // fixme: config

	for {
		select {
		case re := <-events:
			// action flow failed
			if re == nil {
				c.lg.Debugf("EMPTY EVENT RECEIVED")
				continue
			}

			sem <- struct{}{}
			go func() {
				defer func() { <- sem }()

				if err := c.handleNewAction(re); err != nil {
					c.lg.Error(err)
					if err := c.actionFlow.Reject(re); err != nil {
						c.mu.Lock()
						c.failedActionCreations++
						c.mu.Unlock()
						c.lg.Error(err)
					}
				}
			}()
		case efState := <-c.eventFlowStateCh:
			if efState == flow.Failed || efState == flow.Stopped {
				c.markAsFailed()
			}
		case <-stopCh:
			c.lg.Debugf("Received on stop channel")
			c.markAsFailed()
			_ = c.actionFlow.Stop()
			return nil
		}
	}
}

func (c *Consumer) handleNewAction(ra flow.ReceivedAction) error {
	na, err := ra.NewAction()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	if _, err := c.actionService.Create(ctx, na); err != nil {
		return err
	}

	return nil
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
