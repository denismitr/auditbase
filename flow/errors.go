package flow

import "github.com/denismitr/auditbase/utils/errtype"

const ErrTooManyAttempts = errtype.StringError("too many attempts")
const ErrCannotRequeueEvent = errtype.StringError("could not requeue event")
