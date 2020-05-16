package model

type Property struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	EventID     string  `json:"eventId"`
	EntityID    string  `json:"entityId"`
	ChangedFrom *string `json:"changedFrom"`
	ChangedTo   *string `json:"changedTo"`
}

type PropertyStat struct {
	Name       string `json:"name"`
	EventCount int    `json:"eventCount"`
}
