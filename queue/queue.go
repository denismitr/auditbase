package queue

import (
	"fmt"
	"sync"
	"time"

	"github.com/denismitr/auditbase/utils"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

const (
	AuditLogMessages = "audit_log_messages"
)

type ConnectionStatus int

const (
	Connected = iota
	Connecting
	ConnectionDropped
	ConnectionClosed
)

// Scaffolder - scaffolds the Message Queue,
// getting it ready for work
type Scaffolder interface {
	DeclareExchange(name, kind string) error
	DeclareQueue(name string) error
	Bind(queue, exchange, routingKey string) error
}

// MQ is the message queue
type MQ interface {
	Scaffolder

	Inspect(queueName string) (Inspection, error)
	Publish(msg Message, exchange, routingKey string) error
	Subscribe(queue, consumer string, receiveCh chan<- ReceivedMessage)
	Connect() error
	Maintain()
	Status() ConnectionStatus
	NotifyStatusChange(listener chan ConnectionStatus)
	Stop()
}

// RabbitQueue handles message queue
type RabbitQueue struct {
	dsn             string
	conn            *amqp.Connection
	logger          utils.Logger
	stopCh          chan struct{}
	errorCh         chan error
	connErrCh       chan *amqp.Error
	statusListeners []chan ConnectionStatus
	status          ConnectionStatus

	maxConnRetries int

	mu sync.RWMutex
}

// NewRabbitQueue - creates a new message queue with RabbitMQ implementation
func NewRabbitQueue(dsn string, logger utils.Logger, maxConnRetries int) *RabbitQueue {
	return &RabbitQueue{
		dsn:             dsn,
		conn:            nil,
		logger:          logger,
		stopCh:          make(chan struct{}),
		statusListeners: make([]chan ConnectionStatus, 0),
		maxConnRetries:  maxConnRetries,
		mu:              sync.RWMutex{},
	}
}

// Stop the MessageQueue
func (q *RabbitQueue) Stop() {
	close(q.stopCh)
}

// Status - returns current connection status
func (q *RabbitQueue) Status() ConnectionStatus {
	q.mu.RLock()
	defer q.mu.Unlock()

	return q.status
}

// NotifyStatusChange - registers a listener for on connection status change
func (q *RabbitQueue) NotifyStatusChange(listener chan ConnectionStatus) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.statusListeners = append(q.statusListeners, listener)
}

// Publish message to message queue
func (q *RabbitQueue) Publish(msg Message, exchange, routingKey string) error {
	ch, err := q.conn.Channel()
	defer ch.Close()

	if err != nil {
		return errors.Wrapf(
			err, "could not publish message to %s with routing key %s", exchange, routingKey)
	}

	p := amqp.Publishing{
		ContentType: msg.ContentType(),
		Body:        msg.Body(),
	}

	if err := ch.Publish(exchange, routingKey, false, false, p); err != nil {
		return errors.Wrapf(
			err, "failed to send message to exchange %s with routing key %s", exchange, routingKey)
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
func (q *RabbitQueue) Bind(queue, exchange, routingKey string) error {
	ch, err := q.conn.Channel()

	if err != nil {
		return errors.Wrap(err, "failed to get a channel from connection")
	}

	defer ch.Close()

	if err := ch.QueueBind(queue, routingKey, exchange, false, nil); err != nil {
		return errors.Wrapf(
			err,
			"failed to bind queue %s to exchange %s with routing key %s",
			queue,
			routingKey,
			exchange,
		)
	}

	return nil
}

// Inspect queue, check number of messages waiting to be consumed
// and a number of consumers for that given queue
func (q *RabbitQueue) Inspect(queueName string) (Inspection, error) {
	i := Inspection{}

	ch, err := q.conn.Channel()
	if err != nil {
		return i, errors.Wrapf(err, "failed to get a channel to inspect a queue %s", queueName)
	}

	defer ch.Close()

	queue, err := ch.QueueInspect(queueName)
	if err != nil {
		return i, errors.Wrapf(err, "could not inspect a queue %s", queueName)
	}

	i.Messages = queue.Messages
	i.Consumers = queue.Consumers

	return i, nil
}

// Subscribe and consume messages sending them
// to receiveCh
func (q *RabbitQueue) Subscribe(queue, consumer string, receiveCh chan<- ReceivedMessage) {
	ch, err := q.conn.Channel()
	defer ch.Close()

	if err != nil {
		panic(errors.Wrapf(err, "could not get channel for listening queue %s", queue))
	}

	msgs, err := ch.Consume(
		queue,    // queue
		consumer, // consumer
		false,    // auto-ack
		false,    // exclusive
		false,    // no-local
		false,    // no-wait
		nil,      // args
	)

	if err != nil {
		panic(err)
	}

	fmt.Println("Waiting for messages...")

	for {
		select {
		case msg := <-msgs:
			receiveCh <- newRabbitMQReceivedMessage(queue, msg)
		case <-q.stopCh:
			close(receiveCh)
		}
	}
}

// Connect waits for RabbitMQ to start up
// and makes attempts to connect to irt
// this function is not, and not supposed to be thread safe
// only one goroutine should run it at a time
func (q *RabbitQueue) Connect() error {
	attempt := 1

	q.updateStatus(Connecting)

	for attempt <= q.maxConnRetries {
		q.logger.Debugf("Waiting for RabbitMQ on %s: attempt %d", q.dsn, attempt)

		conn, err := amqp.Dial(q.dsn)
		if err != nil {
			q.logger.Error(
				errors.Wrapf(err, "attempt %d failed", attempt),
			)

			attempt++
			time.Sleep(5 * time.Second * time.Duration(attempt))
			continue
		}

		q.conn = conn
		q.updateStatus(Connected)
		q.connErrCh = make(chan *amqp.Error)
		q.conn.NotifyClose(q.connErrCh)

		q.logger.Debugf("Established connection with %s", q.dsn)

		return nil
	}

	return errors.Errorf("failed to connect to rabbitMQ on %s - too many attempts", q.dsn)
}

// Maintain connection to RabbitMQ and listen to close channel
// should be run as a goroutine
func (q *RabbitQueue) Maintain() {
	var connErr *amqp.Error

	for {
		select {
		case connErr = <-q.connErrCh:
			if connErr != nil {
				q.updateStatus(ConnectionDropped)
				q.logger.Error(errors.Errorf("RabbitMQ connection error: %s", connErr.Error()))
				if err := q.Connect(); err != nil {
					panic(errors.Wrap(err, "failed to reconnect to RabbitMQ"))
				}
				continue
			} else {
				// connection was deliberately closed
				q.updateStatus(ConnectionClosed)
				return
			}
		case <-q.stopCh:
			if !q.conn.IsClosed() {
				q.conn.Close()
			}

			q.updateStatus(ConnectionClosed)
		}
	}
}

func (q *RabbitQueue) updateStatus(s ConnectionStatus) {
	q.mu.Lock()
	q.status = s
	q.mu.Unlock()

	q.mu.RLock()
	defer q.mu.RUnlock()

	for _, l := range q.statusListeners {
		if s == ConnectionClosed {
			close(l)
		} else {
			l <- s
		}
	}
}
