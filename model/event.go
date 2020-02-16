package model

type Order string

const DESCOrder Order = "DESC"
const ASCOrder Order = "ASC"

type Sort struct {
	Items []map[string]Order
}

type Pagination struct {
	Page    int
	PerPage int
}

type EventFilter struct {
	ActorID         string
	ActorTypeID     string
	TargetID        string
	TargetTypeID    string
	ActorServiceID  string
	TargetServiceID string
	EventName       string
	EmittedAtGt     int64
	EmittedAtLt     int64
}

func (ef EventFilter) Empty() bool {
	return ef.ActorID == "" && ef.ActorTypeID == ""
}

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
	FindOneByID(ID) (Event, error)
	Select(EventFilter, Sort, Pagination) ([]Event, error)
}
