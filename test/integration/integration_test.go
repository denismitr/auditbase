package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/denismitr/auditbase/model"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/denismitr/auditbase/rest"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestCreateEventWithID(t *testing.T) {
	tt := []rest.CreateEvent{
		rest.CreateEvent{
			ID: "15ee8662-90a1-4a1c-85bf-ef53f1eaaa29",
			EventName:     "foo",
			TargetID:      "1234",
			TargetEntity:  "article",
			TargetService: "article-storage",
			ActorID:       "4321",
			ActorEntity:   "editor",
			ActorService:  "back-office",
			EmittedAt:     int64(1578173213),
			RegisteredAt:  int64(1578173214),
			Delta:         map[string][]interface{}{"name": []interface{}{"PENDING", "PUBLISHED"}},
		},

		rest.CreateEvent{
			ID: "15ee8662-90a1-4f1c-89bf-ef53f1eaaa29",
			EventName:     "bar",
			TargetID:      "938-UE",
			TargetEntity:  "post",
			TargetService: "blog",
			ActorID:       "999",
			ActorEntity:   "writer",
			ActorService:  "user-service",
			EmittedAt:     int64(1578178213),
			RegisteredAt:  int64(1578178314),
			Delta:         map[string][]interface{}{
				"text": []interface{}{"FOO", "BAR"},
				"status": []interface{}{nil, "published"},
			},
		},
	}

	for i, tc := range tt {
		name := fmt.Sprintf("%s_%d", tc.EventName, i)
		t.Run(name, func(t *testing.T) {


			var wg sync.WaitGroup

			wg.Add(1)

			go func(ce rest.CreateEvent) {
				defer wg.Done()

				id, err := createEvent(ce)
				if err != nil {
					t.Error(err)
					return
				}

				wg.Add(1)
				go func() {
					defer wg.Done()
					tick := time.After(1 * time.Second)

					select {
					case <-tick:
						js, err := requestEventJSONByID(id)
						if err != nil {
							t.Error(err)
							return
						}

						assert.Equal(t, tc.EventName, gjson.Get(js, "data.attributes.eventName").String())
					}
				}()
			}(tc)

			wg.Wait()
		})
	}
}

func createEvent(e rest.CreateEvent) (model.ID, error) {
	b, err := json.Marshal(e)
	if err != nil {
		return "", errors.Wrap(err, "could not marshal event")
	}

	reader := bytes.NewReader(b)

	r, err := http.Post("http://localhost:8888/api/v1/events", "application/json", reader);
	if err != nil {
		return "", errors.Wrapf(err, "event creation failed")
	}

	if r.StatusCode != 202 {
		return "", errors.Errorf("could not create an event: [%d] code received", r.StatusCode)
	}

	js, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", errors.New("could not parse accepted response")
	}

	id := gjson.Get(string(js), "data.id").String()
	if id == "" {
		return "", errors.New("could not get ID of created event")
	}

	return model.ID(id), nil
}

func requestEventJSONByID(ID model.ID) (string, error) {
	resp, err := http.Get("http://localhost:8889/api/v1/events/" + ID.String())
	if err != nil {
		return "", errors.Errorf("could not get event with ID [%s]", ID.String())
	}

	if resp.StatusCode != 200 {
		return "", errors.Errorf("could not get event with ID [%s]: code [%d] received", ID.String(), resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Errorf("could not parse json response for ID [%s]", ID.String())
	}

	return string(b), nil
}
