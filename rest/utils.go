package rest

import (
	uuid "github.com/satori/go.uuid"
)

func uuid4() string {
	u := uuid.NewV4()
	return u.String()
}
