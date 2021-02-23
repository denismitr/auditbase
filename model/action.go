package model

import (
	"github.com/denismitr/auditbase/utils/errbag"
	"github.com/denismitr/auditbase/utils/validator"
)

type Crud int

const (
	Unknown Crud = iota
	Create
	Update
	Delete
)

type NewAction struct {
	ID               string      `json:"id"`
	ParentID         *string     `json:"parentId"`
	ActorID          *ID         `json:"actorId"`
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
}

type Action struct {
	ID             ID          `json:"id"`
	ParentID       *ID         `json:"parentId"`
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
		bag.Add("id", ErrMissingEventID)
	}

	if !validator.IsUUID4(a.ID.String()) {
		bag.Add("id", ErrInvalidUUID4)
	}

	return bag
}
