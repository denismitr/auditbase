package consumer

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/queue"
	"github.com/denismitr/auditbase/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Consumer struct {
	logger        *logrus.Logger
	exchange      queue.EventExchange
	microservices model.MicroserviceRepository
	events        model.EventRepository
	targetTypes   model.TargetTypeRepository
	actorTypes    model.ActorTypeRepository
}

func New(
	l *logrus.Logger,
	ee queue.EventExchange,
	ms model.MicroserviceRepository,
	evt model.EventRepository,
	tts model.TargetTypeRepository,
	ats model.ActorTypeRepository,
) *Consumer {
	return &Consumer{
		logger:        l,
		exchange:      ee,
		microservices: ms,
		events:        evt,
		targetTypes:   tts,
		actorTypes:    ats,
	}
}

// Start consumer
func (c *Consumer) Start() error {
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		for e := range c.exchange.Consume() {
			go c.processEvent(e)
		}
		wg.Done()
	}()

	wg.Wait()

	return nil
}

func (c *Consumer) processEvent(msg queue.ReceivedMessage) {
	e := model.Event{}

	if err := json.Unmarshal(msg.Body(), &e); err != nil {
		c.handleFailedMsg(msg, err)
		return
	}

	if err := c.assignActorTypeTo(&e); err != nil {
		c.handleFailedMsg(msg, err)
		return
	}

	if err := c.assignActorServiceTo(&e); err != nil {
		c.handleFailedMsg(msg, err)
		return
	}

	if err := c.assignTargetTypeTo(&e); err != nil {
		c.handleFailedMsg(msg, err)
		return
	}

	if err := c.assignTargetServiceTo(&e); err != nil {
		c.handleFailedMsg(msg, err)
		return
	}

	if err := c.events.Create(e); err != nil {
		c.handleFailedMsg(msg, err)
		return
	}

	msg.Ack()
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

func (c *Consumer) handleFailedMsg(msg queue.ReceivedMessage, err error) {
	fmt.Println(err)
	c.logger.Error(err)
	msg.Reject(false)
}
