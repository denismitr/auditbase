package flow

// Config of the event exchange
type Config struct {
	ExchangeName       string
	ActionsCreateQueue string
	ActionsUpdateQueue string
	ExchangeType       string
	Concurrency        int
	MaxRequeue         int
	IsPeristent        bool
}
