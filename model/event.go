package model

type Event struct {
	ID            string                   `json:"id"`
	ParentEventID string                   `json:"parentEventId"`
	ActorID       string                   `json:"actorId"`
	ActorType     ActorType                `json:"actorType"`
	ActorService  Microservice             `json:"actorService"`
	TargetID      string                   `json:"targetId"`
	TargetType    TargetType               `json:"targetType"`
	TargetService Microservice             `json:"targetService"`
	EventName     string                   `json:"eventName"`
	EmittedAt     int64                    `json:"emittedAt"`
	RegisteredAt  int64                    `json:"registeredAt"`
	Delta         map[string][]interface{} `json:"delta"`
}

func (e *Event) Validate(v Validator) ValidationErrors {
	if !v.IsUUID4(e.ID) {
		v.Add("id", ":id is not a valid UUID4")
	}

	if v.IsEmptyString(e.ActorID) {
		v.Add("actorID", ":actorID must not be empty")
	}

	if v.IsEmptyString(e.ActorType.Name) {
		v.Add("actorType.Name", ":actorType.ID must not be empty")
	}

	if v.IsEmptyString(e.TargetType.Name) {
		v.Add("targetType.Name", ":targetType.ID must not be empty")
	}

	if v.IsEmptyString(e.ActorService.Name) {
		v.Add("actorService.Name", ":actorService.Name must not be empty")
	}

	if v.IsEmptyString(e.TargetService.Name) {
		v.Add("targetService.Name", ":targetService.Name must not be empty")
	}

	return v.Errors()
}

type EventRepository interface {
	Create(Event) error
	Delete(string) error
	Count() (int, error)
	FindOneByID(string) (Event, error)
	SelectAll() ([]Event, error)
}
