package flow

import (
	"encoding/json"

	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/queue"
	"github.com/pkg/errors"
)

type ReceivedAction interface {
	NewAction() (*model.NewAction, error)
	CloneMsgToRequeue() queue.Message
	Tag() uint64
}

type QueueReceivedAction struct {
	msg queue.ReceivedMessage
}

func (re *QueueReceivedAction) NewAction() (*model.NewAction, error) {
	a := model.NewAction{}

	if err := json.Unmarshal(re.msg.Body(), &a); err != nil {
		return nil, errors.Wrap(err, "could not get event from received queue message bytes")
	}

	return &a, nil
}

func (re *QueueReceivedAction) CloneMsgToRequeue() queue.Message {
	return re.msg.CloneToReque()
}

func (re *QueueReceivedAction) Tag() uint64 {
	return re.msg.Tag()
}

func (re *QueueReceivedAction) Attempt() int {
	return re.msg.Attempt()
}
