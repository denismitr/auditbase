package queue

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type Message interface {
	Body() []byte
	ContentType() string
	Attempt() int
}

type JSONMessage struct {
	body       []byte
	attempt    int
}

func NewJSONMessage(b []byte, attempt int) *JSONMessage {
	return &JSONMessage{
		body:       b,
		attempt:    attempt,
	}
}

func (e *JSONMessage) Body() []byte {
	return e.body
}

func (e *JSONMessage) ContentType() string {
	return "application/json"
}

func (e *JSONMessage) Attempt() int {
	return e.attempt
}

type ReceivedMessageID uint64

func (rmID ReceivedMessageID) UInt64() uint64 {
	return uint64(rmID)
}

type ReceivedMessage interface {
	Body() []byte
	Channel() string
	Attempt() int
	CloneToRequeue() Message
	ID() ReceivedMessageID
}

type RabbitMQReceivedMessage struct {
	queueName  string
	body       []byte
	attempt    int
	tag        ReceivedMessageID
}

func (m RabbitMQReceivedMessage) Channel() string {
	return m.queueName
}

func (m RabbitMQReceivedMessage) Body() []byte {
	return m.body
}

func (m RabbitMQReceivedMessage) Attempt() int {
	return m.attempt
}

func (m RabbitMQReceivedMessage) ID() ReceivedMessageID {
	return m.tag
}

func (m *RabbitMQReceivedMessage) CloneToRequeue() Message {
	b := make([]byte, len(m.body))
	copy(b, m.body)
	return NewJSONMessage(b, m.Attempt()+1)
}

func newRabbitMQReceivedMessage(queueName string, msg amqp.Delivery) (*RabbitMQReceivedMessage, error) {
	attempt, err := extractAttemptFromHeader(msg.Headers)
	if err != nil {
		return nil, err
	}

	return &RabbitMQReceivedMessage{
		queueName:  queueName,
		body:       msg.Body,
		tag:        ReceivedMessageID(msg.DeliveryTag),
		attempt:    attempt,
	}, nil
}

func extractAttemptFromHeader(h amqp.Table) (int, error) {
	v, ok := h["Attempt"]
	if !ok {
		return 0, ErrNoAttemptInfo
	}

	attempt, err := strconv.Atoi(fmt.Sprintf("%d", v))
	if err != nil {
		return 0, errors.Wrap(ErrMalformedAttemptInfo, err.Error())
	}

	return attempt, nil
}
