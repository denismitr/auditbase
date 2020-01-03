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
	EmittedAt       string                   `json:"emittedAt"`
	RegisteredAt    string                   `json:"registeredAt"`
	Delta           map[string][]interface{} `json:"delta"`
}

type EventRepository interface {
	Create(Event) error
	Update(int, Event) error
	Delete(int) error
	FindOneByID(int) (Event, error)
	SelectAll() ([]Event, error)
}

type EventExchange interface {
	Publish(Event) error
	Consume() <-chan Event
}
