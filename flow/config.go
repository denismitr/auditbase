package flow

import (
	"os"
	"strconv"
)

// Config of the event exchange
type Config struct {
	ExchangeName      string
	RequeueRoutingKey string
	RoutingKey        string
	ErrorQueueName    string
	QueueName         string
	ExchangeType      string
	MaxRequeue        int
	IsPeristent       bool
}

// NewConfig of the event exchange
func NewConfig(exchangeName, exchangeType, routingKey, queue string, isPersistent bool) Config {
	return Config{
		ExchangeName: exchangeName,
		RoutingKey:   routingKey,
		QueueName:    queue,
		ExchangeType: exchangeType,
		IsPeristent:  isPersistent,
	}
}

func NewConfigFromGlobals() Config {
	maxRequeue, _ := strconv.Atoi(os.Getenv("EVENTS_MAX_REQUEUE"))

	return Config{
		ExchangeName:      os.Getenv("EVENTS_EXCHANGE"),
		RequeueRoutingKey: os.Getenv("EVENTS_REQUEUE_ROUTING_KEY"),
		RoutingKey:        os.Getenv("EVENTS_ROUTING_KEY"),
		QueueName:         os.Getenv("EVENTS_QUEUE_NAME"),
		ErrorQueueName:    os.Getenv("EVENTS_ERROR_QUEUE_NAME"),
		ExchangeType:      os.Getenv("EVENTS_EXCHANGE_TYPE"),
		MaxRequeue:        maxRequeue,
		IsPeristent:       true,
	}
}
