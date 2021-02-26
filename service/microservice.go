package service

import (
	"context"
	"github.com/denismitr/auditbase/db"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/pkg/errors"
)

type MicroserviceService interface {
	FirstByID(context.Context, model.ID) (*model.Microservice, error)
}

type BaseMicroserviceService struct {
	db db.Database
	lg logger.Logger
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
