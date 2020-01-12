package flow

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
