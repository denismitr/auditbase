package service

import (
	"context"
	"github.com/denismitr/auditbase/db"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/logger"
)

type EntityService interface {
	Select(ctx context.Context, f *db.Filter, c *db.Cursor) (*model.EntityCollection, error)
	FirstByID(ctx context.Context, ID model.ID) (*model.Entity, error)
}

var _ EntityService = (*BaseEntityService)(nil)

type BaseEntityService struct {
	db db.Database
	lg logger.Logger
}

func (s *BaseEntityService) Select(
	ctx context.Context,
	f *db.Filter,
	c *db.Cursor,
) (*model.EntityCollection, error) {
	result, err := s.db.ReadOnly(ctx, func(ctx context.Context, tx db.Tx) (interface{}, error) {
		panic("not implemented yet")
	})

	if err != nil {
		return nil, err
	}

	if entityCollection, ok := result.(*model.EntityCollection); !ok {
		panic("how could result not be of type model.EntityCollection")
	} else {
		return entityCollection, nil
	}
}

func (s *BaseEntityService) FirstByID(ctx context.Context, ID model.ID) (*model.Entity, error) {
	result, err := s.db.ReadOnly(ctx, func(ctx context.Context, tx db.Tx) (interface{}, error) {
		entity, err := tx.Entities().FirstByID(ctx, ID)
		if err != nil {
			return nil, err
		}

		return entity, nil
	})

	if err != nil {
		return nil, err
	}

	if entity, ok := result.(*model.Entity); !ok {
		panic("how could result not be of type model.Entity")
	} else {
		return entity, nil
	}
}

func NewEntityService(db db.Database, lg logger.Logger) *BaseEntityService {
	return &BaseEntityService{
		db: db,
		lg: lg,
	}
}
