package rest

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInterfaceToPointer(t *testing.T) {
	tt := []struct {
		name  string
		in    interface{}
		out   string
		isNil bool
	}{
		{
			name:  "string",
			in:    "foo bar baz",
			out:   "foo bar baz",
			isNil: false,
		},
		{
			name:  "integer",
			in:    3456,
			out:   "3456",
			isNil: false,
		},
		{
			name:  "nil",
			in:    nil,
			out:   "",
			isNil: true,
		},
		{
			name:  "float",
			in:    456.98,
			out:   "456.98",
			isNil: false,
		},
		{
			name:  "boolean-false",
			in:    false,
			out:   "0",
			isNil: false,
		},
		{
			name:  "boolean-true",
			in:    true,
			out:   "1",
			isNil: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			out := interfaceToStringPointer(tc.in)
			if tc.isNil {
				assert.Nil(t, out)
			} else if out == nil {
				assert.NotNil(t, out, "out should not have been nil")
			} else {
				assert.Equal(t, *out, tc.out)
			}
		})
	}
}
