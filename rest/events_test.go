package rest

import (
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/test"
	"github.com/denismitr/auditbase/test/factory"
	"github.com/denismitr/auditbase/test/mock_cache"
	"github.com/denismitr/auditbase/test/mock_flow"
	"github.com/denismitr/auditbase/test/mock_model"
	"github.com/denismitr/auditbase/test/mock_utils/mock_clock"
	"github.com/denismitr/auditbase/test/mock_utils/mock_uuid"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"net/http"
	"testing"
	"time"
)

func TestCreateEventWith(t *testing.T) {
	t.Run("admin can get an event by ID", func(t *testing.T) {
		e := echo.New()
		lg := logger.NewStdoutLogger("test", "events_test")

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		eventsMock := mock_model.NewMockEventRepository(ctrl)
		flowMock := mock_flow.NewMockEventFlow(ctrl)
		clockMock := mock_clock.NewMockClock(ctrl)
		uuidMock := mock_uuid.NewMockUUID4Generator(ctrl)
		cacheMock := mock_cache.NewMockCacher(ctrl)

		fakeEvent := factory.MatchingEvent(factory.EventState{
			State: factory.DefaultEvent,
			Now: time.Now(),
		})

		eventsMock.EXPECT().FindOneByID(model.ID(fakeEvent.ID())).Return(fakeEvent.Evt, nil)

		controller := newEventsController(lg, uuidMock, clockMock, eventsMock, flowMock, cacheMock)

		req := test.Request{
			Method:            http.MethodGet,
			Target:            "/api/v1/events/:id",
			IsContentTypeJSON: true,
			Body:              nil,
			Controller:        controller.show,
			Segments:          map[string]string{"id": fakeEvent.ID()},
		}

		resp := test.Invoke(e, req)

		assert.NoError(t, resp.Err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		//panic(gjson.Get(resp.Body, "data").String())
		assert.Equal(t, "34e1d82a-a065-436d-afd0-5fbcb752a4e1", gjson.Get(resp.Body, "data.id").String())
		assert.Equal(t, "444e1d82a-a065-436d-afd0-5fbcb752ae5", gjson.Get(resp.Body, "data.attributes.targetEntity.id").String())
		assert.Equal(t, "subscriptionCanceled", gjson.Get(resp.Body, "data.attributes.eventName").String())
	})
}
