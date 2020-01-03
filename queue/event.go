package queue

import (
	"github.com/denismitr/auditbase/model"
)

type DirectEventExchange struct {
	MQ         MQ
	Exchange   string
	RoutingKey string
}

func (ex *DirectEventExchange) Publish(e model.Event) error {
	d := delivery{Exchange: ex.Exchange, RoutingKey: ex.RoutingKey}

	if err := ex.MQ.Publish(e, d); err != nil {
		return err
	}

	return nil
}

func (e *DirectEventExchange) Consume() <-chan model.Event {
	return make(chan model.Event)
}
