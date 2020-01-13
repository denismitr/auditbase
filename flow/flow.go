package flow

import (
	"encoding/json"

	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/queue"
	"github.com/pkg/errors"
)

// EventFlow ...
type EventFlow interface {
	Send(e model.Event) error
	Receive(consumer string) <-chan ReceivedEvent
	Inspect() (int, int, error)
	Scaffold() error
	Stop() error
}

// MQEventFlow - message queue event flow
type MQEventFlow struct {
	mq      queue.MQ
	cfg     Config
	stopCh  chan struct{}
	msgCh   chan queue.ReceivedMessage
	eventCh chan ReceivedEvent
}

func NewMQEventFlow(mq queue.MQ, cfg Config) *MQEventFlow {
	return &MQEventFlow{
		mq:      mq,
		cfg:     cfg,
		stopCh:  make(chan struct{}),
		msgCh:   make(chan queue.ReceivedMessage),
		eventCh: make(chan ReceivedEvent),
	}
}

// Send event to the event flow
func (ef *MQEventFlow) Send(e model.Event) error {
	b, err := json.Marshal(&e)
	if err != nil {
		return errors.Wrapf(err, "could not convert event with ID %s to json bytes", e.ID)
	}

	msg := queue.NewJSONMessage(b)

	return ef.mq.Publish(msg, ef.cfg.ExchangeName, ef.cfg.RoutingKey)
}

// Receive events from the flow
func (ef *MQEventFlow) Receive(consumer string) <-chan ReceivedEvent {
	go ef.mq.Subscribe(ef.cfg.QueueName, "event_flow_consumer", ef.msgCh)

	go func() {
		for {
			select {
			case msg := <-ef.msgCh:
				ef.eventCh <- &QueueReceivedEvent{msg}
			case <-ef.stopCh:
				close(ef.eventCh)
			}
		}
	}()

	return ef.eventCh
}

func (ef *MQEventFlow) Inspect() (messages int, consumers int, err error) {
	i, err := ef.mq.Inspect(ef.cfg.QueueName)
	if err != nil {
		return
	}

	messages = i.Messages
	consumers = i.Consumers

	return
}

// Stop event flow
func (ef *MQEventFlow) Stop() error {
	close(ef.stopCh)
	return nil
}

// Scaffold the the exchange, queue and binding
func (ef *MQEventFlow) Scaffold() error {
	if err := ef.mq.DeclareExchange(ef.cfg.ExchangeName, ef.cfg.ExchangeType); err != nil {
		return errors.Wrap(err, "could not scaffold DirectEventExchange on exchage declaration")
	}

	if err := ef.mq.DeclareQueue(ef.cfg.QueueName); err != nil {
		return errors.Wrap(err, "could not scaffold DirectEventExchange on queue declaration")
	}

	if err := ef.mq.Bind(ef.cfg.QueueName, ef.cfg.ExchangeName, ef.cfg.RoutingKey); err != nil {
		return errors.Wrap(err, "could not scaffold DirectEventExchange on queue binding")
	}

	return nil
}
