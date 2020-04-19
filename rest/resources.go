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

func newMicroserviceAttributes(m *model.Microservice) *microserviceAttributes {
	return &microserviceAttributes{
		Name:        m.Name,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
