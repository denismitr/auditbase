package model

import (
	"fmt"
	"time"
)

type EntityType struct {
	ID          ID            `json:"id"`
	ServiceID   ID            `json:"serviceId"`
	Service     *Microservice `json:"service,omitempty"`
	IsActor     bool          `json:"is_actor"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	EntitiesCnt int           `json:"entitiesCount"`
	CreatedAt   time.Time     `json:"createdAt,omitempty"`
	UpdatedAt   time.Time     `json:"updatedAt,omitempty"`
}

// Entities - represents something that can act on data
// or be acted on, or both
type Entity struct {
	ID           ID          `json:"id"`
	ExternalID   string      `json:"externalId"`
	EntityTypeID ID          `json:"entityTypeId"`
	EntityType   *EntityType `json:"entityType,omitempty"`
	CreatedAt    time.Time   `json:"createdAt,omitempty"`
	UpdatedAt    time.Time   `json:"updatedAt,omitempty"`
}

type EntityCollection struct {
	Items []Entity `json:"data"`
	Meta  Meta     `json:"meta"`
}

type EntityTypeCollection struct {
	Items []EntityType `json:"data"`
	Meta  Meta         `json:"meta"`
}

func EntityItemCacheKey(name string, microservice *Microservice) string {
	return fmt.Sprintf("entity_%s_%s", name, microservice.ID)
}
