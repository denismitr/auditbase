package consumer

import (
	"context"

	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/queue"
	"github.com/denismitr/auditbase/utils"
	"github.com/pkg/errors"
)

type StopFunc func(ctx context.Context) error

type Consumer struct {
	logger        utils.Logger
	f             flow.EventFlow
	microservices model.MicroserviceRepository
	events        model.EventRepository
	targetTypes   model.TargetTypeRepository
	actorTypes    model.ActorTypeRepository

	receiveCh chan queue.ReceivedMessage
	stopCh    chan struct{}
}

func New(
	f flow.EventFlow,
	l utils.Logger,
	mq queue.MQ,
	ms model.MicroserviceRepository,
	evt model.EventRepository,
	tts model.TargetTypeRepository,
	ats model.ActorTypeRepository,
) *Consumer {
	return &Consumer{
		f:             f,
		logger:        l,
		microservices: ms,
		events:        evt,
		targetTypes:   tts,
		actorTypes:    ats,
		receiveCh:     make(chan queue.ReceivedMessage),
		stopCh:        make(chan struct{}),
	}
}

// Start consumer
func (c *Consumer) Start(consumerName string) StopFunc {
	events := c.f.Receive(consumerName)

	go func() {
		for {
			select {
			case e := <-events:
				go c.processEvent(e)
			case <-c.stopCh:
				c.f.Stop()
				return
			}
		}
	}()

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

func (c *Consumer) processEvent(re flow.ReceivedEvent) {
	if re == nil {
		return
	}

	e, err := re.Event()
	if err != nil {
		c.handleFailedEvent(re, err)
		return
	}

	if err := c.assignActorTypeTo(&e); err != nil {
		c.handleFailedEvent(re, err)
		return
	}

	if err := c.assignActorServiceTo(&e); err != nil {
		c.handleFailedEvent(re, err)
		return
	}

	if err := c.assignTargetTypeTo(&e); err != nil {
		c.handleFailedEvent(re, err)
		return
	}

	if err := c.assignTargetServiceTo(&e); err != nil {
		c.handleFailedEvent(re, err)
		return
	}

	if err := c.events.Create(e); err != nil {
		c.handleFailedEvent(re, err)
		return
	}

	re.Ack()
}

func (c *Consumer) assignTargetTypeTo(e *model.Event) error {
	if e.TargetType.ID != "" {
		tt, err := c.targetTypes.FirstByID(e.TargetType.ID)
		if err != nil {
			return errors.Wrapf(err, "target type with ID %s does not exist", e.TargetType.ID)
		}

		e.TargetType = tt
		return nil
	}

	tt, err := c.targetTypes.FirstOrCreateByName(e.TargetType.Name)
	if err != nil {
		return err
	}

	e.TargetType = tt
	return nil
}

func (c *Consumer) assignTargetServiceTo(e *model.Event) error {
	if e.TargetService.ID != "" {
		ts, err := c.microservices.FirstByID(e.TargetService.ID)
		if err != nil {
			return errors.Wrapf(err, "target type ID %s does not exist", e.TargetService.ID)
		}

		e.TargetService = ts
		return nil
	}

	ts, err := c.microservices.FirstOrCreateByName(e.TargetService.Name)
	if err != nil {
		return err
	}

	e.TargetService = ts

	return nil
}

func (c *Consumer) assignActorTypeTo(e *model.Event) error {
	if e.ActorType.ID != "" {
		at, err := c.actorTypes.FirstByID(e.ActorType.ID)
		if err != nil {
			return errors.Wrapf(err, "actor type with id %s does not exist", e.ActorType.ID)
		}

		e.ActorType = at
		return nil
	}

	at, err := c.actorTypes.FirstOrCreateByName(e.ActorType.Name)
	if err != nil {
		return nil
	}

	e.ActorType = at
	return nil
}

func (c *Consumer) assignActorServiceTo(e *model.Event) error {
	if e.ActorService.ID != "" {
		as, err := c.microservices.FirstByID(e.ActorService.ID)
		if err != nil {
			return errors.Wrapf(err, "actor service with id %s does not exist", e.ActorService.ID)
		}

		e.ActorService = as
		return nil
	}

	as, err := c.microservices.FirstOrCreateByName(e.ActorService.Name)
	if err != nil {
		return err
	}

	e.ActorService = as
	return nil
}

func (c *Consumer) handleFailedEvent(re flow.ReceivedEvent, err error) {
	c.logger.Error(err)
	re.Reject()
}
