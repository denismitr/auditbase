package queue

import (
	"fmt"
	"log"
	"sync"
	"time"

	"encoding/json"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

const (
	AuditLogMessages = "audit_log_messages"
)

// Scaffolder - scaffolds the Message Queue,
// getting it ready for work
type Scaffolder interface {
	DeclareExchange(name, kind string) error
	DeclareQueue(name string) error
	Bind(queue, exhange, routingKey string) error
}

// MQ is the message queue
type MQ interface {
	Scaffolder

	Publish(interface{}, delivery) error
	OpenAndKeepConnection() error
	ListenOnQueue(name string)
	Consume() <-chan []byte
	Stop()
}

// RabbitQueue handles message queue
type RabbitQueue struct {
	dsn       string
	conn      *amqp.Connection
	logger    *logrus.Logger
	stopCh    chan struct{}
	errorCh   chan error
	receiveCh chan []byte

	maxConnRetries int

	mu *sync.RWMutex
}

func NewRabbitQueue(dsn string, logger *logrus.Logger, maxConnRetries int) *RabbitQueue {
	return &RabbitQueue{
		dsn:            dsn,
		conn:           nil,
		logger:         logger,
		stopCh:         make(chan struct{}),
		receiveCh:      make(chan []byte),
		maxConnRetries: maxConnRetries,
		mu:             &sync.RWMutex{},
	}
}

func (q *RabbitQueue) Stop() {

}

func (q *RabbitQueue) Publish(msg interface{}, d delivery) error {
	ch, err := q.conn.Channel()
	defer ch.Close()

	if err != nil {
		return errors.Wrapf(err, "could not publish message to %s with routing key %s", d.Exchange, d.RoutingKey)
	}

	b, err := json.Marshal(msg)
	if err != nil {
		return errors.Wrapf(err, "failed to convert message to JSON before sending to RabbitMQ")
	}

	p := amqp.Publishing{
		ContentType: "application/json",
		Body:        b,
	}

	if err := ch.Publish(d.Exchange, d.RoutingKey, false, false, p); err != nil {
		return errors.Wrapf(err, "failed to send message to exchange %s with routing key %s", d.Exchange, d.RoutingKey)
	}

	return nil
}

// DeclareExchange - declares RabbitMQ exchange
func (q *RabbitQueue) DeclareExchange(name, kind string) error {
	ch, err := q.conn.Channel()
	defer ch.Close()

	if err != nil {
		return errors.Wrap(err, "failed to get a channel from connection")
	}

	if err := ch.ExchangeDeclare(name, kind, true, false, false, false, nil); err != nil {
		return errors.Wrapf(err, "failed to declare exchange %s of kind %s", name, kind)
	}

	return nil
}

// DeclareQueue - declares a new queue if not exists
func (q *RabbitQueue) DeclareQueue(name string) error {
	ch, err := q.conn.Channel()
	defer ch.Close()

	if err != nil {
		return errors.Wrap(err, "failed to get a channel from connection")
	}

	if _, err := ch.QueueDeclare(name, true, false, false, false, nil); err != nil {
		return errors.Wrapf(err, "failed to declare queue %s", name)
	}

	return nil
}

// Bind queue to exchange with routingKey
func (q *RabbitQueue) Bind(queue, exhange, routingKey string) error {
	ch, err := q.conn.Channel()
	defer ch.Close()

	if err != nil {
		return errors.Wrap(err, "failed to get a channel from connection")
	}

	if err := ch.QueueBind(queue, routingKey, exhange, false, nil); err != nil {
		return errors.Wrapf(
			err,
			"failed to bind queue %s to exchange %s with routing key %s",
			queue,
			routingKey,
			exhange,
		)
	}

	return nil
}

func (q *RabbitQueue) OpenAndKeepConnection() error {
	// TODO: implement reconnection - maybe
	return nil
}

func (q *RabbitQueue) ListenOnQueue(name string) {
	ch, err := q.conn.Channel()
	defer ch.Close()
	if err != nil {
		panic(errors.Wrapf(err, "could not get channel for listening queue %s", name))
	}

	msgs, err := ch.Consume(
		name,  // queue
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)

	if err != nil {
		panic(err)
	}

	fmt.Println("Waiting for messages")

	for msg := range msgs {
		q.receiveCh <- msg.Body
	}
}

func (q *RabbitQueue) Consume() <-chan []byte {
	return q.receiveCh
}

func (q *RabbitQueue) WaitForConnection() {
	attempt := 1

	for attempt <= q.maxConnRetries {
		log.Printf("Waiting for RabbitMQ: attempt %d", attempt)

		conn, err := amqp.Dial(q.dsn)
		if err != nil {
			log.Printf("\nattempt %d failed: %s", attempt, err)
			attempt++
			time.Sleep(5 * time.Second)
			continue
		}

		q.conn = conn
		return
	}

	log.Fatal("Failed to connect to Rabbit: too many attempts")
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
