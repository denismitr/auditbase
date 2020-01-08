package queue

import "github.com/streadway/amqp"

type Message interface {
	Body() ([]byte, error)
	ContentType() string
}

type delivery struct {
	Exchange    string
	RoutingKey  string
	IsPeristent bool
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
