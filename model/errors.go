package model

import "github.com/denismitr/auditbase/utils/errtype"

const ErrMissingActionID = errtype.StringError("action ID is empty")
const ErrNameIsRequired = errtype.StringError("name is required")
const ErrInvalidUUID4 = errtype.StringError("not a valid UUID4")
const ErrActorIDEmpty = errtype.StringError("actorEntityID must not be empty")
const ErrMicroserviceNameTooLong = errtype.StringError("microservice name is too long")
const ErrMicroserviceDescriptionTooLong = errtype.StringError("microservice description is too long")
const ErrMicroserviceNotFound = errtype.StringError("microservice not found")

const ErrActorEntityEmpty = errtype.StringError("actorEntity must not be empty")
const ErrTargetEntityEmpty = errtype.StringError("targetEntity must not be empty")
const ErrActorServiceEmpty = errtype.StringError("actorService must not be empty")
const ErrTargetServiceEmpty = errtype.StringError("targetService must not be empty")
const ErrEmittedAtEmpty = errtype.StringError("emittedAt must not be empty")
