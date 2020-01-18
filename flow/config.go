package flow

import "os"

// Config of the event exchange
type Config struct {
	ExchangeName string
	RoutingKey   string
	QueueName    string
	ExchangeType string
	IsPeristent  bool
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
	return Config {
		ExchangeName: os.Getenv("EVENTS_EXCHANGE"),
		RoutingKey:   os.Getenv("EVENTS_ROUTING_KEY"),
		QueueName:    os.Getenv("EVENTS_QUEUE_NAME"),
		ExchangeType: os.Getenv("EVENTS_EXCHANGE_TYPE"),
		IsPeristent:  true,
	}
}
