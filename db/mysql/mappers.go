package mysql

import (
	"github.com/denismitr/auditbase/model"
	"github.com/pkg/errors"
	"time"
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
	cat, err := time.Parse(model.DefaultTimeFormat, e.CreatedAt)
	if err != nil {
		panic(errors.Wrap(err, "how can created at be invalid?"))
	}

	uat, err := time.Parse(model.DefaultTimeFormat, e.UpdatedAt)
	if err != nil {
		panic(errors.Wrap(err, "how can updated at be invalid?"))
	}

	return &model.Entity{
		ID:           model.ID(e.ID),
		ExternalID:   e.ExternalID,
		EntityTypeID: model.ID(e.EntityTypeID),
		IsActor:      e.IsActor,
		CreatedAt:    cat,
		UpdatedAt:    uat,
	}
}

func mapEntityTypeRecordToModel(e entityTypeRecord) *model.EntityType {
	cat, err := time.Parse(model.DefaultTimeFormat, e.CreatedAt)
	if err != nil {
		panic(errors.Wrap(err, "how can created at be invalid?"))
	}

	uat, err := time.Parse(model.DefaultTimeFormat, e.UpdatedAt)
	if err != nil {
		panic(errors.Wrap(err, "how can updated at be invalid?"))
	}

	return &model.EntityType{
		ID:          model.ID(e.ID),
		Name:        e.Name,
		Description: e.Description,
		ServiceID:   model.ID(e.ServiceID),
		EntitiesCnt: e.EntitiesCnt,
		CreatedAt:   cat,
		UpdatedAt:   uat,
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

	return &a
}