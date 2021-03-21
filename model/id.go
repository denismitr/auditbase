package model

import (
	"github.com/denismitr/auditbase/utils/errbag"
)

type ID uint64

func (id ID) Validate() *errbag.ErrorBag {
	eb := errbag.New()

	return eb
}