package seeder

import (
	"bytes"
	"encoding/json"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/rest"
	"github.com/denismitr/auditbase/test/faker"
	"github.com/denismitr/auditbase/utils/random"
	"github.com/denismitr/auditbase/utils/uuid"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

type Result struct {
	ID         string
	StatusCode int
	Err        error
	Elapsed    time.Duration
}

type Seeder struct {
	n      int
	withID bool
	crud   model.Crud
}

func New(n int, withID bool, crud model.Crud) *Seeder {
	return &Seeder{
		n:      n,
		withID: withID,
		crud:   crud,
	}
}

func (s *Seeder) Seed() <-chan rest.CreateEvent {
	c := make(chan rest.CreateEvent)

	go func() {
		for i := 0; i < s.n; i++ {
			minTs, err := time.Parse("2006-01-02", "2017-01-01")
			if err != nil {
				panic(err)
			}
			maxTs, err := time.Parse("2006-01-02", "2020-01-01")
			if err != nil {
				panic(err)
			}

			e := rest.CreateEvent{
				ActorID:       faker.NumericID(150, 250),
				TargetID:      faker.NumericID(150, 250),
				ActorService:  faker.WrappedString("actor", "service", 2),
				TargetService: faker.WrappedString("target", "service", 2),
				ActorEntity:   faker.WrappedString("actor", "entity", 2),
				TargetEntity:  faker.WrappedString("target", "entity", 2),
				Crud:          int(s.crud),
				EmittedAt:     random.Timestamp(minTs, maxTs),
				EventName:     faker.WrappedString("event", "name", 2),
				Changes:       make([]*rest.Change, 0),
			}

			if s.withID {
				e.ID = uuid.NewUUID4Generator().Generate()
			}

			changesCount := random.Int(3, 15)
			for i := 0; i < changesCount; i++ {
				change := &rest.Change{}
				change.PropertyName = faker.WrappedString("property", "name", 2)

				switch s.crud {
				case model.Update, model.Unknown:
					change.From = faker.ChangeValue(false)
					change.To = faker.ChangeValue(true)
				case model.Create:
					change.To = faker.ChangeValue(false)
				case model.Delete:
					change.From = faker.ChangeValue(false)
					change.To = nil
				}

				e.Changes = append(e.Changes, change)
			}

			c <- e
		}

		close(c)
	}()

	return c
}

func Send(endpoint string, events ...<-chan rest.CreateEvent) <-chan Result {
	result := make(chan Result)

	go func() {
		client := &http.Client{}
		var wg sync.WaitGroup

		for i := 0; i < len(events); i++ {
			wg.Add(1)
			go func(event <-chan rest.CreateEvent) {
				defer wg.Done()
				for e := range event {
					b, err := json.Marshal(e)
					if err != nil {
						result <- Result{
							Err: err,
						}
						continue
					}

					log.Println(string(b))

					r := bytes.NewReader(b)
					req, err := http.NewRequest("POST", endpoint, r)
					if err != nil {
						result <- Result{Err: err}
						continue
					}

					req.Header.Add("Content-Type", "application/json")

					start := time.Now()
					resp, err := client.Do(req)
					elapsed := time.Now().Sub(start).Milliseconds()

					if err != nil {
						result <- Result{
							Err:        err,
							Elapsed: time.Duration(elapsed),
						}
						continue
					}

					if resp.StatusCode == 400 {
						b, _ := ioutil.ReadAll(resp.Body)
						log.Printf("%s", string(b))
					}

					_ = resp.Body.Close()

					result <- Result{
						StatusCode: resp.StatusCode,
						Err:        err,
						Elapsed: time.Duration(elapsed),
					}
				}
			}(events[i])
		}

		wg.Wait()
		close(result)
	}()

	return result
}
