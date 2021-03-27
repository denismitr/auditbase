package flow

import (
	"encoding/json"

	"github.com/denismitr/auditbase/internal/model"
	"github.com/denismitr/auditbase/internal/queue"
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

func (qra *QueueReceivedAction) NewAction() (*model.NewAction, error) {
	a := model.NewAction{}

	if err := json.Unmarshal(qra.msg.Body(), &a); err != nil {
		return nil, errors.Wrap(err, "could not get event from received queue message bytes")
	}

	return &a, nil
}

func (qra *QueueReceivedAction) CloneMsgToRequeue() queue.Message {
	return qra.msg.CloneToReque()
}

func (qra *QueueReceivedAction) Tag() uint64 {
	return qra.msg.Tag()
}

func (qra *QueueReceivedAction) Attempt() int {
	return qra.msg.Attempt()
}
