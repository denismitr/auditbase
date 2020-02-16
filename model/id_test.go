package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIDValidation(t *testing.T) {
	invalidIDs := []struct {
		name string
		ID   string
	}{
		{name: "empty", ID: ""},
	}

	for _, tc := range invalidIDs {
		t.Run(tc.name, func(t *testing.T) {
			ID := ID(tc.ID)

			errors := ID.Validate()

			assert.False(t, errors.IsEmpty())
		})
	}
}
