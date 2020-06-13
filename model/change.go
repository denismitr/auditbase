package model

type Change struct {
	ID         string    `json:"id"`
	EventID    string    `json:"eventId"`
	PropertyID string    `json:"propertyId"`
	From       *string   `json:"from"`
	To         *string   `json:"to"`
	Property   *Property `json:"property,omitempty"`
}

type PropertyChange struct {
	ID           string  `json:"id"`
	EventID      string  `json:"eventId"`
	PropertyID   string  `json:"propertyId"`
	Type         string  `json:"type"`
	From         *string `json:"from"`
	To           *string `json:"to"`
	PropertyName string  `json:"property,omitempty"`
	EntityID     string  `json:"entityId,omitempty"`
}
