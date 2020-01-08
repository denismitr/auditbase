package queue

import (
	"github.com/denismitr/auditbase/model"
	"github.com/pkg/errors"
)

type DirectEventExchange struct {
	MQ         MQ
	Exchange   string
	RoutingKey string
	QueueName  string
}

type EventExchange interface {
	Publish(model.Event) error
	Consume() <-chan ReceivedMessage
}

// Publish event
func (ex *DirectEventExchange) Publish(e model.Event) error {
	d := delivery{Exchange: ex.Exchange, RoutingKey: ex.RoutingKey}

	if err := ex.MQ.Publish(e, d); err != nil {
		return err
	}

	return nil
}

// Consume messages
func (ex *DirectEventExchange) Consume() <-chan ReceivedMessage {
	go ex.MQ.ListenOnQueue(ex.QueueName)

	return ex.MQ.Consume()
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

// Scaffold RabbitMQ exchange, queue and binding
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
