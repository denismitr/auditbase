package integration

//import (
//	"bytes"
//	"encoding/json"
//	"fmt"
//	"github.com/denismitr/auditbase/model"
//	"github.com/pkg/errors"
//	"io/ioutil"
//	"net/http"
//	"sync"
//	"testing"
//	"time"
//
//	"github.com/denismitr/auditbase/rest"
//	"github.com/stretchr/testify/assert"
//	"github.com/tidwall/gjson"
//)
//
//const backOfficeAddr = "localhost:8889"
//
//var tt = []rest.CreateEvent{
//	rest.CreateEvent{
//		EventName:     "foo",
//		TargetEntityID:      "1234",
//		TargetEntity:  "article",
//		TargetService: "article-storage",
//		ActorEntityID:       "4321",
//		ActorEntity:   "editor",
//		ActorService:  "back-office",
//		EmittedAt:     int64(1578173213),
//		RegisteredAt:  int64(1578173214),
//		Delta:         map[string][]interface{}{"name": []interface{}{"PENDING", "PUBLISHED"}},
//	},
//
//	rest.CreateEvent{
//		EventName:     "bar",
//		TargetEntityID:      "938-UE",
//		TargetEntity:  "post",
//		TargetService: "blog",
//		ActorEntityID:       "999",
//		ActorEntity:   "writer",
//		ActorService:  "user-service",
//		EmittedAt:     int64(1578178213),
//		RegisteredAt:  int64(1578178314),
//		Delta:         map[string][]interface{}{
//			"text": []interface{}{"FOO", "BAR"},
//			"status": []interface{}{nil, "published"},
//		},
//	},
//}
//
//var wg sync.WaitGroup
//
//func TestEvents(t *testing.T) {
//	setup(t)
//
//	for _, tc := range tt {
//		name := fmt.Sprintf("event_id_%s", tc.ID)
//
//		t.Run(name, func(t *testing.T) {
//			wg.Add(1)
//			go func() {
//				defer wg.Done()
//				tick := time.After(1 * time.Second)
//
//				select {
//				case <-tick:
//					js, err := requestEventJSONByID(model.ID(tc.ID))
//					if err != nil {
//						t.Error(err)
//						return
//					}
//
//					assert.Equal(t, tc.ID, gjson.Get(js, "data.id").String())
//					assert.Equal(t, tc.EventName, gjson.Get(js, "data.attributes.eventName").String())
//					assert.Equal(t, time.Unix(tc.EmittedAt, 0).UTC().Format(model.DefaultTimeFormat), gjson.Get(js, "data.attributes.emittedAt").String())
//					assert.Equal(t, tc.TargetEntityID, gjson.Get(js, "data.attributes.targetId").String())
//					assert.Equal(t, tc.ActorEntityID, gjson.Get(js, "data.attributes.actorId").String())
//					assert.Equal(t, tc.TargetEntity, gjson.Get(js, "data.attributes.targetEntity.name").String())
//					assert.Equal(t, tc.TargetService, gjson.Get(js, "data.attributes.targetService.name").String())
//					assert.Equal(t, tc.ActorEntity, gjson.Get(js, "data.attributes.actorEntity.name").String())
//					assert.Equal(t, tc.ActorService, gjson.Get(js, "data.attributes.actorService.name").String())
//				}
//			}()
//
//			wg.Wait()
//		})
//	}
//
//	t.Run("select all events", func(t *testing.T) {
//		wg.Add(1)
//
//		go func() {
//			defer wg.Done()
//			tick := time.After(1 * time.Second)
//
//			select {
//			case <-tick:
//				js, err := requestAllEvents()
//				if err != nil {
//					t.Error(err)
//					return
//				}
//
//				assert.True(t, gjson.Get(js, "data").IsArray())
//				assert.GreaterOrEqual(t, len(gjson.Get(js, "data").Array()), len(tt))
//			}
//		}()
//
//		wg.Wait()
//	})
//}
//
//func setup(t *testing.T) {
//	wg.Add(len(tt))
//
//	for i := range tt {
//		go func(ce *rest.CreateEvent) {
//			defer wg.Done()
//			id, err := createEvent(ce)
//			if err != nil {
//				t.Error(err)
//				return
//			}
//			ce.ID = id.String()
//		}(&tt[i])
//	}
//
//	wg.Wait()
//}
//
//func createEvent(e *rest.CreateEvent) (model.ID, error) {
//	b, err := json.Marshal(e)
//	if err != nil {
//		return "", errors.Wrap(err, "could not marshal event")
//	}
//
//	reader := bytes.NewReader(b)
//
//	r, err := http.Post("http://localhost:8888/api/v1/events", "application/json", reader);
//	if err != nil {
//		return "", errors.Wrapf(err, "event creation failed")
//	}
//
//	if r.StatusCode != 202 {
//		return "", errors.Errorf("could not create an event: [%d] code received", r.StatusCode)
//	}
//
//	js, err := ioutil.ReadAll(r.Body)
//	if err != nil {
//		return "", errors.New("could not parse accepted response")
//	}
//
//	id := gjson.Get(string(js), "data.id").String()
//	if id == "" {
//		return "", errors.New("could not get ID of created event")
//	}
//
//	return model.ID(id), nil
//}
//
//func requestEventJSONByID(ID model.ID) (string, error) {
//	resp, err := http.Get("http://localhost:8889/api/v1/events/" + ID.String())
//	if err != nil {
//		return "", errors.Errorf("could not get event with ID [%s]", ID.String())
//	}
//
//	if resp.StatusCode != 200 {
//		return "", errors.Errorf("could not get event with ID [%s]: code [%d] received", ID.String(), resp.StatusCode)
//	}
//
//	b, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		return "", errors.Errorf("could not parse json response for ID [%s]", ID.String())
//	}
//
//	return string(b), nil
//}
//
//func requestAllEvents() (string, error) {
//	resp, err := http.Get(fmt.Sprintf("http://%s/api/v1/events", backOfficeAddr))
//	if err != nil {
//		return "", errors.Errorf("could not get list of events")
//	}
//
//	if resp.StatusCode != 200 {
//		return "", errors.Errorf("could not get event list: code [%d] received", resp.StatusCode)
//	}
//
//	b, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		return "", errors.New("could not parse json response for event list")
//	}
//
//	return string(b), nil
//}
