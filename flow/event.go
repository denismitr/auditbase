package flow

import (
	"encoding/json"

	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/queue"
	"github.com/pkg/errors"
)

type ReceivedEvent interface {
	Event() (model.Event, error)
	Ack() error
	Reject() error
	Postpone() error
}

type QueueReceivedEvent struct {
	msg queue.ReceivedMessage
}

func (re *QueueReceivedEvent) Event() (model.Event, error) {
	e := model.Event{}

	if err := json.Unmarshal(re.msg.Body(), &e); err != nil {
		return e, errors.Wrap(err, "could not get event from received queue message bytes")
	}

	return e, nil
}

func (re *QueueReceivedEvent) Ack() error {
	return re.msg.Ack()
}

func (re *QueueReceivedEvent) Reject() error {
	return re.msg.Reject(false)
}

func (re *QueueReceivedEvent) Postpone() error {
	return re.msg.Reject(true)
}
