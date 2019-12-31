package model

type Event struct {
	ID              int                    `json:"id"`
	ActorID         int                    `json:"actorId"`
	ActorType       string                 `json:"actorType"`
	ActorServiceID  int                    `json:"actorServiceId"`
	TargetID        int                    `json:"targetId"`
	TargetType      string                 `json:"targetType"`
	TargetServiceID int                    `json:"targetServiceId"`
	EventType       string                 `json:"eventType"`
	EmittedAt       string                 `json:"emittedAt"`
	RegisteredAt    string                 `json:"registeredAt"`
	Delta           map[string]interface{} `json:"delta"`
}

type EventRepository interface {
	Create(Event) error
	Update(Event) error
	Delete(int) error
	FindOneByID(int) (Event, error)
}

type EventExchange interface {
	Publish(Event) error
	Consume() <-chan Event
}
