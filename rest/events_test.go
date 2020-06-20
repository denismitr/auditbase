package rest

import (
	"encoding/json"
	"github.com/denismitr/auditbase/test"
	"github.com/denismitr/auditbase/test/mock_cache"
	"github.com/denismitr/auditbase/test/mock_flow"
	"github.com/denismitr/auditbase/test/mock_model"
	"github.com/denismitr/auditbase/test/mock_utils/mock_clock"
	"github.com/denismitr/auditbase/test/mock_utils/mock_uuid"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestCreateEventWith(t *testing.T) {
	e := echo.New()
	log := logger.NewStdoutLogger("test", "events_test")

	t.Run("create event with ID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		eventsMock := mock_model.NewMockEventRepository(ctrl)
		flowMock := mock_flow.NewMockEventFlow(ctrl)
		clockMock := mock_clock.NewMockClock(ctrl)
		uuidMock := mock_uuid.NewMockUUID4Generator(ctrl)
		cacheMock := mock_cache.NewMockCacher(ctrl)

		// Todo: create event factory
		fakeEvent := CreateEvent{
			ID:            "897fe984-4445-43b0-9c71-797da1da242b",
			TargetID:      "1234",
			TargetEntity:  "article",
			TargetService: "article-storage",
			ActorID:       "4321",
			ActorEntity:   "editor",
			ActorService:  "back-office",
			EventName:     "article_published",
			EmittedAt:     int64(1578173213),
			RegisteredAt:  int64(1578173214),
			Changes:       make([]Change, 0),
		}

		body, _ := json.Marshal(fakeEvent)

		clockMock.EXPECT().CurrentTime().Return(time.Unix(1578173214, 0))
		flowMock.EXPECT().Send(gomock.Any()).Return(nil)
		cacheMock.EXPECT().Has(gomock.Any()).Return(false, nil) // fixme
		cacheMock.EXPECT().CreateKey(gomock.Any(), 1 * time.Minute).Return(nil)

		controller := newEventsController(log, uuidMock, clockMock, eventsMock, flowMock, cacheMock)

		req := test.Request{
			Method:            http.MethodPost,
			Target:            "/api/v1/events",
			IsContentTypeJSON: true,
			Body:              body,
			Controller:        controller.create,
		}

		resp := test.Invoke(e, req)

		assert.NoError(t, resp.Err)
		assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	})

	// 	t.Run("create event without explicitly provided ID", func(t *testing.T) {
	// 		ctrl := gomock.NewController(t)
	// 		defer ctrl.Finish()

	// 		eventsMock := mock_model.NewMockEventRepository(ctrl)
	// 		flowMock := mock_flow.NewMockEventFlow(ctrl)
	// 		clockMock := mock_utils.NewMockClock(ctrl)
	// 		uuidMock := mock_utils.NewMockUUID4Generatgor(ctrl)

	// 		id := "11122233-4445-43b0-9c71-797da1da242b"

	// 		fakeEvent := model.Events{
	// 			ID:            "",
	// 			TargetID:      "1234",
	// 			TargetType:    model.TargetType{Name: "article"},
	// 			TargetService: model.Microservices{Name: "article-storage"},
	// 			ActorID:       "4321",
	// 			ActorType:     model.ActorType{Name: "editor"},
	// 			ActorService:  model.Microservices{Name: "back-office"},
	// 			EventName:     "article_published",
	// 			EmittedAt:     1578173213,
	// 			RegisteredAt:  1578173214,
	// 			Delta:         map[string][]interface{}{"status": []interface{}{"PENDING", "PUBLISHED"}},
	// 		}

	// 		body, _ := json.Marshal(fakeEvent)

	// 		uuidMock.EXPECT().Generate().Return(id)
	// 		clockMock.EXPECT().CurrentTimestamp().Return(int64(1578173214))

	// 		fakeEvent.ID = id

	// 		flowMock.EXPECT().Send(fakeEvent).Return(nil)

	// 		controller := newEventsController(logger, uuidMock, clockMock, eventsMock, flowMock)

	// 		req := test.Request{
	// 			Method:            http.MethodPost,
	// 			Target:            "/api/v1/events",
	// 			IsContentTypeJSON: true,
	// 			Body:              body,
	// 			Controller:        controller.CreateEvent,
	// 		}

	// 		resp := test.Invoke(e, req)

	// 		assert.NoError(t, resp.Err)
	// 		assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	// 	})
	// }

	// func TestGetEvent(t *testing.T) {
	// 	e := echo.BackOfficeAPI()
	// 	logger := utils.NewStdoutLogger("test", "events_test")

	// 	t.Run("admin can get an event by ID", func(t *testing.T) {
	// 		ctrl := gomock.NewController(t)
	// 		defer ctrl.Finish()

	// 		eventsMock := mock_model.NewMockEventRepository(ctrl)
	// 		flowMock := mock_flow.NewMockEventFlow(ctrl)
	// 		clockMock := mock_utils.NewMockClock(ctrl)
	// 		uuidMock := mock_utils.NewMockUUID4Generatgor(ctrl)

	// 		id := "11122233-4445-43b0-9c71-797da1da242b"

	// 		fakeEvent := model.Events{
	// 			ID:            id,
	// 			TargetID:      "1234",
	// 			TargetType:    model.TargetType{Name: "article"},
	// 			TargetService: model.Microservices{Name: "article-storage"},
	// 			ActorID:       "4321",
	// 			ActorType:     model.ActorType{Name: "editor"},
	// 			ActorService:  model.Microservices{Name: "back-office"},
	// 			EventName:     "article_published",
	// 			EmittedAt:     1578173213,
	// 			RegisteredAt:  1578173214,
	// 			Delta:         map[string][]interface{}{"status": []interface{}{"PENDING", "PUBLISHED"}},
	// 		}

	// 		eventsMock.EXPECT().FindOneByID(model.ID(id)).Return(fakeEvent, nil)

	// 		controller := newEventsController(logger, uuidMock, clockMock, eventsMock, flowMock)

	// 		req := test.Request{
	// 			Method:            http.MethodGet,
	// 			Target:            "/api/v1/events/:id",
	// 			IsContentTypeJSON: true,
	// 			Body:              nil,
	// 			Controller:        controller.GetEvent,
	// 			Segments:          map[string]string{"id": id},
	// 		}

	// 		resp := test.Invoke(e, req)

	// 		assert.NoError(t, resp.Err)
	// 		assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 		assert.Equal(t, id, gjson.Get(resp.Body, "data.id").String())
	// 		assert.Equal(t, "1234", gjson.Get(resp.Body, "data.targetId").String())
	// 		assert.Equal(t, "4321", gjson.Get(resp.Body, "data.actorId").String())
	// 		assert.Equal(t, "article_published", gjson.Get(resp.Body, "data.eventName").String())
	// 	})
}
