package queue

import (
	"context"
	"fmt"
	"github.com/denismitr/auditbase/internal/utils/retry"
	"sync"
	"time"

	"github.com/denismitr/auditbase/internal/utils/logger"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type ConnectionStatus int

const (
	Idle = iota
	Connected
	Connecting
	ConnectionDropped
	ConnectionClosed
	ConnectionFailed
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

	// fixme: refactor following 4 methods to some sort of PubSub interface
	Publish(msg Message, exchange, routingKey string) error
	Reject(tag uint64) error
	Ack(tag uint64) error
	Subscribe(queue, consumer string, receiveCh chan<- ReceivedMessage) error

	Connect(ctx context.Context) error
	Maintain()
	Status() ConnectionStatus
	NotifyStatusChange(listener chan ConnectionStatus)
	Stop()
}

// RabbitQueue handles message queue
type RabbitQueue struct {
	dsn             string
	conn            *amqp.Connection
	channel         *amqp.Channel
	logger          logger.Logger
	stopCh          chan struct{}
	connErrCh       chan *amqp.Error
	statusListeners []chan ConnectionStatus
	status          ConnectionStatus

	maxConnRetries int

	mu sync.RWMutex
}

// Rabbit - creates a new message queue with RabbitMQ implementation
func Rabbit(dsn string, logger logger.Logger, maxConnRetries int) *RabbitQueue {
	return &RabbitQueue{
		dsn:             dsn,
		conn:            nil,
		channel:         nil,
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
	defer q.mu.RUnlock()

	return q.status
}

// NotifyStatusChange - registers a listener for on connection status change
func (q *RabbitQueue) NotifyStatusChange(listener chan ConnectionStatus) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.statusListeners = append(q.statusListeners, listener)
}

func (q *RabbitQueue) Reject(tag uint64) error {
	if err := q.channel.Reject(tag, false); err != nil {
		return errors.Wrapf(err, "could not reject tag %d", tag)
	}

	return nil
}

func (q *RabbitQueue) Ack(tag uint64) error {
	if err := q.channel.Ack(tag, false); err != nil {
		return errors.Wrapf(err, "could not ack tag %d", tag)
	}

	return nil
}

// Publish message to message queue
func (q *RabbitQueue) Publish(msg Message, exchange, routingKey string) error {
	p := amqp.Publishing{
		ContentType: msg.ContentType(),
		Body:        msg.Body(),
		Headers:     amqp.Table{"Attempt": msg.Attempt()},
	}

	if msg.Attempt() != 1 {
		q.logger.Debugf("Requing an errored message attempt %d", msg.Attempt())
	}

	if err := q.channel.Publish(exchange, routingKey, false, false, p); err != nil {
		return errors.Wrapf(
			err, "failed to send message to exchange %s with routing key %s", exchange, routingKey)
	}

	return nil
}

// DeclareExchange - declares RabbitMQ exchange
func (q *RabbitQueue) DeclareExchange(name, kind string) error {
	if err := q.channel.ExchangeDeclare(name, kind, true, false, false, false, nil); err != nil {
		return errors.Wrapf(err, "failed to declare exchange %s of kind %s", name, kind)
	}

	return nil
}

// DeclareQueue - declares a new queue if not exists
func (q *RabbitQueue) DeclareQueue(name string) error {
	if _, err := q.channel.QueueDeclare(name, true, false, false, false, nil); err != nil {
		return errors.Wrapf(err, "failed to declare queue %s", name)
	}

	return nil
}

// Bind queue to exchange with routingKey
func (q *RabbitQueue) Bind(queue, exchange, routingKey string) error {
	if err := q.channel.QueueBind(queue, routingKey, exchange, false, nil); err != nil {
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

	queue, err := q.channel.QueueInspect(queueName)
	if err != nil {
		return i, errors.Wrapf(err, "could not inspect a queue %s", queueName)
	}

	i.Messages = queue.Messages
	i.Consumers = queue.Consumers

	return i, nil
}

// Subscribe and consume messages sending them
// to receiveCh
func (q *RabbitQueue) Subscribe(queue, consumer string, receiveCh chan<- ReceivedMessage) error {
	msgs, err := q.channel.Consume(
		queue,    // queue
		consumer, // consumer
		false,    // auto-ack
		false,    // exclusive
		false,    // no-local
		false,    // no-wait
		nil,      // args
	)

	if err != nil {
		return errors.Wrapf(err, "could not consume from queue %s", queue)
	}

	fmt.Println("Waiting for messages...")

	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				// probably must reconnect
				close(receiveCh)
				return nil
			}

			rMsg, err := newRabbitMQReceivedMessage(queue, msg)
			if err != nil {
				q.logger.Error(err)
				continue
			}

			q.logger.Debugf("consumer %s received message from queue %s on its %d attempt", consumer, queue, rMsg.Attempt())

			receiveCh <- rMsg
		case <-q.stopCh:
			close(receiveCh)
			return nil
		}
	}
}

// Connect waits for RabbitMQ to start up
// and makes attempts to connect to irt
// this function is not, and not supposed to be thread safe
// only one goroutine should run it at a time
func (q *RabbitQueue) Connect(ctx context.Context) error {
	q.updateStatus(Connecting)

	// max retries are not very important
	// given that context is usually responsible for timeout
	// just in case
	maxConnRetries := 300

	if err := retry.Incremental(ctx, 1 * time.Second, maxConnRetries, func(attempt int) (err error) {
		q.logger.Debugf("Waiting for RabbitMQ on %s: attempt %d", q.dsn, attempt)

		conn, err := amqp.Dial(q.dsn)
		if err != nil {
			q.logger.Error(
				errors.Wrapf(err, "attempt %d failed", attempt),
			)
			return retry.Error(err, attempt)
		}

		ch, err := conn.Channel()
		if err != nil {
			return errors.Wrap(err, "Queue connection failed: could not open AMQP channel")
		}

		q.conn = conn
		q.channel = ch
		q.updateStatus(Connected)
		q.connErrCh = make(chan *amqp.Error)
		q.conn.NotifyClose(q.connErrCh)

		return nil
	}); err != nil {
		q.updateStatus(ConnectionFailed)
		return errors.Wrapf(err, "failed to connect to rabbitMQ on %s", q.dsn)
	}

	q.logger.Debugf("Established connection with %s", q.dsn)

	return nil
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
				ctx, cancel := context.WithTimeout(context.Background(), 60 *time.Second)
				if err := q.Connect(ctx); err != nil {
					cancel()
					panic(errors.Wrap(err, "failed to reconnect to RabbitMQ"))
				}
				cancel()
				continue
			} else {
				// connection was deliberately closed
				q.updateStatus(ConnectionClosed)
				return
			}
		case <-q.stopCh:
			if q.channel != nil {
				_ = q.channel.Close()
			}

			if !q.conn.IsClosed() {
				_ = q.conn.Close()
			}

			q.updateStatus(ConnectionClosed)
			q.closeStatusListeners()
		}
	}
}

func (q *RabbitQueue) updateStatus(s ConnectionStatus) {
	q.mu.Lock()
	q.status = s
	defer q.mu.Unlock()
	q.logger.Debugf("QUEUE STATUS CHANGED To %#v", s)
	for _, l := range q.statusListeners {
		l <- s
	}
}

func (q *RabbitQueue) closeStatusListeners() {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, l := range q.statusListeners {
		close(l)
	}
}
