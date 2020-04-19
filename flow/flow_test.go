package flow

// import (
// 	"testing"

// 	"github.com/denismitr/auditbase/test/mock_queue"
// 	"github.com/denismitr/auditbase/utils/logger"
// 	"github.com/golang/mock/gomock"
// 	"github.com/stretchr/testify/assert"
// )

// func TestFlow(t *testing.T) {
// 	assert.Equal(t, true, true)
// }

// func TestScaffold(t *testing.T) {
// 	logger := logger.NewStdoutLogger("test", "test")

// 	fixtures := []struct {
// 		name                        string
// 		queue                       string
// 		errorQueue                  string
// 		exchange                    string
// 		routingKey                  string
// 		requeueRoutingKey           string
// 		exchangeType                string
// 		shouldFailOnDeclareExchange bool
// 		shouldFailOnDeclareQueue    bool
// 		shouldFailOnBind            bool
// 	}{
// 		{
// 			name:              "normal scaffold",
// 			queue:             "some.queue",
// 			errorQueue:        "some.error.queue",
// 			exchange:          "some.exchange",
// 			routingKey:        "some.routing.key",
// 			requeueRoutingKey: "some.requeue.routing.key",
// 			exchangeType:      "direct",
// 		},
// 		{
// 			name:                        "normal scaffold",
// 			queue:                       "some.queue",
// 			errorQueue:                  "some.error.queue",
// 			exchange:                    "some.exchange",
// 			routingKey:                  "some.routing.key",
// 			requeueRoutingKey:           "some.requeue.routing.key",
// 			exchangeType:                "direct",
// 			shouldFailOnDeclareExchange: true,
// 		},
// 	}

// 	for _, f := range fixtures {
// 		t.Run(f.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			cfg := Config{
// 				QueueName:         f.queue,
// 				ErrorQueueName:    f.errorQueue,
// 				ExchangeName:      f.exchange,
// 				ExchangeType:      f.exchangeType,
// 				RoutingKey:        f.routingKey,
// 				RequeueRoutingKey: f.requeueRoutingKey,
// 				IsPeristent:       false,
// 			}

// 			fmq := mock_queue.NewMockMQ(ctrl)
// 			fmq.EXPECT().DeclareExchange(cfg.ExchangeName, cfg.ExchangeType).Return(nil)
// 			fmq.EXPECT().DeclareQueue(cfg.QueueName).Return(nil)
// 			fmq.EXPECT().DeclareQueue(cfg.ErrorQueueName).Return(nil)
// 			fmq.EXPECT().Bind(cfg.QueueName, cfg.ExchangeName, cfg.RoutingKey).Return(nil)
// 			fmq.EXPECT().Bind(cfg.ErrorQueueName, cfg.ExchangeName, cfg.RequeueRoutingKey).Return(nil)

// 			fl := New(fmq, logger, cfg)

// 			err := fl.Scaffold()
// 			assert.NoError(t, err)
// 		})
// 	}
// }
