package seeder

import (
	"bytes"
	"encoding/json"
	"github.com/denismitr/auditbase/internal/model"
	"github.com/denismitr/auditbase/internal/utils/faker"
	"github.com/denismitr/auditbase/internal/utils/random"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

type Result struct {
	ID         string
	StatusCode int
	Response   string
	Elapsed    time.Duration
}

type Sender struct {
	endpoint string
	client   *http.Client
	lg       *log.Logger
}

func NewSender(endpoint string, lg *log.Logger) *Sender {
	return &Sender{
		client:   &http.Client{
			Timeout: 2 * time.Second,
		},
		endpoint: endpoint,
		lg: lg,
	}
}

func GenerateNewActions(n int, crud model.Crud) <-chan model.NewAction {
	c := make(chan model.NewAction)

	go func() {
		for i := 0; i < n; i++ {
			minTs, err := time.Parse("2006-01-02", "2017-01-01")
			if err != nil {
				panic(err)
			}
			maxTs, err := time.Parse("2006-01-02", "2020-01-01")
			if err != nil {
				panic(err)
			}

			e := model.NewAction{
				TargetExternalID: faker.NumericID(150, 250),
				ActorService:     faker.WrappedString("actor", "service", 1),
				TargetService:    faker.WrappedString("target", "service", 1),
				ActorEntity:      faker.WrappedString("actor", "entity", 2),
				TargetEntity:     faker.WrappedString("target", "entity", 2),
				ActorExternalID:  faker.NumericID(150, 250),
				EmittedAt:        model.JSONTime{Time: random.Time(minTs, maxTs)},
				Name:             faker.WrappedString("event", "name", 2),
			}

			detailsCount := random.Int(3, 15)
			details := make(map[string]interface{})
			for i := 0; i < detailsCount; i++ {
				key := faker.WrappedString("property", "name", 2)

				switch crud {
				case model.UpdateAction, model.AnyAction:
					details[key] = [2]interface{}{faker.ChangeValue(false), faker.ChangeValue(true)}
				case model.CreateAction:
					details[key] = [2]interface{}{nil, faker.ChangeValue(true)}
				case model.DeleteAction:
					details[key] = [2]interface{}{faker.ChangeValue(true), nil}
				}
			}

			e.Details = details

			c <- e
		}

		close(c)
	}()

	return c
}

func (s *Sender) DispatchNewActions(errCh chan<- error, actionsChBulk ...<-chan model.NewAction) <-chan Result {
	resultCh := make(chan Result)

	go func() {
		defer close(resultCh)

		var wg sync.WaitGroup
		for _, actionsCh := range actionsChBulk {
			wg.Add(1)
			go s.dispatch(&wg, errCh, resultCh, actionsCh)
		}

		wg.Wait()
	}()

	return resultCh
}

func (s *Sender) dispatch(
	wg *sync.WaitGroup,
	errCh chan<- error,
	resultCh chan<- Result,
	actionsCh <-chan model.NewAction,
) {
	defer wg.Done()

	for action := range actionsCh {
		b, err := json.Marshal(action)
		if err != nil {
			errCh <- err
			continue
		}

		s.lg.Println(string(b))

		r := bytes.NewReader(b)
		req, err := http.NewRequest("POST", s.endpoint, r)
		if err != nil {
			errCh <- err
			continue
		}

		req.Header.Add("Content-Type", "application/json")

		start := time.Now()
		resp, err := s.client.Do(req)
		elapsed := time.Now().Sub(start).Milliseconds()
		if err != nil {
			errCh <- err
			continue
		}

		result := Result{
			StatusCode: resp.StatusCode,
			Elapsed:    time.Duration(elapsed),
		}

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			errCh <- err
			continue
		}

		if resp.StatusCode >= 300 {
			result.Response = string(respBody)
		} else {
			status := struct {
				Data struct {
					ID string `json:"id"`
				} `json:"data"`
			}{}

			if err := json.Unmarshal(respBody, &status); err != nil {
				errCh <- err
				continue
			}

			result.ID = status.Data.ID
		}

		resultCh <- result
	}
}
