package rest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/tests/mock_flow"
	"github.com/denismitr/auditbase/tests/mock_model"
	"github.com/denismitr/auditbase/tests/mock_utils"
	"github.com/denismitr/auditbase/utils"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

const eventWithID = `
	{"id":"897fe984-4445-43b0-9c71-797da1da242b","targetId":"1234","targetType":{"name":"article"},"targetService":{"name":"article-storage"},"actorId":"4321","actorType":{"name":"editor"},"actorService":{"name":"back-office"},"eventName":"article_published","emittedAt":1578173213,"delta":{"status":["PENDING","PUBLISHED"]}}
`

const eventWithoutID = `
	{"targetId":"1234","targetType":{"name":"article"},"targetService":{"name":"article-storage"},"actorId":"4321","actorType":{"name":"editor"},"actorService":{"name":"back-office"},"eventName":"article_published","emittedAt":1578173213,"delta":{"status":["PENDING","PUBLISHED"]}}
`

func TestCreateEventWithID(t *testing.T) {
	e := echo.New()

	logger := utils.NewStdoutLogger("test", "events_test")

	t.Run("create event with ID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		eventsMock := mock_model.NewMockEventRepository(ctrl)
		flowMock := mock_flow.NewMockEventFlow(ctrl)
		clockMock := mock_utils.NewMockClock(ctrl)
		uuidMock := mock_utils.NewMockUUID4Generatgor(ctrl)

		fakeEvent := model.Event{
			ID:            "897fe984-4445-43b0-9c71-797da1da242b",
			TargetID:      "1234",
			TargetType:    model.TargetType{Name: "article"},
			TargetService: model.Microservice{Name: "article-storage"},
			ActorID:       "4321",
			ActorType:     model.ActorType{Name: "editor"},
			ActorService:  model.Microservice{Name: "back-office"},
			EventName:     "article_published",
			EmittedAt:     1578173213,
			RegisteredAt:  1578173214,
			Delta:         map[string][]interface{}{"status": []interface{}{"PENDING", "PUBLISHED"}},
		}

		clockMock.EXPECT().CurrentTimestamp().Return(int64(1578173214))
		flowMock.EXPECT().Send(fakeEvent).Return(nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/events", strings.NewReader(eventWithID))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		ctx := e.NewContext(req, rec)
		ctx.SetPath("/api/v1/events")

		controller := newEventsController(logger, uuidMock, clockMock, eventsMock, flowMock)

		err := controller.CreateEvent(ctx)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusAccepted, rec.Code)
	})

	t.Run("create event with no explicit ID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		eventsMock := mock_model.NewMockEventRepository(ctrl)
		flowMock := mock_flow.NewMockEventFlow(ctrl)
		clockMock := mock_utils.NewMockClock(ctrl)
		uuidMock := mock_utils.NewMockUUID4Generatgor(ctrl)

		fakeEvent := model.Event{
			ID:            "11122233-4445-43b0-9c71-797da1da242b",
			TargetID:      "1234",
			TargetType:    model.TargetType{Name: "article"},
			TargetService: model.Microservice{Name: "article-storage"},
			ActorID:       "4321",
			ActorType:     model.ActorType{Name: "editor"},
			ActorService:  model.Microservice{Name: "back-office"},
			EventName:     "article_published",
			EmittedAt:     1578173213,
			RegisteredAt:  1578173214,
			Delta:         map[string][]interface{}{"status": []interface{}{"PENDING", "PUBLISHED"}},
		}

		uuidMock.EXPECT().Generate().Return("11122233-4445-43b0-9c71-797da1da242b")
		clockMock.EXPECT().CurrentTimestamp().Return(int64(1578173214))
		flowMock.EXPECT().Send(fakeEvent).Return(nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/events", strings.NewReader(eventWithoutID))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		ctx := e.NewContext(req, rec)
		ctx.SetPath("/api/v1/events")

		controller := newEventsController(logger, uuidMock, clockMock, eventsMock, flowMock)

		err := controller.CreateEvent(ctx)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusAccepted, rec.Code)
	})
}
