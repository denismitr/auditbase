package mysql

import (
	"github.com/denismitr/auditbase/internal/db"
	"github.com/denismitr/auditbase/internal/model"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_createActionQuery(t *testing.T) {
	emitted, err := time.Parse(model.DefaultTimeFormat, "2021-02-23 16:51:35")
	if err != nil {
		panic(err)
	}
	registered, err := time.Parse(model.DefaultTimeFormat, "2021-02-23 16:54:49")
	if err != nil {
		panic(err)
	}

	tt := []struct{
		action *model.Action
		expected string
		args []interface{}
	}{
		{
			action: &model.Action{
				ParentUID:    "69502edbf207452eae7ec258271ee98c",
				UID:          "23a02edbf207452eae7ec258271ee92d",
				Name:         "foo-bar",
				Hash:         "foo-hash",
				IsAsync:      false,
				Status:       model.Processing,
				EmittedAt:    model.JSONTime{Time: emitted},
				RegisteredAt: model.JSONTime{Time: registered},
			},
			expected: "INSERT INTO `actions` (`actor_entity_id`, `emitted_at`, `hash`, `is_async`, `name`, `parent_uid`, `registered_at`, `status`, `target_entity_id`, `uid`) VALUES (?, ?, UNHEX(?), ?, ?, ?, ?, ?, ?, ?)",
		},
		{
			action: &model.Action{
				UID:           "23a02edbf207452eae7ec258271ee92d",
				ParentUID:     "55502edbf207452eae7ec258271ee11e",
				ActorEntityID: 57,
				Name:          "foo-bar-2",
				Hash:          "foo-hash-2",
				IsAsync:       false,
				Status:        model.Dynamic,
				EmittedAt:     model.JSONTime{Time: emitted},
				RegisteredAt:  model.JSONTime{Time: registered},
			},
			expected: "INSERT INTO `actions` (`actor_entity_id`, `emitted_at`, `hash`, `is_async`, `name`, `parent_uid`, `registered_at`, `status`, `target_entity_id`, `uid`) VALUES (?, ?, UNHEX(?), ?, ?, ?, ?, ?, ?, ?)",
		},
		{
			action: &model.Action{
				UID:            "23a02edbf207452eae7ec258271ee92d",
				ActorEntityID:  11,
				TargetEntityID: 12,
				Name:           "foo-bar-3",
				Hash:           "foo-hash-3",
				IsAsync:        true,
				Status:         model.Failed,
				EmittedAt:      model.JSONTime{Time: emitted},
				RegisteredAt:   model.JSONTime{Time: registered},
			},
			expected: "INSERT INTO `actions` (`actor_entity_id`, `emitted_at`, `hash`, `is_async`, `name`, `parent_uid`, `registered_at`, `status`, `target_entity_id`, `uid`) VALUES (?, ?, UNHEX(?), ?, ?, ?, ?, ?, ?, ?)",
		},
	}

	for _, tc := range tt {
		t.Run(tc.action.Name, func(t *testing.T) {
			q, args, err := createActionQuery(tc.action)

			assert.NoError(t, err)
			assert.NotNil(t, args) // fixme
			assert.Equal(t, tc.expected, q)
		})
	}
}

func Test_selectActionsQuery(t *testing.T) {
	t.Run("no filters", func(t *testing.T) {
		t.Log("testing select actions query with no filters")

		c := db.NewCursor(1, 100, nil, []string{"id"})
		f := db.NewFilter([]string{})
		selectQuery, err := selectActionsQuery(c, f)
		if ! assert.NoError(t, err) {
			t.Fatal(err)
		}

		expectedSelectSQL := "SELECT `id`, `uid`, `name`, HEX(`hash`) AS `hash`, `parent_uid`, `actor_entity_id`, `target_entity_id`, `is_async`, `status`, `emitted_at`, `registered_at` FROM `actions` ORDER BY `registered_at` DESC LIMIT 100"
		assert.Equal(t, expectedSelectSQL, selectQuery.selectSQL)
		assert.Len(t, selectQuery.selectArgs, 0)

		expectedCountSQL := "SELECT count(*) AS `cnt` FROM `actions`"
		assert.Equal(t, expectedCountSQL, selectQuery.countSQL)
		assert.Len(t, selectQuery.countArgs, 0)
	})
}
