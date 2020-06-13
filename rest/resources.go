package rest

import "github.com/denismitr/auditbase/model"

type resourceSerializer interface {
	ToJSON() responseItem
}

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
	Changes     []model.Change `json:"changes"`
}

type propertyStatAttribute struct {
	Name       string `json:"name"`
	EventCount int    `json:"eventCount"`
}

func newMicroserviceAttributes(m *model.Microservice) *microserviceAttributes {
	return &microserviceAttributes{
		Name:        m.Name,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
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
		Type:        p.Type,
		ChangeCount: p.ChangeCount,
		Changes:     p.Changes,
	}
}
