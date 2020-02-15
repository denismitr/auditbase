package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/test/mock_model"
	"github.com/denismitr/auditbase/test/mock_utils"
	"github.com/denismitr/auditbase/utils"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

const idValidationFailed = `
{"error":{"title":"Validation failed","code":422,"details":"ID is incorrect","errors":{"ID":[":id must be a valid UUID4 or be null for auto assigning"]}}}
`

func TestShowMicroservice(t *testing.T) {
	e := echo.New()
	logger := utils.NewStdoutLogger("test", "events_test")

	t.Run("user can request a microservice by ID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		id := "35e46ed5-fe56-4445-878e-9c32ae54bfd0"

		uuidMock := mock_utils.NewMockUUID4Generatgor(ctrl)
		mrMock := mock_model.NewMockMicroserviceRepository(ctrl)

		m := model.Microservice{
			ID:          id,
			Name:        "FOO",
			Description: "BAR",
		}

		mrMock.
			EXPECT().
			FirstByID(model.ID(id)).
			Return(m, nil)

		target := "/api/v1/microservices/" + id
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, target, nil)
		c := newMicroservicesController(logger, uuidMock, mrMock)

		ctx := e.NewContext(req, rec)
		ctx.SetPath("/api/v1/microservices/:id")
		ctx.SetParamNames("id")
		ctx.SetParamValues(id)

		err := c.GetMicroservice(ctx)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		js := rec.Body.String()

		assert.Equal(t, id, gjson.Get(js, "data.id").String())
		assert.Equal(t, "FOO", gjson.Get(js, "data.name").String())
		assert.Equal(t, "BAR", gjson.Get(js, "data.description").String())
	})

	t.Run("user will receive an error if microservice is not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		id := "35e46ed5-fe56-4445-878e-9c32ae54bfd0"

		uuidMock := mock_utils.NewMockUUID4Generatgor(ctrl)
		mrMock := mock_model.NewMockMicroserviceRepository(ctrl)

		mrMock.
			EXPECT().
			FirstByID(model.ID(id)).
			Return(model.Microservice{}, model.ErrMicroserviceNotFound)

		target := "/api/v1/microservices/" + id
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, target, nil)
		c := newMicroservicesController(logger, uuidMock, mrMock)

		ctx := e.NewContext(req, rec)
		ctx.SetPath("/api/v1/microservices/:id")
		ctx.SetParamNames("id")
		ctx.SetParamValues(id)

		err := c.GetMicroservice(ctx)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)

		js := rec.Body.String()

		assert.Equal(t, "Not found", gjson.Get(js, "error.title").String())
		assert.Equal(t, "could not get microservice with ID 35e46ed5-fe56-4445-878e-9c32ae54bfd0 from database: not found", gjson.Get(js, "error.details").String())
	})

	t.Run("user cannot request a microservice by invalid ID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		uuidMock := mock_utils.NewMockUUID4Generatgor(ctrl)
		mrMock := mock_model.NewMockMicroserviceRepository(ctrl)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/microservices/123", nil)
		c := newMicroservicesController(logger, uuidMock, mrMock)

		ctx := e.NewContext(req, rec)
		ctx.SetPath("/api/v1/microservices/:id")
		ctx.SetParamNames("id")
		ctx.SetParamValues("foo")

		err := c.GetMicroservice(ctx)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

		js := rec.Body.String()

		assert.Equal(t, "Validation failed", gjson.Get(js, "error.title").String())
		assert.Equal(t, ":id is incorrect", gjson.Get(js, "error.details").String())
	})
}
