package mysql

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateSelectChangesByIDsQuery(t *testing.T) {
	tt := []struct{
		ids []string
		sql string
		args []interface{}
		err error
	}{
		{
			ids: []string{"34e1d82a-a065-436d-afd0-5fbcb752a4e1", "55e1d82a-a065-436d-bfd0-5fbcb752a4f2"},
			sql: "SELECT BIN_TO_UUID(c.id) as id, BIN_TO_UUID(c.property_id) as property_id, BIN_TO_UUID(c.event_id) as event_id, BIN_TO_UUID(p.entity_id) as entity_id, p.name as property_name, from_value, to_value FROM changes as c JOIN properties as p ON p.id = c.property_id WHERE event_id IN (UUID_TO_BIN(?),UUID_TO_BIN(?))",
			args: []interface{}{"34e1d82a-a065-436d-afd0-5fbcb752a4e1", "55e1d82a-a065-436d-bfd0-5fbcb752a4f2"},
			err: nil,
		},
		{
			ids: []string{},
			sql: "",
			args: nil,
			err: ErrEmptyWhereInList,
		},
	}

	for _, tc := range tt {
		t.Run(fmt.Sprintf("%#v", tc.ids), func(t *testing.T) {
			sql, args, err := createSelectChangesByIDsQuery(tc.ids)

			if tc.err == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, tc.err, err)
			}

			assert.Equal(t, tc.sql, sql)
			assert.Equal(t, tc.args, args)
		})
	}
}
