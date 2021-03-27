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
			ID:             10,
			UID:            "76502edbf207452eae7ec258271ee9aa",
			ParentUID:      "69502edbf207452eae7ec258271ee98c",
			ActorEntityID:  45,
			TargetEntityID: 456,
			Name:           "foo-bar-2",
			Hash:           "foo-hash-2",
			IsAsync:        true,
			Status:         Failed,
			EmittedAt:      JSONTime{Time: emitted},
			RegisteredAt:   JSONTime{Time: registered},
			Details:        map[string]interface{}{"foo": 123, "bar": "baz"},
			Delta:          map[string]interface{}{"foo": []string{"a", "b"}, "bar": []int{1, 2}},
		}

		b, err := json.Marshal(&a)
		if err != nil {
			panic(err)
		}

		expected := `{"id":10,"uid":"76502edbf207452eae7ec258271ee9aa","parentUid":"69502edbf207452eae7ec258271ee98c","parent":null,"childrenCount":0,"hash":"foo-hash-2","actorEntityId":45,"actor":null,"targetId":456,"target":null,"name":"foo-bar-2","status":6,"isAsync":true,"emittedAt":"2021-02-23 16:51:35","registeredAt":"2021-02-23 16:54:49","details":{"bar":"baz","foo":123},"delta":{"bar":[1,2],"foo":["a","b"]}}`

		assert.Equal(t, expected, string(b))
	})
}
