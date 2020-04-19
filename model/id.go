package model

import (
	"github.com/denismitr/auditbase/utils/errbag"
	"github.com/denismitr/auditbase/utils/validator"
)

type ID string

func (id ID) Validate() *errbag.ErrorBag {
	eb := errbag.New()

	if !validator.IsUUID4(id.String()) {
		eb.Add("ID", ErrInvalidUUID4)
	}

	return eb
}

func (id ID) String() string {
	return string(id)
}
