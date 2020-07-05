package rest

import (
	"github.com/denismitr/auditbase/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDataType(t *testing.T) {
	tt := []struct {
		name string
		from *string
		to   *string
		dt   model.DataType
	}{
		{
			name: "string_to_int_as_string",
			from: interfaceToStringPointer("foo bar"),
			to:   interfaceToStringPointer("1234"),
			dt:   model.StringDataType,
		},
		{
			name: "string_to_int_as_string",
			from: interfaceToStringPointer("foo bar"),
			to:   interfaceToStringPointer(1234),
			dt:   model.StringDataType,
		},
		{
			name: "int->int",
			from: interfaceToStringPointer(4567),
			to:   interfaceToStringPointer(1234),
			dt:   model.IntegerDataType,
		},
		{
			name: "float->int",
			from: interfaceToStringPointer(4567.984),
			to:   interfaceToStringPointer(1234),
			dt:   model.FloatDataType,
		},
		{
			name: "int->float",
			from: interfaceToStringPointer(9384),
			to:   interfaceToStringPointer(1234.7645),
			dt:   model.FloatDataType,
		},
		{
			name: "nil->string",
			from: interfaceToStringPointer(nil),
			to:   interfaceToStringPointer("foooo"),
			dt:   model.StringDataType,
		},
		{
			name: "nil->integer",
			from: interfaceToStringPointer(nil),
			to:   interfaceToStringPointer(1234),
			dt:   model.IntegerDataType,
		},
		{
			name: "nil->float",
			from: interfaceToStringPointer(nil),
			to:   interfaceToStringPointer(1234.874),
			dt:   model.FloatDataType,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			dt := guessPairDataType(tc.from, tc.to)
			assert.Equal(t, tc.dt, dt)
		})
	}
}

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
