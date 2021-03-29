package flow

import (
	"encoding/json"
	"sync"

	"github.com/denismitr/auditbase/internal/flow/queue"
	"github.com/denismitr/auditbase/internal/model"
	"github.com/denismitr/auditbase/internal/utils/logger"
	"github.com/pkg/errors"
)

type Scaffolder interface {
	Scaffold() error
}

// ActionFlow interface
type ActionFlow interface {
	SendNewAction(e *model.NewAction) error
	SendUpdateAction(e *model.UpdateAction) error
	ReceiveNewActions(consumer string, h NewActionHandler)
	ReceiveUpdateActions(consumer string, h UpdateActionHandler)
	NotifyOnConnectionLoss(chan<- struct{})
	Start()
	Stop() error

	Scaffolder
}

var _ ActionFlow = (*MQActionFlow)(nil)

// MQActionFlow - message queue event flow
type MQActionFlow struct {
	mq                queue.MQ
	cfg               Config
	state             State
	lg                logger.Logger
	mu                sync.RWMutex
	connLossListeners []chan<- struct{}
	stopCh            chan struct{}
	msgCh             chan queue.ReceivedMessage
}

// New event flow
func New(mq queue.MQ, lg logger.Logger, cfg Config) *MQActionFlow {
	return &MQActionFlow{
		mq:                mq,
		cfg:               cfg,
		state:             Idle,
		lg:                lg,
		mu:                sync.RWMutex{},
		connLossListeners: make([]chan<- struct{}, 0),
		stopCh:            make(chan struct{}),
		msgCh:             make(chan queue.ReceivedMessage),
	}
}

func (af *MQActionFlow) updateState(state State) {
	af.mu.Lock()
	defer af.mu.Unlock()
	af.state = state

	if af.state == Failed || af.state == Stopped {
		for _, l := range af.connLossListeners {
			select {
			 	case l <- struct{}{}:
			}
		}
	}
}

func (af *MQActionFlow) closeStateChangeListeners() {
	af.mu.Lock()
	defer af.mu.Unlock()
	for _, l := range af.connLossListeners {
		close(l)
	}
}

// Start event flow
func (af *MQActionFlow) Start() {
	listener := make(chan queue.ConnectionStatus)
	af.mq.NotifyStatusChange(listener)

	go af.mq.MaintainConnection()

	go func() {
		for {
			select {
			case status := <-listener:
				af.updateState(queueStatusToFlowState(status))
			case <-af.stopCh:
				af.closeStateChangeListeners()
			}
		}
	}()
}

// Ack received message
func (af *MQActionFlow) Ack(rm queue.ReceivedMessage) error {
	return af.mq.Ack(rm.ID())
}

// Reject received message
func (af *MQActionFlow) Reject(rm queue.ReceivedMessage) error {
	return af.mq.Reject(rm.ID(), false)
}

// Stop the actions flow
func (af *MQActionFlow) Stop() error {
	close(af.stopCh)
	return nil
}

// NotifyOnConnectionLoss - registers a state change listener
func (af *MQActionFlow) NotifyOnConnectionLoss(l chan<- struct{}) {
	af.mu.Lock()
	defer af.mu.Unlock()
	af.connLossListeners = append(af.connLossListeners, l)
}

// SendNewAction event to the event flow
func (af *MQActionFlow) SendNewAction(newAction *model.NewAction) error {
	b, err := json.Marshal(newAction)
	if err != nil {
		return errors.Wrapf(err, "could not convert event with UID %s to json bytes", newAction.UID)
	}

	msg := queue.NewJSONMessage(b, 1)

	return af.mq.Publish(msg, af.cfg.ExchangeName, af.cfg.ActionsCreateQueue)
}

// SendUpdateAction event to the event flow
func (af *MQActionFlow) SendUpdateAction(updateAction *model.UpdateAction) error {
	b, err := json.Marshal(updateAction)
	if err != nil {
		return errors.Wrapf(err, "could not convert event with UID %s to json bytes", updateAction.UID)
	}

	msg := queue.NewJSONMessage(b, 1)

	return af.mq.Publish(msg, af.cfg.ExchangeName, af.cfg.ActionsCreateQueue)
}

type ProcessFunc func(message queue.ReceivedMessage) error
type NewActionHandler func(*model.NewAction) error
type UpdateActionHandler func(*model.UpdateAction) error

func (af *MQActionFlow) ReceiveNewActions(consumerName string, h NewActionHandler) {
	processor := newActionsProcessor(h)
	af.receive(af.cfg.ActionsCreateQueue, consumerName + "_new_actions", af.cfg.Concurrency, processor)
}

func (af *MQActionFlow) ReceiveUpdateActions(consumerName string, h UpdateActionHandler) {
	processor := updateActionsProcessor(h)
	af.receive(af.cfg.ActionsUpdateQueue, consumerName + "_update_actions", af.cfg.Concurrency, processor)
}

func newActionsProcessor(h NewActionHandler) ProcessFunc {
	return func(msg queue.ReceivedMessage) error {
		na := model.NewAction{}

		if err := json.Unmarshal(msg.Body(), &na); err != nil {
			return errors.Wrap(err, "could not parse 'newAction' model from received queue message bytes")
		}

		if err := h(&na); err != nil {
			return err
		}

		return nil
	}
}

func updateActionsProcessor(h UpdateActionHandler) ProcessFunc {
	return func(msg queue.ReceivedMessage) error {
		ua := model.UpdateAction{}

		if err := json.Unmarshal(msg.Body(), &ua); err != nil {
			return errors.Wrap(err, "could not parse 'updateAction' model from received queue message bytes")
		}

		if err := h(&ua); err != nil {
			return err
		}

		return nil
	}
}

// receive actions from the flow of data
func (af *MQActionFlow) receive(queue, consumer string, concurrency int, msgProcessor ProcessFunc) {
	go func() {
		if err := af.mq.Subscribe(queue, consumer, af.msgCh); err != nil {
			panic(err)
		}
	}()

	sem := make(chan struct{}, concurrency)

	for {
		select {
		case msg, ok := <-af.msgCh:
			if !ok {
				if err := af.Stop(); err != nil {
					af.lg.Error(err)
				}
				continue
			}

			sem <- struct{}{}
			go func() {
				defer func() { <-sem }()

				if err := msgProcessor(msg); err != nil {
					af.lg.Error(err)
					if err := af.requeue(msg, queue); err != nil {
						af.lg.Error(err)
						if err := af.Reject(msg); err != nil {
							af.lg.Error(err)
						}
					}

					// message was requeued or rejected at this point
					// if reject has also failed, there is nothing we can do
					// we return anyway and try to process the next message
					return
				}

				// all good - message can be officially acked
				if err := af.Ack(msg); err != nil {
					af.lg.Error(err)
				}
			}()
		case <-af.stopCh:
			return
		}
	}
}

// Requeue previously received message
func (af *MQActionFlow) requeue(rm queue.ReceivedMessage, queue string) error {
	// reject original message version
	if err := af.mq.Reject(rm.ID(), false); err != nil {
		af.lg.Error(err)
		return ErrCannotRequeueAction
	}

	// create a copy
	msg := rm.CloneToRequeue()
	if msg.Attempt() > af.cfg.MaxRequeue {
		return ErrTooManyAttempts
	}

	// requeue
	if err := af.mq.Publish(msg, af.cfg.ExchangeName, queue); err != nil {
		return err
	}

	return nil
}
