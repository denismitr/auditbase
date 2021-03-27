package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIDValidation(t *testing.T) {
	invalidIDs := []struct {
		name string
		valid bool
		ID   int
	}{
		{name: "empty", valid: false, ID: 0},
		{name: "non-empty", valid: true, ID: 4},
	}

	for _, tc := range invalidIDs {
		t.Run(tc.name, func(t *testing.T) {
			ID := ID(tc.ID)

			assert.Equal(t,tc.valid, ID.Valid())
		})
	}
}
