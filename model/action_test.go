package model

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_ActionCanBeSerialized(t *testing.T) {
	emitted, err := time.Parse(DefaultTimeFormat, "2021-02-23 16:51:35")
	if err != nil {
		panic(err)
	}
	registered, err := time.Parse(DefaultTimeFormat, "2021-02-23 16:54:49")
	if err != nil {
		panic(err)
	}

	t.Run("all basic data", func(t *testing.T) {
		a := Action{
			ID:             ID("6d94b604-560b-4f3d-8cf0-c341faca3139"),
			ParentID:       IDToPointer("837468fd-d35b-4504-bb18-bf1903113727"),
			ActorEntityID:  IDToPointer("537468fd-d35b-4504-bb18-bf1903113725"),
			TargetEntityID: IDToPointer("355568fd-d35b-4504-bb18-bf1903113701"),
			Name:           "foo-bar-2",
			Hash:           "foo-hash-2",
			IsAsync:        true,
			Status:         Failed,
			EmittedAt:      JSONTime{Time: emitted},
			RegisteredAt:   JSONTime{Time: registered},
			Details: map[string]interface{}{"foo": 123, "bar": "baz"},
			Delta: map[string]interface{}{"foo": []string{"a", "b"}, "bar": []int{1, 2}},
		}

		b, err := json.Marshal(&a)
		if err != nil {
			panic(err)
		}

		expected := `{"id":"6d94b604-560b-4f3d-8cf0-c341faca3139","parentId":"837468fd-d35b-4504-bb18-bf1903113727","childrenCount":0,"hash":"foo-hash-2","actorEntityId":"537468fd-d35b-4504-bb18-bf1903113725","actor":null,"targetId":"355568fd-d35b-4504-bb18-bf1903113701","target":null,"name":"foo-bar-2","status":6,"isAsync":true,"emittedAt":"2021-02-23 16:51:35","registeredAt":"2021-02-23 16:54:49","details":{"bar":"baz","foo":123},"delta":{"bar":[1,2],"foo":["a","b"]}}`

		assert.Equal(t, expected, string(b))
	})
}
