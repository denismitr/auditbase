package consumer

import (
	"fmt"
	"sync"

	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils"
	"github.com/sirupsen/logrus"
)

type Consumer struct {
	logger        *logrus.Logger
	exchange      model.EventExchange
	microservices model.MicroserviceRepository
	events        model.EventRepository
	targetTypes   model.TargetTypeRepository
	actorTypes    model.ActorTypeRepository
}

func New(
	l *logrus.Logger,
	ee model.EventExchange,
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

func (c *Consumer) processEvent(e model.Event) {
	if e.TargetType.ID == "" {

		// Refactor to FirstOrCreateByName
		tt, err := c.targetTypes.FirstByName(e.TargetType.Name)
		if err != nil {
			fmt.Println(err)
			c.logger.Error(err)

			tt = model.TargetType{
				ID:          utils.UUID4(),
				Name:        e.TargetType.Name,
				Description: "",
			}

			if err := c.targetTypes.Create(tt); err != nil {
				fmt.Println(err)
				c.logger.Error(err)
				return
			}
		}

		e.TargetType = tt
	}

	if e.ActorType.ID == "" {
		// Refactor to FirstOrCreateByName
		at, err := c.actorTypes.FirstByName(e.ActorType.Name)
		if err != nil {
			fmt.Println(err)
			c.logger.Error(err)

			at = model.ActorType{
				ID:          utils.UUID4(),
				Name:        e.ActorType.Name,
				Description: "",
			}

			if err := c.actorTypes.Create(at); err != nil {
				fmt.Println(err)
				c.logger.Error(err)
				return
			}
		}

		e.ActorType = at
	}

	if err := c.events.Create(e); err != nil {
		fmt.Println("Something went very wrong!!!!")
		c.logger.Error(err)
	}
}
