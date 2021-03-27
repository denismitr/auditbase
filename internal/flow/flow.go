package flow

import (
	"encoding/json"
	"sync"

	"github.com/denismitr/auditbase/internal/model"
	"github.com/denismitr/auditbase/internal/queue"
	"github.com/denismitr/auditbase/internal/utils/logger"
	"github.com/pkg/errors"
)

// ActionFlow interface
type ActionFlow interface {
	Send(e *model.NewAction) error
	Receive(queue, consumer string) <-chan ReceivedAction
	Requeue(ReceivedAction) error
	Ack(ReceivedAction) error
	Reject(ReceivedAction) error
	Inspect() (Status, error)
	Scaffold() error
	NotifyOnStateChange(chan State)
	Start()
	Stop() error
}

// MQActionFlow - message queue event flow
type MQActionFlow struct {
	mq             queue.MQ
	cfg            Config
	state          State
	lg             logger.Logger
	mu             sync.RWMutex
	stateListeners []chan State
	stopCh         chan struct{}
	msgCh          chan queue.ReceivedMessage
	eventCh        chan ReceivedAction
}

// New event flow
func New(mq queue.MQ, lg logger.Logger, cfg Config) *MQActionFlow {
	return &MQActionFlow{
		mq:             mq,
		cfg:            cfg,
		state:          Idle,
		lg:             lg,
		mu:             sync.RWMutex{},
		stateListeners: make([]chan State, 0),
		stopCh:         make(chan struct{}),
		msgCh:          make(chan queue.ReceivedMessage),
		eventCh:        make(chan ReceivedAction),
	}
}

func (ef *MQActionFlow) updateState(state State) {
	ef.mu.Lock()
	defer ef.mu.Unlock()
	ef.state = state
}

func (ef *MQActionFlow) closeStateChangeListeners() {
	ef.mu.Lock()
	defer ef.mu.Unlock()
	for _, l := range ef.stateListeners {
		close(l)
	}
}

// Start event flow
func (ef *MQActionFlow) Start() {
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
func (ef *MQActionFlow) NotifyOnStateChange(l chan State) {
	ef.mu.Lock()
	defer ef.mu.Unlock()
	ef.stateListeners = append(ef.stateListeners, l)
}

// Send event to the event flow
func (ef *MQActionFlow) Send(newAction *model.NewAction) error {
	b, err := json.Marshal(newAction)
	if err != nil {
		return errors.Wrapf(err, "could not convert event with ID %s to json bytes", newAction.UID)
	}

	msg := queue.NewJSONMessage(b, 1)

	return ef.mq.Publish(msg, ef.cfg.ExchangeName, ef.cfg.RoutingKey)
}

// Receive events from the flow
func (ef *MQActionFlow) Receive(queue, consumer string) <-chan ReceivedAction {
	go func() {
		if err := ef.mq.Subscribe(queue, consumer, ef.msgCh); err != nil {
			panic(err)
		}
	}()

	go func() {
		for {
			select {
			case msg := <-ef.msgCh:
				if msg == nil {
					_ = ef.Stop()
					continue
				}

				ef.eventCh <- &QueueReceivedAction{msg: msg}
			case <-ef.stopCh:
				//close(ef.eventCh)
			}
		}
	}()

	return ef.eventCh
}

// Requeue previously received message
func (ef *MQActionFlow) Requeue(re ReceivedAction) error {
	if err := ef.mq.Reject(re.Tag()); err != nil {
		ef.lg.Error(err)
		return ErrCannotRequeueAction
	}

	msg := re.CloneMsgToRequeue()
	if msg.Attempt() > ef.cfg.MaxRequeue {
		return ErrTooManyAttempts
	}

	if err := ef.mq.Publish(msg, ef.cfg.ExchangeName, ef.cfg.RequeueRoutingKey); err != nil {
		return err
	}

	return nil
}

// Ack received message
func (ef *MQActionFlow) Ack(re ReceivedAction) error {
	return ef.mq.Ack(re.Tag())
}

// Reject received message
func (ef *MQActionFlow) Reject(re ReceivedAction) error {
	return ef.mq.Reject(re.Tag())
}

// Inspect event flow
func (ef *MQActionFlow) Inspect() (Status, error) {
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
func (ef *MQActionFlow) Stop() error {
	close(ef.stopCh)
	return nil
}

// Scaffold the the exchange, queue and binding
func (ef *MQActionFlow) Scaffold() error {
	if err := ef.mq.DeclareExchange(ef.cfg.ExchangeName, ef.cfg.ExchangeType); err != nil {
		return errors.Wrap(err, "could not scaffold DirectActionExchange on exchage declaration")
	}

	if err := ef.mq.DeclareQueue(ef.cfg.QueueName); err != nil {
		return errors.Wrap(err, "could not scaffold DirectActionExchange on queue declaration")
	}

	if err := ef.mq.Bind(ef.cfg.QueueName, ef.cfg.ExchangeName, ef.cfg.RoutingKey); err != nil {
		return errors.Wrap(err, "could not scaffold DirectActionExchange on queue binding")
	}

	if ef.cfg.ErrorQueueName != ef.cfg.QueueName {
		if err := ef.mq.DeclareQueue(ef.cfg.ErrorQueueName); err != nil {
			return errors.Wrap(err, "could not scaffold DirectActionExchange on error queue declaration")
		}

		if err := ef.mq.Bind(ef.cfg.ErrorQueueName, ef.cfg.ExchangeName, ef.cfg.RequeueRoutingKey); err != nil {
			return errors.Wrap(err, "could not scaffold DirectActionExchange on error queue binding")
		}
	}

	return nil
}
