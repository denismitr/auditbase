package types

import "github.com/denismitr/auditbase/model"

func PointerToString(s string) *string {
	return &s
}

func IDToPointer(id model.ID) *model.ID {
	return &id
}
