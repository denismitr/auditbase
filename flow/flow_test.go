package flow

import (
	"errors"
	"testing"

	"github.com/denismitr/auditbase/queue"
	"github.com/stretchr/testify/assert"
)

func TestFlow(t *testing.T) {
	assert.Equal(t, true, true)
}

type declareExchangeExpectations func(exchangeName, exchangeType string)

type fakeMQ struct {
	cfg                         Config
	shouldFailOnDeclareExchange bool
	shouldFailOnDeclareQueue    bool
	shouldFailOnBind            bool
	t                           *testing.T
}

func (f fakeMQ) DeclareExchange(exchangeName, exchangeType string) error {
	if f.shouldFailOnDeclareExchange {
		return errors.New("failed on declare exchange")
	}

	assert.Equal(f.t, f.cfg.ExchangeName, exchangeName)
	assert.Equal(f.t, f.cfg.ExchangeType, exchangeType)

	return nil
}

func (f fakeMQ) DeclareQueue(queueName string) error {
	if f.shouldFailOnDeclareQueue {
		return errors.New("failed on queue declare")
	}

	assert.Equal(f.t, f.cfg.QueueName, queueName)

	return nil
}

func (f fakeMQ) Bind(queue, exchange, routingKey string) error {
	if f.shouldFailOnBind {
		return errors.New("failed on queu bind")
	}

	assert.Equal(f.t, f.cfg.QueueName, queue)
	assert.Equal(f.t, f.cfg.ExchangeName, exchange)
	assert.Equal(f.t, f.cfg.RoutingKey, routingKey)

	return nil
}

func (f fakeMQ) Publish(msg queue.Message, exchange, routingKey string) error {
	return nil
}

func (f fakeMQ) Subscribe(queue, consumer string, receiveCh chan<- queue.ReceivedMessage) {
}

func (f fakeMQ) Inspect(queueName string) (queue.Inspection, error) {
	return queue.Inspection{}, nil
}

func (f fakeMQ) Stop() {

}

func TestScaffold(t *testing.T) {
	fixtures := []struct {
		name                        string
		queue                       string
		exchange                    string
		routingKey                  string
		exchangeType                string
		shouldFailOnDeclareExchange bool
		shouldFailOnDeclareQueue    bool
		shouldFailOnBind            bool
	}{
		{
			name:         "normal scaffold",
			queue:        "some.queue",
			exchange:     "some.exchange",
			routingKey:   "some.routing.key",
			exchangeType: "direct",
		},
		{
			name:                        "normal scaffold",
			queue:                       "some.queue",
			exchange:                    "some.exchange",
			routingKey:                  "some.routing.key",
			exchangeType:                "direct",
			shouldFailOnDeclareExchange: true,
		},
	}

	for _, f := range fixtures {
		t.Run(f.name, func(t *testing.T) {
			cfg := Config{
				QueueName:    f.queue,
				ExchangeName: f.exchange,
				ExchangeType: f.exchangeType,
				RoutingKey:   f.routingKey,
				IsPeristent:  false,
			}

			fmq := fakeMQ{
				cfg:                         cfg,
				t:                           t,
				shouldFailOnDeclareExchange: f.shouldFailOnDeclareExchange,
				shouldFailOnDeclareQueue:    f.shouldFailOnDeclareQueue,
				shouldFailOnBind:            f.shouldFailOnBind,
			}

			fl := NewMQEventFlow(fmq, cfg)

			err := fl.Scaffold()

			if fmq.shouldFailOnDeclareExchange || fmq.shouldFailOnDeclareQueue || fmq.shouldFailOnBind {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

		})
	}
}
