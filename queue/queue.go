package queue

import (
	"log"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

const (
	AuditLogMessages = "audit_log_messages"
)

// MQ is the message queu
type MQ interface {
	Produce(Message, Delivery) error
	Declare(name string) error
	OpenAndKeepConnection() error
	ListenOnQueue(name string)
	Consumer() <-chan ReceivedMessage
	Stop()
}

// RabbitQueue handles message queue
type RabbitQueue struct {
	dsn       string
	conn      *amqp.Connection
	logger    *logrus.Logger
	stopCh    chan struct{}
	errorCh   chan error
	receiveCh chan ReceivedMessage

	maxConnRetries int

	mu *sync.RWMutex
}

func NewRabbitQueue(dsn string, logger *logrus.Logger, maxConnRetries int) *RabbitQueue {
	return &RabbitQueue{
		dsn:            dsn,
		conn:           nil,
		logger:         logger,
		stopCh:         make(chan struct{}),
		receiveCh:      make(chan ReceivedMessage),
		maxConnRetries: maxConnRetries,
		mu:             &sync.RWMutex{},
	}
}

func (q *RabbitQueue) Stop() {

}

func (q *RabbitQueue) Produce(m Message, d Delivery) error {
	return nil
}

func (q *RabbitQueue) Declare(name string) error {
	return nil
}

func (q *RabbitQueue) OpenAndKeepConnection() error {
	return nil
}

func (q *RabbitQueue) ListenOnQueue(name string) {

}

func (q *RabbitQueue) Consumer() <-chan ReceivedMessage {
	return nil
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
