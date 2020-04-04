package flow

import (
	"encoding/json"

	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/queue"
	"github.com/pkg/errors"
)

type ReceivedEvent interface {
	Event() (model.Event, error)
	CloneMsgToRequeue() queue.Message
	Tag() uint64
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

func (re *QueueReceivedEvent) CloneMsgToRequeue() queue.Message {
	return re.msg.CloneToReque()
}

func (re *QueueReceivedEvent) Tag() uint64 {
	return re.msg.Tag()
}

func (re *QueueReceivedEvent) Attempt() int {
	return re.msg.Attempt()
}
