package mysql

import (
	"github.com/denismitr/auditbase/model"
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
				ID: 10,
				ParentID: 123,
				Name: "foo-bar",
				Hash: "foo-hash",
				IsAsync: false,
				Status: model.Processing,
				EmittedAt: model.JSONTime{Time: emitted},
				RegisteredAt: model.JSONTime{Time: registered},
			},
			expected: "INSERT INTO `actions` (`actor_entity_id`, `emitted_at`, `hash`, `is_async`, `name`, `parent_id`, `registered_at`, `status`, `target_entity_id`) VALUES (NULL, '2021-02-23 16:51:35', 'foo-hash', 0, 'foo-bar', 123, '2021-02-23 16:54:49', 2, NULL)",
		},
		{
			action: &model.Action{
				ID: 34,
				ParentID: 56,
				ActorEntityID: 57,
				Name: "foo-bar-2",
				Hash: "foo-hash-2",
				IsAsync: false,
				Status: model.Dynamic,
				EmittedAt: model.JSONTime{Time: emitted},
				RegisteredAt: model.JSONTime{Time: registered},
			},
			expected: "INSERT INTO `actions` (`actor_entity_id`, `emitted_at`, `hash`, `is_async`, `name`, `parent_id`, `registered_at`, `status`, `target_entity_id`) VALUES (57, '2021-02-23 16:51:35', 'foo-hash-2', 0, 'foo-bar-2', 56, '2021-02-23 16:54:49', 0, NULL)",
		},
		{
			action: &model.Action{
				ID: 10,
				ActorEntityID: 11,
				TargetEntityID: 12,
				Name: "foo-bar-2",
				Hash: "foo-hash-2",
				IsAsync: true,
				Status: model.Failed,
				EmittedAt: model.JSONTime{Time: emitted},
				RegisteredAt: model.JSONTime{Time: registered},
			},
			expected: "INSERT INTO `actions` (`actor_entity_id`, `emitted_at`, `hash`, `is_async`, `name`, `parent_id`, `registered_at`, `status`, `target_entity_id`) VALUES (11, '2021-02-23 16:51:35', 'foo-hash-2', 1, 'foo-bar-2', NULL, '2021-02-23 16:54:49', 6, 12)",
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
