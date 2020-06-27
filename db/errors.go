package db

import "github.com/denismitr/auditbase/utils/errtype"

const ErrDBWriteFailed = errtype.StringError("db write failed")
const ErrDBReadFailed = errtype.StringError("db read failed")
const ErrUniqueConstrainedFailed = errtype.StringError("db unique constrained failed")
const ErrEntityDoesNotExist = errtype.StringError("requested entity does not exist in DB")
const ErrPersisterCouldNotPrepareEvent = errtype.StringError("persister could not prepare event")
const ErrCouldNotCreateEvent = errtype.StringError("could not create event")
const ErrEmptyUUID4 = errtype.StringError("empty string given instead of uuid4")
const ErrInvalidUUID4 = errtype.StringError("invalid string given instead of uuid4")
const ErrCouldNotCommit = errtype.StringError("could not commit transaction")
const ErrCouldNotBuildQuery = errtype.StringError("could not build query")
const ErrEmptyWhereInList = errtype.StringError("WHERE IN clause is empty")

func covertToPersistenceResult(err error) PersistenceResult {
	switch err {
	case ErrDBWriteFailed, ErrDBReadFailed:
		return CriticalDatabaseFailure
	case ErrUniqueConstrainedFailed, ErrEntityDoesNotExist:
		return LogicalError
	default:
		return UnknownError
	}
}
