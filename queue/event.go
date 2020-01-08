package queue

import (
	"github.com/pkg/errors"
)

type DirectEventExchange struct {
	mq MQ
	d  Delivery
}

type EventExchange interface {
	Publish(Message) error
	Consume() <-chan ReceivedMessage
}

// Publish event
func (ex *DirectEventExchange) Publish(msg Message) error {

	if err := ex.mq.Publish(msg, ex.d); err != nil {
		return err
	}

	return nil
}

// Consume messages
func (ex *DirectEventExchange) Consume() <-chan ReceivedMessage {
	go ex.mq.ListenOnQueue(ex.d.Queue)

	return ex.mq.Consume()
}

func NewDirectEventExchange(mq MQ, d Delivery) *DirectEventExchange {
	return &DirectEventExchange{
		mq: mq,
		d:  d,
	}
}

// Scaffold RabbitMQ exchange, queue and binding
func (ex *DirectEventExchange) Scaffold() error {
	if err := ex.mq.DeclareExchange(ex.d.Exchange, ex.d.ExchangeType); err != nil {
		return errors.Wrap(err, "could not scaffold DirectEventExchange on exchage declaration")
	}

	if err := ex.mq.DeclareQueue(ex.d.Queue); err != nil {
		return errors.Wrap(err, "could not scaffold DirectEventExchange on queue declaration")
	}

	if err := ex.mq.Bind(ex.d.Queue, ex.d.Exchange, ex.d.RoutingKey); err != nil {
		return errors.Wrap(err, "could not scaffold DirectEventExchange on queue binding")
	}

	return nil
}
