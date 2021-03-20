package model

import (
	"github.com/denismitr/auditbase/utils/errbag"
	"github.com/denismitr/auditbase/utils/validator"
)

type Crud int

const (
	AnyAction Crud = iota
	CreateAction
	UpdateAction
	DeleteAction
)

type NewAction struct {
	ID               string      `json:"id"`
	ParentID         *string     `json:"parentId"`
	ActorExternalID  *string     `json:"actorExternalId"`
	ActorEntity      *string     `json:"actorEntity"`
	ActorService     string      `json:"actorService"`
	TargetExternalID *string     `json:"targetExternalId"`
	TargetEntity     *string     `json:"targetEntity"`
	TargetService    string      `json:"targetService"`
	Name             string      `json:"name"`
	EmittedAt        JSONTime    `json:"emittedAt"`
	RegisteredAt     JSONTime    `json:"registeredAt"`
	Status           Status      `json:"status"`
	IsAsync          bool        `json:"isAsync"`
	Details          interface{} `json:"details"`
	Delta            interface{} `json:"delta"`
	Hash             string      `json:"hash"`
}

type Action struct {
	ID             ID          `json:"id"`
	ParentID       *ID         `json:"parentId"`
	Parent         *Action     `json:"parent"`
	ChildrenCount  int         `json:"childrenCount"`
	Hash           string      `json:"hash"`
	ActorEntityID  *ID         `json:"actorEntityId"`
	Actor          *Entity     `json:"actor"`
	TargetEntityID *ID         `json:"targetId"`
	Target         *Entity     `json:"target"`
	Name           string      `json:"name"`
	Status         Status      `json:"status"`
	IsAsync        bool        `json:"isAsync"`
	EmittedAt      JSONTime    `json:"emittedAt"`
	RegisteredAt   JSONTime    `json:"registeredAt"`
	Details        interface{} `json:"details"`
	Delta          interface{} `json:"delta"`
}

type ActionCollection struct {
	Items []Action `json:"data"`
	Meta  Meta     `json:"meta"`
}

func (a *Action) Validate() *errbag.ErrorBag {
	bag := errbag.New()

	if !validator.IsEmptyString(a.ID.String()) {
		bag.Add("id", ErrMissingActionID)
	}

	if !validator.IsUUID4(a.ID.String()) {
		bag.Add("id", ErrInvalidUUID4)
	}

	return bag
}

func (na *NewAction) Validate() *errbag.ErrorBag {
	eb := errbag.New()

	if !validator.IsEmptyString(na.ID) && !validator.IsUUID4(na.ID) {
		eb.Add("id", ErrInvalidUUID4)
	}

	if na.ActorEntity != nil && validator.IsEmptyString(*na.ActorEntity) {
		eb.Add("actorEntity", ErrActorEntityEmpty)
	}

	if na.TargetEntity != nil && validator.IsEmptyString(*na.TargetEntity) {
		eb.Add("targetEntity", ErrTargetEntityEmpty)
	}

	if validator.IsEmptyString(na.ActorService) {
		eb.Add("actorService", ErrActorServiceEmpty)
	}

	if validator.IsEmptyString(na.TargetService) {
		eb.Add("targetService", ErrTargetServiceEmpty)
	}

	if na.EmittedAt.IsZero() {
		eb.Add("emittedAt", ErrEmittedAtEmpty)
	}

	return eb
}
