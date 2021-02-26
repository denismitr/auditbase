package service

import (
	"context"
	"github.com/denismitr/auditbase/db"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/pkg/errors"
)

type ActionService interface {
	Create(ctx context.Context, action *model.NewAction) (*model.Action, error)
	FirstByID(ctx context.Context, ID model.ID) (*model.Action, error)
}

type BaseActionService struct {
	db db.Database
	lg logger.Logger
}

func NewActionService(db db.Database, lg logger.Logger) *BaseActionService {
	return &BaseActionService{
		db: db,
		lg: lg,
	}
}

func (s *BaseActionService) FirstByID(ctx context.Context, ID model.ID) (*model.Action, error) {
	result, err := s.db.ReadWrite(ctx, func(ctx context.Context, tx db.Tx) (interface{}, error) {
		actions := tx.Actions()
		action, err := actions.FirstByID(ctx, ID)
		if err != nil {
			return nil, err
		}

		if action.ParentID != nil {
			parent, err := actions.FirstByID(ctx, *action.ParentID)
			if err != nil {
				return nil, errors.Wrap(err, "could not get parent ID")
			}

			action.Parent = parent
		}

		entities := tx.Entities()
		if action.ActorEntityID != nil {
			actor, err := entities.FirstByIDWithEntityType(ctx, *action.ActorEntityID)
			if err != nil {
				return nil, errors.Wrap(err, "could not join actor to action")
			}

			action.Actor = actor
		}

		if action.TargetEntityID != nil {
			target, err := entities.FirstByIDWithEntityType(ctx, *action.TargetEntityID)
			if err != nil {
				return nil, errors.Wrap(err, "could not join target to action")
			}

			action.Target = target
		}

		return action, nil
	})

	if err != nil {
		return nil, err
	}

	action, ok := result.(*model.Action);
	if !ok {
		panic("how result could have of different type than Action?")
	}

	return action, nil
}

func (s *BaseActionService) Create(ctx context.Context, newAction *model.NewAction) (*model.Action, error) {
	result, err := s.db.ReadWrite(ctx, func(ctx context.Context, tx db.Tx) (interface{}, error) {
		actingService, err := tx.Microservices().FirstOrCreateByName(ctx, newAction.ActorService)
		if err != nil {
			return nil, err
		}

		targetService, err := tx.Microservices().FirstOrCreateByName(ctx, newAction.TargetService)
		if err != nil {
			return nil, err
		}

		action := new(model.Action)

		var actingEntity *model.Entity
		if newAction.ActorExternalID != nil && newAction.ActorEntity != nil {
			var err error

			actingEntityType, err := tx.EntityTypes().FirstOrCreateByNameAndServiceID(ctx, *newAction.ActorEntity, actingService.ID)
			if err != nil {
				return nil, err
			}

			actingEntity, err = tx.Entities().FirstOrCreateByExternalIDAndEntityTypeID(ctx, *newAction.ActorExternalID, actingEntityType.ID, true)
			if err != nil {
				return nil, err
			}

			action.ActorEntityID = &actingEntity.ID
		}

		var targetEntity *model.Entity
		if newAction.TargetExternalID != nil && newAction.TargetEntity != nil {
			var err error

			targetEntityType, err := tx.EntityTypes().FirstOrCreateByNameAndServiceID(ctx, *newAction.TargetEntity, targetService.ID)
			if err != nil {
				return nil, err
			}

			targetEntity, err = tx.Entities().FirstOrCreateByExternalIDAndEntityTypeID(ctx, *newAction.TargetExternalID, targetEntityType.ID, false)
			if err != nil {
				return nil, err
			}

			action.TargetEntityID = &targetEntity.ID
		}

		action.Name = newAction.Name
		action.EmittedAt = newAction.EmittedAt
		action.RegisteredAt = newAction.RegisteredAt
		action.Status = newAction.Status
		action.IsAsync = newAction.IsAsync

		action, err = tx.Actions().Create(ctx, action)
		if err != nil {
			return nil, err
		}

		return action, nil
	})

	if err != nil {
		return nil, err
	}

	action, ok := result.(*model.Action)
	if !ok {
		panic("how result could have of different type than Action?")
	}

	return action, nil
}


