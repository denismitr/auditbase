package rest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestGetMicroservice(t *testing.T) {
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

func TestSelectMicroservices(t *testing.T) {
	e := echo.New()
	logger := utils.NewStdoutLogger("test", "events_test")

	t.Run("select all microservices", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		uuidMock := mock_utils.NewMockUUID4Generatgor(ctrl)
		mrMock := mock_model.NewMockMicroserviceRepository(ctrl)

		mm := []model.Microservice{
			{ID: "35e46ed5-fe56-4445-878e-9c32ae54bfd0", Name: "Foo", Description: "Bar"},
			{ID: "22e46ed5-fe56-4445-878e-9c32ae54bf11", Name: "Foo2", Description: "Bar2"},
		}

		mrMock.
			EXPECT().
			SelectAll().
			Return(mm, nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/microservices/123", nil)
		c := newMicroservicesController(logger, uuidMock, mrMock)

		ctx := e.NewContext(req, rec)

		err := c.SelectMicroservices(ctx)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		js := rec.Body.String()

		assert.Equal(t, int64(2), gjson.Get(js, "data.#").Int())
		assert.Equal(t, "35e46ed5-fe56-4445-878e-9c32ae54bfd0", gjson.Get(js, "data.0.id").String())
		assert.Equal(t, "22e46ed5-fe56-4445-878e-9c32ae54bf11", gjson.Get(js, "data.1.id").String())
		assert.Equal(t, "Foo", gjson.Get(js, "data.0.name").String())
		assert.Equal(t, "Foo2", gjson.Get(js, "data.1.name").String())
		assert.Equal(t, "Bar", gjson.Get(js, "data.0.description").String())
		assert.Equal(t, "Bar2", gjson.Get(js, "data.1.description").String())
	})
}

func TestUpdateMicroservice(t *testing.T) {
	e := echo.New()
	logger := utils.NewStdoutLogger("test", "events_test")

	t.Run("admin can update existing microservice", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		id := "35e46ed5-fe56-4445-878e-9c32ae54bfd0"

		uuidMock := mock_utils.NewMockUUID4Generatgor(ctrl)
		mrMock := mock_model.NewMockMicroserviceRepository(ctrl)

		m := model.Microservice{
			ID:          id,
			Name:        "FOO 23",
			Description: "BAR 23",
		}

		b, _ := json.Marshal(m)
		body := string(b)

		mrMock.EXPECT().Update(model.ID(id), m).Return(nil)
		mrMock.EXPECT().FirstByID(model.ID(id)).Return(m, nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/api/v1/microservices/"+id, strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		c := newMicroservicesController(logger, uuidMock, mrMock)

		ctx := e.NewContext(req, rec)
		ctx.SetPath("/api/v1/microservices/:id")
		ctx.SetParamNames("id")
		ctx.SetParamValues(id)

		err := c.UpdateMicroservice(ctx)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		js := rec.Body.String()

		assert.Equal(t, id, gjson.Get(js, "data.id").String())
		assert.Equal(t, "FOO 23", gjson.Get(js, "data.name").String())
		assert.Equal(t, "BAR 23", gjson.Get(js, "data.description").String())
	})
}
