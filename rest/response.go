package rest

import "github.com/denismitr/auditbase/model"

type itemResponse struct {
	Data *jsonApiResponse `json:"data"`
	Meta interface{}      `json:"meta,omitempty"`
}

type collectionResponse struct {
	Data []*jsonApiResponse `json:"data"`
	Meta interface{}        `json:"meta,omitempty"`
}

type jsonApiResponse struct {
	Type       string      `json:"type"`
	ID         string      `json:"id"`
	Attributes interface{} `json:"attributes,omitempty"`
}

func newItemResponse(r *jsonApiResponse) *itemResponse {
	return &itemResponse{Data: r}
}

func newItemResponseWithMeta(r *jsonApiResponse, meta interface{}) *itemResponse {
	return &itemResponse{Data: r, Meta: meta}
}

func newCollectionResponse(r []*jsonApiResponse) *collectionResponse {
	return &collectionResponse{Data: r}
}

func newJsonApiResponse(typ, id string, attributes interface{}) *jsonApiResponse {
	return &jsonApiResponse{
		Type:       typ,
		ID:         id,
		Attributes: attributes,
	}
}

type statusMessage map[string]string

func respondAccepted(typ, id string) (int, *itemResponse) {
	return 202, newItemResponse(newJsonApiResponse(typ, id, nil))
}

func newEventCountResponse(count int) *itemResponse {
	return newItemResponseWithMeta(nil, map[string]int{"count": count})
}

func newEventResponse(m *model.Event) *itemResponse {
	return newItemResponse(newJsonApiResponse("events", m.ID, m))
}

func newEventsResponse(events []*model.Event) *collectionResponse {
	items := make([]*jsonApiResponse, len(events))
	for i := range events {
		items[i] = newJsonApiResponse("events", events[i].ID, events[i])
	}
	return newCollectionResponse(items)
}

func newMicroserviceResponse(m *model.Microservice) *itemResponse {
	return newItemResponse(newJsonApiResponse("microservices", m.ID, newMicroserviceAttributes(m)))
}

func newEntityResponse(e *model.Entity) *itemResponse {
	return newItemResponse(newJsonApiResponse("entities", e.ID, newEntityAttributes(e)))
}

func newEntitiesResponse(es []*model.Entity) *collectionResponse {
	items := make([]*jsonApiResponse, len(es))
	for i := range es {
		items[i] = newJsonApiResponse("entities", es[i].ID, newEntityAttributes(es[i]))
	}
	return newCollectionResponse(items)
}

func newEntityWithPropertiesResponse(e *model.Entity, ps []*model.PropertyStat) *itemResponse {
	return newItemResponse(
		newJsonApiResponse("entities", e.ID, newEntityWithPropertiesAttributes(e, ps)))
}

func newMicroservicesResponse(ms []*model.Microservice) *collectionResponse {
	items := make([]*jsonApiResponse, len(ms))
	for i := range ms {
		items[i] = newJsonApiResponse("microservices", ms[i].ID, newMicroserviceAttributes(ms[i]))
	}
	return newCollectionResponse(items)
}
