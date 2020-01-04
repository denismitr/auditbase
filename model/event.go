package model

type Event struct {
	ID              string                   `json:"id"`
	ParentEventID   string                   `json:"parentEventId"`
	ActorID         string                   `json:"actorId"`
	ActorType       string                   `json:"actorType"`
	ActorServiceID  string                   `json:"actorServiceId"`
	TargetID        string                   `json:"targetId"`
	TargetType      string                   `json:"targetType"`
	TargetServiceID string                   `json:"targetServiceId"`
	EventName       string                   `json:"eventName"`
	EmittedAt       int64                    `json:"emittedAt"`
	RegisteredAt    int64                    `json:"registeredAt"`
	Delta           map[string][]interface{} `json:"delta"`
}

type EventRepository interface {
	Create(Event) error
	Delete(string) error
	FindOneByID(string) (Event, error)
	SelectAll() ([]Event, error)
}

type EventExchange interface {
	Publish(Event) error
	Consume() <-chan Event
}
