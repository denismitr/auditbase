package model

type Property struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	EntityID    string   `json:"entityId"`
	Type        string   `json:"type"`
	ChangeCount int      `json:"changeCount"`
	Changes     []Change `json:"changes"`
}

type PropertyStat struct {
	Name       string `json:"name"`
	EventCount int    `json:"eventCount"`
}

type PropertyRepository interface {
	GetIDOrCreate(name, entityID string) (string, error)
}