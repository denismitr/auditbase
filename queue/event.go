package queue

import (
	"encoding/json"
	"fmt"

	"github.com/denismitr/auditbase/model"
	"github.com/labstack/gommon/log"
	"github.com/pkg/errors"
)

type DirectEventExchange struct {
	MQ         MQ
	Exchange   string
	RoutingKey string
	QueueName  string
}

func (ex *DirectEventExchange) Publish(e model.Event) error {
	d := delivery{Exchange: ex.Exchange, RoutingKey: ex.RoutingKey}

	if err := ex.MQ.Publish(e, d); err != nil {
		return err
	}

	return nil
}

func (e *DirectEventExchange) Consume() <-chan model.Event {
	go e.MQ.ListenOnQueue(e.QueueName)

	ch := make(chan model.Event)

	go func() {
		defer log.Error("Whoops!")
		for b := range e.MQ.Consume() {
			evt := model.Event{}
			if err := json.Unmarshal(b, &evt); err != nil {
				log.Error(err)
				continue
			}
			fmt.Println("Received message from QUEUE")
			ch <- evt
			fmt.Println("Sent message to consumer")
		}
	}()

	return ch
}

func NewDirectEventExchange(
	mq MQ,
	exchange string,
	queueName string,
	routingKey string,
) *DirectEventExchange {
	return &DirectEventExchange{
		MQ:         mq,
		Exchange:   exchange,
		RoutingKey: routingKey,
		QueueName:  queueName,
	}
}

func Scaffold(s Scaffolder, exchange, queue, routingKey string) error {
	if err := s.DeclareExchange(exchange, "direct"); err != nil {
		return errors.Wrap(err, "could not scaffold DirectEventExchange on exchage declaration")
	}

	if err := s.DeclareQueue(queue); err != nil {
		return errors.Wrap(err, "could not scaffold DirectEventExchange on queue declaration")
	}

	if err := s.Bind(queue, exchange, routingKey); err != nil {
		return errors.Wrap(err, "could not scaffold DirectEventExchange on queue binding")
	}

	return nil
}
