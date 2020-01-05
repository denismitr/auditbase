package utils

import (
	uuid "github.com/satori/go.uuid"
)

func UUID4() string {
	u := uuid.NewV4()
	return u.String()
}
