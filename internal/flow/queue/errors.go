package queue

import "github.com/denismitr/auditbase/internal/utils/errtype"

const ErrNoAttemptInfo = errtype.StringError("no attempt information found in queue message")
const ErrMalformedAttemptInfo = errtype.StringError("malformed attempt information in queue message")
const ErrNoActionType = errtype.StringError("no action type")
