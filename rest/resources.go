package rest

import (
	"github.com/denismitr/auditbase/model"
	"time"
)

type responseItem struct {
	Data interface{} `json:"data"`
}

type inspectResource struct {
	ConnectionStatus string `json:"connectionStatus"`
	Messages         int    `json:"messages"`
	Consumers        int    `json:"consumers"`
}

func (ir inspectResource) ToJSON() responseItem {
	return responseItem{Data: ir}
}

type microserviceAttributes struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt,omitempty"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
}

type entityAttributes struct {
	Name        string                  `json:"name"`
	ServiceID   string                  `json:"serviceId"`
	Description string                  `json:"description"`
	CreatedAt   string                  `json:"createdAt,omitempty"`
	UpdatedAt   string                  `json:"updatedAt,omitempty"`
	Properties  []propertyStatAttribute `json:"properties,omitempty"`
}

type propertyAttributes struct {
	Name        string         `json:"name"`
	EntityID    string         `json:"entityId"`
	Type        string         `json:"type"`
	ChangeCount int            `json:"changeCount"`
	Changes     []model.Change `json:"changes,omitempty"`
	LastEventAt *time.Time     `json:"lastEventAt"`
}

type propertyStatAttribute struct {
	Name       string `json:"name"`
	EventCount int    `json:"eventCount"`
}

type changeAttributes struct {
	ID              string         `json:"id"`
	EventID         string         `json:"eventId"`
	PropertyID      string         `json:"propertyId"`
	CurrentDataType model.DataType `json:"currentDataType"`
	From            *string        `json:"from"`
	To              *string        `json:"to"`
}

func newMicroserviceAttributes(m *model.Microservice) *microserviceAttributes {
	return &microserviceAttributes{
		Name:        m.Name,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func newChangeAttributes(c *model.Change) *changeAttributes {
	return &changeAttributes{
		ID:              c.ID,
		EventID:         c.EventID,
		PropertyID:      c.PropertyID,
		CurrentDataType: c.CurrentDataType,
		From:            c.From,
		To:              c.To,
	}
}

func newEntityAttributes(e *model.Entity) *entityAttributes {
	return &entityAttributes{
		Name:        e.Name,
		ServiceID:   e.ServiceID,
		Description: e.Description,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

func newPropertyAttributes(p *model.Property) *propertyAttributes {
	return &propertyAttributes{
		Name:        p.Name,
		EntityID:    p.EntityID,
		ChangeCount: p.ChangeCount,
		LastEventAt: p.LastEventAt,
		Changes:     p.Changes,
	}
}
