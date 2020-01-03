package consumer

import (
	"fmt"
	"sync"

	"github.com/denismitr/auditbase/model"
	"github.com/sirupsen/logrus"
)

type Consumer struct {
	logger        *logrus.Logger
	exchange      model.EventExchange
	microservices model.MicroserviceRepository
	events        model.EventRepository
}

func New(
	l *logrus.Logger,
	ee model.EventExchange,
	ms model.MicroserviceRepository,
	evt model.EventRepository,
) *Consumer {
	return &Consumer{
		logger:        l,
		exchange:      ee,
		microservices: ms,
		events:        evt,
	}
}

func (c *Consumer) Start() error {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		for e := range c.exchange.Consume() {
			go c.processEvent(e)
		}
		fmt.Println("Something went very wrong!!!! I'm done!!!!!!!!!!!!!")
		wg.Done()
	}()
	wg.Wait()

	return nil
}

func (c *Consumer) processEvent(e model.Event) {
	if err := c.events.Create(e); err != nil {
		fmt.Println("Something went very wrong!!!!")
		c.logger.Error(err)
	}
}
