package queue

type Delivery struct {
	Queue        string
	Exchange     string
	RoutingKey   string
	ExchangeType string
	IsPeristent  bool
}

func NewDelivery(queue, exchange, routingKey, exchangeType string, isPersistent bool) Delivery {
	return Delivery{
		Queue:        queue,
		Exchange:     exchange,
		RoutingKey:   routingKey,
		ExchangeType: exchangeType,
		IsPeristent:  isPersistent,
	}
}
