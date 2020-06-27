package rest

import (
	"github.com/denismitr/auditbase/test"
	"github.com/denismitr/auditbase/test/factory"
	"github.com/denismitr/auditbase/test/mock_cache"
	"github.com/denismitr/auditbase/test/mock_flow"
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

func TestReceiverController(t *testing.T) {
	e := echo.New()
	lg := logger.NewStdoutLogger("test", "receiver_test")

	t.Run("event with ID can be pushed into auditbase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		now := time.Now()

		bytes, evt := factory.MatchingIncomingEvent(factory.IncomingEventState{
			State: factory.EventWithID,
			Now: now,
		})

		uuidMock := mock_uuid.NewMockUUID4Generator(ctrl)
		efMock := mock_flow.NewMockEventFlow(ctrl)
		clock := mock_clock.NewMockClock(ctrl)
		cacher := mock_cache.NewMockCacher(ctrl)

		gomock.InOrder(
			cacher.EXPECT().Has("hash_key_1790e8a793ecd7f0b3e46c5dc5f71d18fc24c45a").Return(false, nil),
			clock.EXPECT().CurrentTime().Return(now),
			cacher.EXPECT().CreateKey("hash_key_1790e8a793ecd7f0b3e46c5dc5f71d18fc24c45a", 1 * time.Minute).Return(nil),
			efMock.EXPECT().Send(evt).Return(nil),
		)

		c := &receiverController{lg: lg, uuid4: uuidMock, ef: efMock, clock: clock, cacher: cacher}

		req := test.Request{
			Method:            http.MethodPost,
			Target:            "/api/v1/events",
			IsContentTypeJSON: true,
			Body: bytes,
			Controller:        c.create,
		}

		resp := test.Invoke(e, req, hashRequestBody)

		assert.NoError(t, resp.Err)
		assert.Equal(t, http.StatusAccepted, resp.StatusCode)
		assert.Equal(t, evt.ID(), gjson.Get(resp.Body, "data.id").String())
	})

	t.Run("event without ID can be pushed into auditbase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		now := time.Now()

		bytes, evt := factory.MatchingIncomingEvent(factory.IncomingEventState{
			State: factory.EventWithoutID,
			Now: now,
		})

		uuidMock := mock_uuid.NewMockUUID4Generator(ctrl)
		efMock := mock_flow.NewMockEventFlow(ctrl)
		clock := mock_clock.NewMockClock(ctrl)
		cacher := mock_cache.NewMockCacher(ctrl)

		gomock.InOrder(
			cacher.EXPECT().Has("hash_key_fb01901eb94091e8dd6c38c81f7d2576ff4ec735").Return(false, nil),
			uuidMock.EXPECT().Generate().Return("22e1d82a-a065-436d-afd0-5fbcb752a4f3"),
			clock.EXPECT().CurrentTime().Return(now),
			cacher.EXPECT().CreateKey("hash_key_fb01901eb94091e8dd6c38c81f7d2576ff4ec735", 1 * time.Minute).Return(nil),
			efMock.EXPECT().Send(evt).Return(nil),
		)

		c := &receiverController{lg: lg, uuid4: uuidMock, ef: efMock, clock: clock, cacher: cacher}

		req := test.Request{
			Method:            http.MethodPost,
			Target:            "/api/v1/events",
			IsContentTypeJSON: true,
			Body: bytes,
			Controller:        c.create,
		}

		resp := test.Invoke(e, req, hashRequestBody)

		assert.NoError(t, resp.Err)
		assert.Equal(t, http.StatusAccepted, resp.StatusCode)
		assert.Equal(t, evt.ID(), gjson.Get(resp.Body, "data.id").String())
	})

	t.Run("event without emittedAt cannot be pushed into auditbase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		now := time.Now()

		bytes, _ := factory.MatchingIncomingEvent(factory.IncomingEventState{
			State: factory.EventWithoutEmittedAt,
			Now: now,
		})

		uuidMock := mock_uuid.NewMockUUID4Generator(ctrl)
		efMock := mock_flow.NewMockEventFlow(ctrl)
		clock := mock_clock.NewMockClock(ctrl)
		cacher := mock_cache.NewMockCacher(ctrl)

		c := &receiverController{lg: lg, uuid4: uuidMock, ef: efMock, clock: clock, cacher: cacher}

		req := test.Request{
			Method:            http.MethodPost,
			Target:            "/api/v1/events",
			IsContentTypeJSON: true,
			Body: bytes,
			Controller:        c.create,
		}

		resp := test.Invoke(e, req, hashRequestBody)

		assert.NoError(t, resp.Err)
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
		assert.Equal(t, "Validation failed", gjson.Get(resp.Body, "errors.0.title").String())
		assert.Equal(t, "emittedAt must not be empty", gjson.Get(resp.Body, "errors.0.details").String())
	})
}
