package queue

import (
	"github.com/streadway/amqp"
)

type Message interface {
	Body() []byte
	ContentType() string
}

type JSONMessage struct {
	body []byte
}

func NewJSONMessage(b []byte) *JSONMessage {
	return &JSONMessage{
		body: b,
	}
}

func (e *JSONMessage) Body() []byte {
	return e.body
}

func (e *JSONMessage) ContentType() string {
	return "application/json"
}

type ReceivedMessage interface {
	Body() []byte
	Queue() string
	Ack() error
	Reject(requeue bool) error
}

type RabbitMQReceivedMessage struct {
	queueName string
	msg       amqp.Delivery
}

func (m *RabbitMQReceivedMessage) Queue() string {
	return m.queueName
}

func (m *RabbitMQReceivedMessage) Body() []byte {
	return m.msg.Body
}

func (m *RabbitMQReceivedMessage) Ack() error {
	return m.msg.Ack(false)
}

func (m *RabbitMQReceivedMessage) Reject(requeue bool) error {
	return m.msg.Reject(requeue)
}

func newRabbitMQReceivedMessage(queueName string, msg amqp.Delivery) *RabbitMQReceivedMessage {
	return &RabbitMQReceivedMessage{
		queueName: queueName,
		msg:       msg,
	}
}
