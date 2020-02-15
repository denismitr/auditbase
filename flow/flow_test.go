package flow

import (
	"testing"

	"github.com/denismitr/auditbase/test/mock_queue"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestFlow(t *testing.T) {
	assert.Equal(t, true, true)
}

type declareExchangeExpectations func(exchangeName, exchangeType string)

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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cfg := Config{
				QueueName:    f.queue,
				ExchangeName: f.exchange,
				ExchangeType: f.exchangeType,
				RoutingKey:   f.routingKey,
				IsPeristent:  false,
			}

			fmq := mock_queue.NewMockMQ(ctrl)
			fmq.EXPECT().DeclareExchange(cfg.ExchangeName, cfg.ExchangeType).Return(nil)
			fmq.EXPECT().DeclareQueue(cfg.QueueName).Return(nil)
			fmq.EXPECT().Bind(cfg.QueueName, cfg.ExchangeName, cfg.RoutingKey).Return(nil)

			fl := New(fmq, cfg)

			err := fl.Scaffold()
			assert.NoError(t, err)
		})
	}
}
