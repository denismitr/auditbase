package consumer

import (
	"context"
	"errors"
	"github.com/denismitr/auditbase/internal/flow"
	"github.com/denismitr/auditbase/internal/model"
	"github.com/denismitr/auditbase/internal/service"
	"github.com/denismitr/auditbase/internal/utils/logger"
	"os"
	"sync"
	"time"
)

type Stats struct {
	mu                    sync.RWMutex
	processedActions      int
	failedActionCreations int
	statusOK              bool
	startedAt             time.Time
	finishedAt           time.Time
}

// Consumer - consumers from the action flow and
// persists actions to the permanent storage
type Consumer struct {
	lg            logger.Logger
	actionFlow    flow.ActionFlow
	actionService service.ActionService
	consumerName       string
	stats Stats
}

// New consumer
func New(
	consumerName string,
	ef flow.ActionFlow,
	lg logger.Logger,
	actionService service.ActionService,
) *Consumer {
	return &Consumer{
		actionFlow:         ef,
		actionService:      actionService,
		lg:                 lg,
		consumerName:       consumerName,
	}
}

var ErrConnectionLoss = errors.New("connection to message broker was lost")
var ErrInterrupted = errors.New("consumer was interrupted")

// Start consumer
func (c *Consumer) Start(stopCh chan os.Signal) <-chan error {
	c.stats.mu.Lock()
	c.stats.statusOK = true
	c.stats.startedAt = time.Now()
	c.stats.mu.Unlock()

	connLossCh := make(chan struct{})
	c.actionFlow.NotifyOnConnectionLoss(connLossCh)

	c.actionFlow.Start()
	go c.processNewActions()
	go c.processUpdateActions()

	doneCh := make(chan error, 1)
	go func() {
		defer func() {
			c.stats.mu.Lock()
			c.stats.statusOK = false
			c.stats.finishedAt = time.Now()
			c.stats.mu.Unlock()
		}()
		
		for {
			select {
				case <-connLossCh:
					doneCh <- ErrConnectionLoss
					return
				case <-stopCh:
					_ = c.actionFlow.Stop()
					doneCh <- ErrInterrupted
					return
			}
		}
	}()

	return doneCh
}


func (c *Consumer) processNewActions() {
	h := func(na *model.NewAction) error {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if _, err := c.actionService.Create(ctx, na); err != nil {
			return err
		}

		return nil
	}

	c.actionFlow.ReceiveNewActions(c.consumerName, h)
}

func (c *Consumer) processUpdateActions() {
	h := func(ua *model.UpdateAction) error {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if _, err := c.actionService.Update(ctx, ua); err != nil {
			return err
		}

		return nil
	}

	c.actionFlow.ReceiveUpdateActions(c.consumerName, h)
}
