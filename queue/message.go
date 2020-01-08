package queue

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type Message interface {
	Body() []byte
	ContentType() string
}

type JSONMessage struct {
	body []byte
}

func (m *JSONMessage) Body() []byte {
	return m.body
}

func (m *JSONMessage) ContentType() string {
	return "application/json"
}

func NewJSONMessage(v interface{}) (*JSONMessage, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, errors.Wrap(err, "could not create JSON queue message")
	}

	return &JSONMessage{b}, nil
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
