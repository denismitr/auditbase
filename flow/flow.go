package flow

import (
	"encoding/json"
	"sync"

	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/queue"
	"github.com/denismitr/auditbase/utils"
	"github.com/pkg/errors"
)

type ReceivedEvent interface {
	Event() (model.Event, error)
	CloneMsgToRequeue() queue.Message
	Tag() uint64
}

// EventFlow interface
type EventFlow interface {
	Send(e model.Event) error
	Receive(queue, consumer string) <-chan ReceivedEvent
	Requeue(ReceivedEvent) error
	Ack(ReceivedEvent) error
	Inspect() (Status, error)
	Scaffold() error
	NotifyOnStateChange(chan State)
	Start()
	Stop() error
}

// MQEventFlow - message queue event flow
type MQEventFlow struct {
	mq             queue.MQ
	cfg            Config
	state          State
	logger         utils.Logger
	mu             sync.RWMutex
	stateListeners []chan State
	stopCh         chan struct{}
	msgCh          chan queue.ReceivedMessage
	eventCh        chan ReceivedEvent
}

// New event flow
func New(mq queue.MQ, logger utils.Logger, cfg Config) *MQEventFlow {
	return &MQEventFlow{
		mq:             mq,
		cfg:            cfg,
		state:          Idle,
		mu:             sync.RWMutex{},
		stateListeners: make([]chan State, 0),
		stopCh:         make(chan struct{}),
		msgCh:          make(chan queue.ReceivedMessage),
		eventCh:        make(chan ReceivedEvent),
	}
}

func (ef *MQEventFlow) updateState(state State) {
	ef.mu.Lock()
	defer ef.mu.Unlock()
	ef.state = state
}

func (ef *MQEventFlow) closeStateChangeListeners() {
	ef.mu.Lock()
	defer ef.mu.Unlock()
	for _, l := range ef.stateListeners {
		close(l)
	}
}

// Start event flow
func (ef *MQEventFlow) Start() {
	listener := make(chan queue.ConnectionStatus)
	ef.mq.NotifyStatusChange(listener)

	go func() {
		for {
			select {
			case status := <-listener:
				ef.updateState(queueStatusToFlowState(status))
			case <-ef.stopCh:
				ef.closeStateChangeListeners()
			}
		}
	}()
}

// NotifyOnStateChange - registers a state change listener
func (ef *MQEventFlow) NotifyOnStateChange(l chan State) {
	ef.mu.Lock()
	defer ef.mu.Unlock()
	ef.stateListeners = append(ef.stateListeners, l)
}

// Send event to the event flow
func (ef *MQEventFlow) Send(e model.Event) error {
	b, err := json.Marshal(&e)
	if err != nil {
		return errors.Wrapf(err, "could not convert event with ID %s to json bytes", e.ID)
	}

	msg := queue.NewJSONMessage(b, 1)

	return ef.mq.Publish(msg, ef.cfg.ExchangeName, ef.cfg.RoutingKey)
}

// Receive events from the flow
func (ef *MQEventFlow) Receive(queue, consumer string) <-chan ReceivedEvent {
	go ef.mq.Subscribe(queue, consumer, ef.msgCh)

	go func() {
		for {
			select {
			case msg := <-ef.msgCh:
				if msg == nil {
					ef.Stop()
					continue
				}

				ef.eventCh <- &QueueReceivedEvent{msg: msg}
			case <-ef.stopCh:
				close(ef.eventCh)
			}
		}
	}()

	return ef.eventCh
}

// Requeue previously received message
func (ef *MQEventFlow) Requeue(re ReceivedEvent) error {
	if err := ef.mq.Reject(re.Tag()); err != nil {
		ef.logger.Error(err)
		return errors.Wrap(err, "could not requeu event")
	}

	msg := re.CloneMsgToRequeue()
	if msg.Attempt() > ef.cfg.MaxRequeue {
		return errors.New("too many attempts")
	}

	if err := ef.mq.Publish(msg, ef.cfg.ExchangeName, ef.cfg.RequeueRoutingKey); err != nil {
		return err
	}

	return nil
}

// Ack previously received message
func (ef *MQEventFlow) Ack(re ReceivedEvent) error {
	return ef.mq.Ack(re.Tag())
}

// Inspect event flow
func (ef *MQEventFlow) Inspect() (Status, error) {
	i, err := ef.mq.Inspect(ef.cfg.QueueName)
	if err != nil {
		return Status{}, err
	}

	ef.mu.RLock()
	defer ef.mu.RUnlock()

	return Status{
		Messages:  i.Messages,
		Consumers: i.Consumers,
		State:     ef.state,
	}, nil
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

	if ef.cfg.ErrorQueueName != ef.cfg.QueueName {
		if err := ef.mq.DeclareQueue(ef.cfg.ErrorQueueName); err != nil {
			return errors.Wrap(err, "could not scaffold DirectEventExchange on error queue declaration")
		}

		if err := ef.mq.Bind(ef.cfg.ErrorQueueName, ef.cfg.ExchangeName, ef.cfg.RequeueRoutingKey); err != nil {
			return errors.Wrap(err, "could not scaffold DirectEventExchange on error queue binding")
		}
	}

	return nil
}
