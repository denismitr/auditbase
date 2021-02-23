package model

import "github.com/denismitr/auditbase/utils/errtype"

const ErrIDEmpty = errtype.StringError("id cannot be empty")
const ErrMissingEventID = errtype.StringError("event ID is empty")
const ErrNameIsRequired = errtype.StringError("name is required")
const ErrInvalidUUID4 = errtype.StringError("not a valid UUID4")
const ErrActorIDEmpty = errtype.StringError("ActorEntityID must not be empty")
const ErrTargetIDEmpty = errtype.StringError("TargetEntityID must not be empty")
const ErrActorEntityNameEmpty = errtype.StringError("ActorEntity name must not be empty")
const ErrTargetEntityNameEmpty = errtype.StringError("TargetEntity name must not be empty")
const ErrActorServiceNameEmpty = errtype.StringError("ActorService name must not be empty")
const ErrTargetServiceNameEmpty = errtype.StringError("TargetService name must not be empty")
const ErrMicroserviceNameTooLong = errtype.StringError("microservice name is too long")
const ErrMicroserviceDescriptionTooLong = errtype.StringError("microservice description is too long")
const ErrMicroserviceNotFound = errtype.StringError("microservice not found")
const ErrEntityNotFound = errtype.StringError("entity not found")
const ErrChangeNotFound = errtype.StringError("change not found")
