package db

import (
	"context"
	"github.com/denismitr/auditbase/model"
)

type Tx interface {
	Entities() EntityRepository
	EntityTypes() EntityTypeRepository
	Actions() ActionRepository
	Microservices() MicroserviceRepository
}

type TxCallback func(context.Context, Tx) (interface{}, error)

type Database interface {
	ReadOnly(context.Context, TxCallback) (interface{}, error)
	ReadWrite(context.Context, TxCallback) (interface{}, error)
}

// EntityRepository provides entities data interactions
type EntityRepository interface {
	Select(
		ctx context.Context,
		cursor *Cursor,
		filter *Filter,
	) (*model.EntityCollection, error)

	Create(ctx context.Context, e *model.Entity) (*model.Entity, error)

	FirstOrCreateByExternalIDAndEntityTypeID(
		ctx context.Context,
		externalID string,
		entityTypeID model.ID,
		isActor bool,
	) (*model.Entity, error)

	FirstByID(ctx context.Context, ID model.ID) (*model.Entity, error)

	FirstByIDWithEntityType(ctx context.Context, ID model.ID) (*model.Entity, error)
}

// EntityTypeRepository provides entity types data interactions
type EntityTypeRepository interface {
	Select(
		ctx context.Context,
		cursor *Cursor,
		filter *Filter,
	) (*model.EntityTypeCollection, error)

	Create(ctx context.Context, e *model.EntityType) (*model.EntityType, error)

	FirstOrCreateByNameAndServiceID(
		ctx context.Context,
		name string,
		serviceID model.ID,
	) (*model.EntityType, error)

	FirstByID(ctx context.Context, ID model.ID) (*model.EntityType, error)

	FirstByNameAndServiceID(
		ctx context.Context,
		name string,
		serviceID model.ID,
	) (*model.EntityType, error)
}

type MicroserviceRepository interface {
	Create(ctx context.Context, m *model.Microservice) (*model.Microservice, error)
	Delete(context.Context, model.ID) error
	Update(context.Context, model.ID, *model.Microservice) error
	FirstByID(context.Context, model.ID) (*model.Microservice, error)
	FirstByName(ctx context.Context, name string) (*model.Microservice, error)
	FirstOrCreateByName(ctx context.Context, name string) (*model.Microservice, error)
}

type ActionRepository interface {
	Create(context.Context, *model.Action) (*model.Action, error)
	Names(context.Context) ([]string, error)
	Delete(context.Context, model.ID) error
	FirstByID(context.Context, model.ID) (*model.Action, error)
	Select(context.Context, *Cursor, *Filter) (*model.ActionCollection, error)
}