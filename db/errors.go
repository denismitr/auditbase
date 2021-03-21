package db

import "github.com/denismitr/auditbase/utils/errtype"

const ErrDBWriteFailed = errtype.StringError("db write failed")
const ErrDBReadFailed = errtype.StringError("db read failed")
const ErrUniqueConstrainedFailed = errtype.StringError("db unique constrained failed")
const ErrNotFound = errtype.StringError("requested entity or collection not found in DB")
const ErrPersisterCouldNotPrepareEvent = errtype.StringError("persister could not prepare event")
const ErrCouldNotCreateEvent = errtype.StringError("could not create event")
const ErrInvalidID = errtype.StringError("invalid database ID")
const ErrInvalidUUID4 = errtype.StringError("invalid string given instead of uuid4")
const ErrCouldNotCommit = errtype.StringError("could not commit transaction")
const ErrCouldNotBuildQuery = errtype.StringError("could not build query")
const ErrEmptyWhereInList = errtype.StringError("WHERE IN clause is empty")

const ErrActionNotFound = errtype.StringError("action not found")
