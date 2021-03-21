package service

import (
	"context"
	"fmt"
	"github.com/denismitr/auditbase/db"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/pkg/errors"
)

type ActionService interface {
	Select(context.Context, *db.Cursor, *db.Filter) (*model.ActionCollection, error)
	Create(context.Context, *model.NewAction) (*model.Action, error)
	FirstByID(context.Context, model.ID) (*model.Action, error)
	Count(ctx context.Context) (int, error)
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

func (s *BaseActionService) Select(ctx context.Context, c *db.Cursor, f *db.Filter) (*model.ActionCollection, error) {
	result, err := s.db.ReadOnly(ctx, func(ctx context.Context, tx db.Tx) (interface{}, error) {
		actions, err := tx.Actions().Select(ctx, c, f)
		if err != nil {
			return nil, err
		}

		// TODO: join entities

		return actions, nil
	})

	if err != nil {
		return nil, err
	}

	if actions, ok := result.(*model.ActionCollection); !ok {
		panic("how could result not be of type *model.ActionCollection")
	} else {
		return actions, nil
	}
}

func (s *BaseActionService) Count(ctx context.Context) (int, error) {
	result, err := s.db.ReadOnly(ctx, func(ctx context.Context, tx db.Tx) (interface{}, error) {
		count, err := tx.Actions().CountAll(ctx)
		if err != nil {
			return 0, err
		}

		return count, nil
	})

	if err != nil {
		return 0, err
	}

	if count, ok := result.(int); !ok {
		panic("how could count not be of type int")
	} else {
		return count, nil
	}
}

func (s *BaseActionService) FirstByID(ctx context.Context, ID model.ID) (*model.Action, error) {
	result, err := s.db.ReadWrite(ctx, func(ctx context.Context, tx db.Tx) (interface{}, error) {
		actions := tx.Actions()
		action, err := actions.FirstByID(ctx, ID)
		if err != nil {
			return nil, err
		}

		if action.ParentID != 0 {
			parent, err := actions.FirstByID(ctx, action.ParentID)
			if err != nil {
				return nil, errors.Wrap(err, "could not get parent ID")
			}

			action.Parent = parent
		}

		entities := tx.Entities()
		if action.ActorEntityID != 0 {
			actor, err := entities.FirstByIDWithEntityType(ctx, action.ActorEntityID)
			if err != nil {
				return nil, errors.Wrap(err, "could not join actor to action")
			}

			action.Actor = actor
		}

		if action.TargetEntityID != 0 {
			target, err := entities.FirstByIDWithEntityType(ctx, action.TargetEntityID)
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
		if newAction.ActorExternalID != "" && newAction.ActorEntity != "" {
			var err error

			actingEntityType, err := tx.EntityTypes().FirstOrCreateByNameAndServiceID(ctx, newAction.ActorEntity, actingService.ID)
			if err != nil {
				return nil, err
			}

			actingEntity, err = tx.Entities().FirstOrCreateByExternalIDAndEntityTypeID(ctx, newAction.ActorExternalID, actingEntityType.ID)
			if err != nil {
				return nil, err
			}

			action.ActorEntityID = actingEntity.ID
		}

		var targetEntity *model.Entity
		if newAction.TargetExternalID != "" && newAction.TargetEntity != "" {
			var err error

			targetEntityType, err := tx.EntityTypes().FirstOrCreateByNameAndServiceID(ctx, newAction.TargetEntity, targetService.ID)
			if err != nil {
				return nil, err
			}

			targetEntity, err = tx.Entities().FirstOrCreateByExternalIDAndEntityTypeID(ctx, newAction.TargetExternalID, targetEntityType.ID)
			if err != nil {
				return nil, err
			}

			action.TargetEntityID = targetEntity.ID
		}

		action.Name = newAction.Name
		action.EmittedAt = newAction.EmittedAt
		action.RegisteredAt = newAction.RegisteredAt
		action.Status = newAction.Status
		action.IsAsync = newAction.IsAsync
		action.Details = newAction.Details

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
		panic(fmt.Sprintf("how result could have of different type than Action? %#v", result))
	}

	return action, nil
}


