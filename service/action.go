package service

import (
	"context"
	"github.com/denismitr/auditbase/db"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/logger"
)

type ActionService interface {
	Create(ctx context.Context, action *model.NewAction) (*model.Action, error)
}

type BaseActionService struct {
	db db.Database
	lg logger.Logger
}

func (s *BaseActionService) Create(ctx context.Context, newAction *model.NewAction) (*model.Action, error) {
	var result *model.Action
	if err := s.db.ReadWrite(ctx, func(ctx context.Context, tx db.Tx) error {
		actingService, err := tx.Microservices().FirstOrCreateByName(ctx, newAction.ActorService)
		if err != nil {
			return err
		}

		targetService, err := tx.Microservices().FirstOrCreateByName(ctx, newAction.TargetService)
		if err != nil {
			return err
		}

		action := new(model.Action)

		var actingEntity *model.Entity
		if newAction.ActorExternalID != nil && newAction.ActorEntity != nil {
			var err error

			actingEntityType, err := tx.EntityTypes().FirstOrCreateByNameAndServiceID(ctx, *newAction.ActorEntity, actingService.ID)
			if err != nil {
				return err
			}

			actingEntity, err = tx.Entities().FirstOrCreateByExternalIDAndEntityTypeID(ctx, *newAction.ActorExternalID, actingEntityType.ID, true)
			if err != nil {
				return err
			}

			action.ActorEntityID = &actingEntity.ID
		}

		var targetEntity *model.Entity
		if newAction.TargetExternalID != nil && newAction.TargetEntity != nil {
			var err error

			targetEntityType, err := tx.EntityTypes().FirstOrCreateByNameAndServiceID(ctx, *newAction.TargetEntity, targetService.ID)
			if err != nil {
				return err
			}

			targetEntity, err = tx.Entities().FirstOrCreateByExternalIDAndEntityTypeID(ctx, *newAction.TargetExternalID, targetEntityType.ID, false)
			if err != nil {
				return err
			}

			action.TargetEntityID = &targetEntity.ID
		}

		action.Name = newAction.Name
		action.EmittedAt = newAction.EmittedAt
		action.RegisteredAt = newAction.RegisteredAt
		action.Status = newAction.Status
		action.IsAsync = newAction.IsAsync

		result, err = tx.Actions().Create(ctx, action)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return result, nil
}


