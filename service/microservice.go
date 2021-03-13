package service

import (
	"context"
	"github.com/denismitr/auditbase/db"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/pkg/errors"
)

type MicroserviceService interface {
	FirstByID(context.Context, model.ID, *db.Include) (*model.Microservice, error)
	Create(context.Context, *model.Microservice) (*model.Microservice, error)
	SelectAll(context.Context, *db.Include) (*model.MicroserviceCollection, error)
	Update(context.Context, model.ID, *model.Microservice) (*model.Microservice, error)
	FirstByName(ctx context.Context, name string, inc *db.Include) (*model.Microservice, error)
}

type BaseMicroserviceService struct {
	db db.Database
	lg logger.Logger
}

func NewMicroserviceService(db db.Database, lg logger.Logger) *BaseMicroserviceService {
	return &BaseMicroserviceService{
		db: db,
		lg: lg,
	}
}

func (s *BaseMicroserviceService) Create(
	ctx context.Context,
	microservice *model.Microservice,
) (*model.Microservice, error) {
	result, err := s.db.ReadWrite(ctx, func(ctx context.Context, tx db.Tx) (interface{}, error) {
		created, err := tx.Microservices().Create(ctx, microservice)
		if err != nil {
			return nil, err
		}

		return created, nil
	})

	if err != nil {
		return nil, err
	}

	if newService, ok := result.(*model.Microservice); !ok {
		panic("how result could bee of different type than a model.Microservice?")
	} else {
		return newService, nil
	}
}

func (s *BaseMicroserviceService) SelectAll(ctx context.Context, inc *db.Include) (*model.MicroserviceCollection, error) {
	result, err := s.db.ReadOnly(ctx, func(ctx context.Context, tx db.Tx) (interface{}, error) {
		microservices, err := tx.Microservices().SelectAll(ctx)
		if err != nil {
			return nil, err
		}

		// todo: join entity types

		return microservices, nil
	})

	if err != nil {
		return nil, err
	}

	if collection, ok := result.(*model.MicroserviceCollection); !ok {
		panic("how result could bee of different type than a model.MicroserviceCollection?")
	} else {
		return collection, nil
	}
}

func (s *BaseMicroserviceService) Update(ctx context.Context, id model.ID, microservice *model.Microservice) (*model.Microservice, error) {
	result, err := s.db.ReadOnly(ctx, func(ctx context.Context, tx db.Tx) (interface{}, error) {
		updated, err := tx.Microservices().Update(ctx, id, microservice)
		if err != nil {
			return nil, err
		}

		return updated, nil
	})

	if err != nil {
		return nil, err
	}

	if updated, ok := result.(*model.Microservice); !ok {
		panic("how result could bee of different type than a model.Microservice?")
	} else {
		return updated, nil
	}
}

func (s *BaseMicroserviceService) FirstByID(
	ctx context.Context,
	ID model.ID,
	inc *db.Include,
) (*model.Microservice, error) {
	result, err := s.db.ReadOnly(ctx, func(ctx context.Context, tx db.Tx) (interface{}, error) {
		result, err := tx.Microservices().FirstByID(ctx, ID)
		if err != nil {
			return nil, err
		}

		if inc.Has("entityTypes") {
			c := &db.Cursor{Page: 1, PerPage: 200}
			f := db.NewFilter([]string{"serviceId"})
			f.Add("serviceId", result.ID.String())
			entityTypes, err := tx.EntityTypes().Select(ctx, c, f)
			if err != nil {
				return nil, errors.Wrapf(err, "could not join entity types to microservice [%s]", result.Name)
			}

			result.EntityTypes = entityTypes.Items
		}

		return result, nil
	})

	if err != nil {
		return nil, err
	}

	microservice, ok := result.(*model.Microservice)
	if !ok {
		panic("how could result not be of microservice type")
	}

	return microservice, err
}

func (s *BaseMicroserviceService) FirstByName(
	ctx context.Context,
	name string,
	inc *db.Include,
) (*model.Microservice, error) {
	result, err := s.db.ReadOnly(ctx, func(ctx context.Context, tx db.Tx) (interface{}, error) {
		result, err := tx.Microservices().FirstByName(ctx, name)
		if err != nil {
			return nil, err
		}

		// todo: join entity types

		return result, nil
	})

	if err != nil {
		return nil, err
	}

	microservice, ok := result.(*model.Microservice)
	if !ok {
		panic("how could result not be of microservice type")
	}

	return microservice, err
}
