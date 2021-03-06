package flow

import (
	"github.com/denismitr/auditbase/internal/flow/queue"
)

// State of event flow
type State int

const (
	Idle = iota
	Active
	Failed
	Stopped
)

func queueStatusToFlowState(status queue.ConnectionStatus) State {
	switch status {
	case queue.ConnectionDropped:
		return Failed
	case queue.ConnectionClosed:
		return Stopped
	case queue.Connected:
		return Active
	case queue.Connecting:
		return Failed
	case queue.Idle:
		return Stopped
	default:
		return Stopped
	}
}

type Status struct {
	State     State
	Messages  int
	Consumers int
}

func (s Status) OK() bool {
	return s.State == Active
}

func (s Status) Error() string {
	switch s.State {
	case Failed:
		return "Actions flow has failed"
	case Stopped:
		return "Actions flow is stopped"
	case Idle:
		return "Actions flow is not active"
	default:
		return "Actions flow is not working well"
	}
}
