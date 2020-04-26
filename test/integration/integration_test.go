package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/denismitr/auditbase/rest"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestCreateEvent(t *testing.T) {
	tt := []rest.CreateEvent{
		rest.CreateEvent{
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
	}

	for i, tc := range tt {
		t.Run(fmt.Sprintf("%s_%d", tc.EventName, i), func(t *testing.T) {
			b, err := json.Marshal(tc)
			if err != nil {
				t.Fatal(err)
			}

			r := bytes.NewReader(b)

			var wg sync.WaitGroup

			wg.Add(1)

			go func() {
				if _, err := http.Post("http://auditbase_rest:3000/api/events", "application/json", r); err != nil {
					t.Error(err)
					return
				}

				wg.Add(1)
				go func() {
					tick := time.After(2 * time.Second)

					select {
					case <-tick:
						resp, err := http.Get("http://auditbase_rest:3000/api/events")
						if err != nil {
							t.Error(err)
							return
						}

						b, err := ioutil.ReadAll(resp.Body())
						if err != nil {
							t.Error(err)
							return
						}

						name := gjson.Get(string(b), "data.0.Name")

						assert.Equal(t, name.String(), tc.Name)
					}
				}()
			}()

			wg.Wait()
		})
	}
}
