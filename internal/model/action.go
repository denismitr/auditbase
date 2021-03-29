package model

import (
	"github.com/denismitr/auditbase/internal/utils/validator"
	"time"
)

type UpdateAction struct {
	ID           int       `json:"id"`
	UID          string    `json:"uid"`
	Hash         string    `json:"hash"`
	RegisteredAt time.Time `json:"registeredAt"`
	Status       Status    `json:"status"`
}

func (ua UpdateAction) Validate() *validator.ValidationErrors {
	eb := validator.NewValidationError()

	if ua.UID != "" && len(ua.UID) != 32 {
		eb.Add("uid", ErrInvalidUID)
	}

	return eb
}

type NewAction struct {
	UID              string      `json:"uid"`
	ParentUID        string      `json:"parentUid"`
	ActorExternalID  string      `json:"actorExternalId"`
	ActorEntity      string      `json:"actorEntity"`
	ActorService     string      `json:"actorService"`
	TargetExternalID string      `json:"targetExternalId"`
	TargetEntity     string      `json:"targetEntity"`
	TargetService    string      `json:"targetService"`
	Name             string      `json:"name"`
	EmittedAt        JSONTime    `json:"emittedAt"`
	RegisteredAt     time.Time   `json:"registeredAt"`
	Status           Status      `json:"status"`
	IsAsync          bool        `json:"isAsync"`
	Details          interface{} `json:"details"`
	Hash             string      `json:"hash"`
}

type Action struct {
	ID             ID          `json:"id"`
	UID            UID         `json:"uid"`
	ParentUID      UID         `json:"parentUid"`
	Parent         *Action     `json:"parent,omitempty"`
	ChildrenCount  int         `json:"childrenCount"`
	Hash           string      `json:"hash"`
	ActorEntityID  ID          `json:"actorEntityId"`
	Actor          *Entity     `json:"actor,omitempty"`
	TargetEntityID ID          `json:"targetId"`
	Target         *Entity     `json:"target,omitempty"`
	Name           string      `json:"name"`
	Status         Status      `json:"status"`
	IsAsync        bool        `json:"isAsync"`
	EmittedAt      JSONTime    `json:"emittedAt"`
	RegisteredAt   JSONTime    `json:"registeredAt"`
	Details        interface{} `json:"details"`
}

type ActionCollection struct {
	Items []Action `json:"data"`
	Meta  Meta     `json:"meta"`
}

func (na *NewAction) Validate() *validator.ValidationErrors {
	eb := validator.NewValidationError()

	if na.ActorEntity != "" && validator.IsEmptyString(na.ActorEntity) {
		eb.Add("actorEntity", ErrActorEntityEmpty)
	}

	if na.TargetEntity != "" && validator.IsEmptyString(na.TargetEntity) {
		eb.Add("targetEntity", ErrTargetEntityEmpty)
	}

	if validator.IsEmptyString(na.ActorService) {
		eb.Add("actorService", ErrActorServiceEmpty)
	}

	if validator.IsEmptyString(na.TargetService) {
		eb.Add("targetService", ErrTargetServiceEmpty)
	}

	if na.ParentUID != "" && len(na.ParentUID) != 32 {
		eb.Add("parentUid", ErrInvalidUID)
	}

	if na.UID != "" && len(na.UID) != 32 {
		eb.Add("uid", ErrInvalidUID)
	}

	if na.EmittedAt.IsZero() {
		eb.Add("emittedAt", ErrEmittedAtEmpty)
	}

	return eb
}
