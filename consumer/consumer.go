package consumer

import (
	"context"
	"fmt"

	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/queue"
	"github.com/denismitr/auditbase/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type StopFunc func(ctx context.Context) error

type Consumer struct {
	logger        *logrus.Logger
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
	l *logrus.Logger,
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
			return errors.Wrapf(err, "target type with ID %s does not exist an can not be created", e.TargetType.ID)
		}

		e.TargetType = tt
		return nil
	}

	tt, err := c.targetTypes.FirstByName(e.TargetType.Name)
	if err != nil {
		c.logger.Error(err)

		tt = model.TargetType{
			ID:          utils.UUID4(),
			Name:        e.TargetType.Name,
			Description: "",
		}

		if err := c.targetTypes.Create(tt); err != nil {
			return errors.Wrapf(err, "tartget type with name %s does not exist and cannot be created", e.TargetType.Name)
		}
	}

	e.TargetType = tt
	return nil
}

func (c *Consumer) assignTargetServiceTo(e *model.Event) error {
	if e.TargetService.ID != "" {
		ts, err := c.microservices.GetOneByID(e.TargetService.ID)
		if err != nil {
			return errors.Wrapf(err, "microservice with ID %s does not exist and cannot be created", e.TargetService.ID)
		}

		e.TargetService = ts
		return nil
	}

	ts, err := c.microservices.GetOneByName(e.TargetService.Name)
	if err != nil {
		ts = model.Microservice{
			ID:          utils.UUID4(),
			Name:        e.TargetService.Name,
			Description: "",
		}

		if err := c.microservices.Create(ts); err != nil {
			return errors.Wrapf(err, "microservice with name %s does not exist and cannot be created", e.TargetService.Name)
		}
	}

	e.TargetService = ts
	return nil
}

func (c *Consumer) assignActorTypeTo(e *model.Event) error {
	if e.ActorType.ID != "" {
		at, err := c.actorTypes.FirstByID(e.ActorType.ID)
		if err != nil {
			return errors.Wrapf(err, "actor type with id %s does not exist and cannot be created", e.ActorType.ID)
		}

		e.ActorType = at
		return nil
	}

	at, err := c.actorTypes.FirstByName(e.ActorType.Name)
	if err != nil {
		at = model.ActorType{
			ID:          utils.UUID4(),
			Name:        e.ActorType.Name,
			Description: "",
		}

		if err := c.actorTypes.Create(at); err != nil {
			return errors.Wrapf(err, "actor type with name %s does not exist and could not be created", e.ActorType.Name)
		}
	}

	e.ActorType = at
	return nil
}

func (c *Consumer) assignActorServiceTo(e *model.Event) error {
	if e.ActorService.ID != "" {
		as, err := c.microservices.GetOneByID(e.ActorService.ID)
		if err != nil {
			return errors.Wrapf(err, "actor service with id %s does not exist and cannot be created", e.ActorService.ID)
		}

		e.ActorService = as
		return nil
	}

	as, err := c.microservices.GetOneByName(e.ActorService.Name)
	if err != nil {
		as = model.Microservice{
			ID:          utils.UUID4(),
			Name:        e.ActorService.Name,
			Description: "",
		}

		if err := c.microservices.Create(as); err != nil {
			return errors.Wrapf(err, "actor service with name %s does not exist and could not be created", e.ActorService.Name)
		}
	}

	e.ActorService = as
	return nil
}

func (c *Consumer) handleFailedEvent(re flow.ReceivedEvent, err error) {
	fmt.Println(err)
	c.logger.Error(err)
	re.Reject()
}
