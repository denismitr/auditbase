package flow

import (
	"github.com/denismitr/auditbase/queue"
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
		return "Events flow has failed"
	case Stopped:
		return "Events flow is stopped"
	case Idle:
		return "Events flow is not active"
	default:
		return "Events flow is not working well"
	}
}
