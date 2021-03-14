package mysql

import (
	"encoding/json"
	"github.com/denismitr/auditbase/model"
)

func mapEntitiesToCollection(items []entityRecord, cnt int, page, perPage uint) *model.EntityCollection {
	result := model.EntityCollection{}
	for _, e := range items {
		result.Items = append(result.Items, *mapEntityRecordToModel(e))
	}

	result.Meta.Total = cnt
	result.Meta.Page = int(page)
	result.Meta.PerPage = int(perPage)

	return &result
}

func mapEntityRecordToModel(e entityRecord) *model.Entity {
	return &model.Entity{
		ID:           model.ID(e.ID),
		ExternalID:   e.ExternalID,
		EntityTypeID: model.ID(e.EntityTypeID),
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
	}
}

func mapEntityRecordAllJoinedToModel(e entityRecordAllJoined) *model.Entity {
	return &model.Entity{
		ID:           model.ID(e.EntityID),
		ExternalID:   e.EntityExternalID,
		EntityTypeID: model.ID(e.EntityTypeID),
		CreatedAt:    e.EntityCreatedAt,
		UpdatedAt:    e.EntityUpdatedAt,
		EntityType: &model.EntityType{
			ID: model.ID(e.EntityTypeID),
			Name: e.EntityTypeName,
			Description: e.EntityTypeDescription,
			ServiceID: model.ID(e.ServiceID),
			CreatedAt: e.EntityTypeCreatedAt,
			UpdatedAt: e.EntityTypeUpdatedAt,
			Service: &model.Microservice{
				ID: model.ID(e.ServiceID),
				Name: e.ServiceName,
				Description: e.ServiceDescription,
				CreatedAt: model.JSONTime{Time: e.ServiceCreatedAt},
				UpdatedAt: model.JSONTime{Time: e.ServiceUpdatedAt},
			},
		},
	}
}

func mapEntityTypeRecordToModel(e entityTypeRecord) *model.EntityType {
	return &model.EntityType{
		ID:          model.ID(e.ID),
		Name:        e.Name,
		Description: e.Description,
		ServiceID:   model.ID(e.ServiceID),
		EntitiesCnt: e.EntitiesCnt,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

func mapEntityTypesToCollection(items []entityTypeRecord, cnt int, page, perPage uint) *model.EntityTypeCollection {
	result := model.EntityTypeCollection{}
	for _, e := range items {
		result.Items = append(result.Items, *mapEntityTypeRecordToModel(e))
	}

	result.Meta.Total = cnt
	result.Meta.Page = int(page)
	result.Meta.PerPage = int(perPage)

	return &result
}

func mapActionRecordsToCollection(items []actionRecord, total int, page uint, perPage uint) *model.ActionCollection {
	result := model.ActionCollection{}
	for _, a := range items {
		result.Items = append(result.Items, *mapActionRecordToModel(a))
	}

	result.Meta.Total = total
	result.Meta.Page = int(page)
	result.Meta.PerPage = int(perPage)

	return &result
}

func mapActionRecordToModel(ar actionRecord) *model.Action {
	a := model.Action{
		ID: model.ID(ar.ID),
		Name: ar.Name,
		Hash: ar.Hash,
		EmittedAt: model.JSONTime{Time: ar.EmittedAt},
		RegisteredAt: model.JSONTime{Time: ar.RegisteredAt},
	}

	if ar.ParentEventID.Valid {
		a.ParentID = model.IDPointer(ar.ParentEventID.String)
	}

	if ar.Delta != nil {
		if err := json.Unmarshal([]byte(*ar.Delta), &a.Delta); err != nil {
			panic(err)
		}
	}

	if ar.Details != nil {
		if err := json.Unmarshal([]byte(*ar.Details), &a.Details); err != nil {
			panic(err)
		}
	}

	return &a
}

