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
				ID: model.ID("7c94b604-560b-4f3d-8cf0-c341faca3139"),
				ParentID: model.IDToPointer("937468fd-d35b-4504-bb18-bf1903113727"),
				Name: "foo-bar",
				Hash: "foo-hash",
				IsAsync: false,
				Status: model.Processing,
				EmittedAt: model.JSONTime{Time: emitted},
				RegisteredAt: model.JSONTime{Time: registered},
			},
			expected: "INSERT INTO `actions` (`action_entity_id`, `emitted_at`, `hash`, `id`, `is_async`, `name`, `parent_id`, `registered_at`, `status`, `target_entity_id`) VALUES (NULL, '2021-02-23 16:51:35', 'foo-hash', uuid_to_bin('7c94b604-560b-4f3d-8cf0-c341faca3139'), 0, 'foo-bar', uuid_to_bin('937468fd-d35b-4504-bb18-bf1903113727'), '2021-02-23 16:54:49', 2, NULL)",
		},
		{
			action: &model.Action{
				ID: model.ID("6c94b604-560b-4f3d-8cf0-c341faca3139"),
				ParentID: model.IDToPointer("837468fd-d35b-4504-bb18-bf1903113727"),
				ActorEntityID: model.IDToPointer("537468fd-d35b-4504-bb18-bf1903113725"),
				Name: "foo-bar-2",
				Hash: "foo-hash-2",
				IsAsync: false,
				Status: model.Dynamic,
				EmittedAt: model.JSONTime{Time: emitted},
				RegisteredAt: model.JSONTime{Time: registered},
			},
			expected: "INSERT INTO `actions` (`action_entity_id`, `emitted_at`, `hash`, `id`, `is_async`, `name`, `parent_id`, `registered_at`, `status`, `target_entity_id`) VALUES (uuid_to_bin('537468fd-d35b-4504-bb18-bf1903113725'), '2021-02-23 16:51:35', 'foo-hash-2', uuid_to_bin('6c94b604-560b-4f3d-8cf0-c341faca3139'), 0, 'foo-bar-2', uuid_to_bin('837468fd-d35b-4504-bb18-bf1903113727'), '2021-02-23 16:54:49', 0, NULL)",
		},
		{
			action: &model.Action{
				ID: model.ID("6d94b604-560b-4f3d-8cf0-c341faca3139"),
				ActorEntityID: model.IDToPointer("537468fd-d35b-4504-bb18-bf1903113725"),
				TargetEntityID: model.IDToPointer("355568fd-d35b-4504-bb18-bf1903113701"),
				Name: "foo-bar-2",
				Hash: "foo-hash-2",
				IsAsync: true,
				Status: model.Failed,
				EmittedAt: model.JSONTime{Time: emitted},
				RegisteredAt: model.JSONTime{Time: registered},
			},
			expected: "INSERT INTO `actions` (`action_entity_id`, `emitted_at`, `hash`, `id`, `is_async`, `name`, `parent_id`, `registered_at`, `status`, `target_entity_id`) VALUES (uuid_to_bin('537468fd-d35b-4504-bb18-bf1903113725'), '2021-02-23 16:51:35', 'foo-hash-2', uuid_to_bin('6d94b604-560b-4f3d-8cf0-c341faca3139'), 1, 'foo-bar-2', NULL, '2021-02-23 16:54:49', 6, uuid_to_bin('355568fd-d35b-4504-bb18-bf1903113701'))",
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
