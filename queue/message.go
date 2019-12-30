package queue

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type Message interface {
	Body() ([]byte, error)
	ContentType() string
}

type Delivery struct {
	QueueName   string
	RoutingKey  string
	IsPeristent bool
}

type QueueMessage struct {
	Namespace     string                 `json:"namespace"`
	ActorID       string                 `json:"actorId"`
	ActorType     string                 `json:"actorType"`
	ActorService  string                 `json:"actorService"`
	TargetID      string                 `json:"targetId"`
	TargetType    string                 `json:"targetType"`
	TargetService string                 `json:"targetService"`
	Operation     string                 `json:"operation"`
	EmittedAt     string                 `json:"emittedAt"`
	Delta         map[string]interface{} `json:"delta"`
}

type ReceivedMessage struct {
	Queue   string
	Message QueueMessage
}

func (m QueueMessage) ContentType() string {
	return "application/json"
}

func (m QueueMessage) Body() ([]byte, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, errors.Wrapf(err, "could not serialize message body in namespace %s", m.Namespace)
	}

	return b, nil
}
