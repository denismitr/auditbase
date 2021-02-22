package db

import (
	"context"
	"github.com/denismitr/auditbase/model"
)

type Tx interface {
	Entities() EntityRepository
}

type Database interface {
	ReadOnly(ctx context.Context, cb func(context.Context, *Tx) error) error
	ReadWrite(ctx context.Context, cb func(context.Context, *Tx) error) error
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