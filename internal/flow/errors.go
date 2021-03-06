package flow

import "github.com/denismitr/auditbase/internal/utils/errtype"

const ErrTooManyAttempts = errtype.StringError("too many attempts")
const ErrCannotRequeueAction = errtype.StringError("could not requeue action")
