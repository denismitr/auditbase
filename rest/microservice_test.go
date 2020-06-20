package rest

// import (
// 	"encoding/json"
// 	"net/http"
// 	"testing"

// 	"github.com/denismitr/auditbase/model"
// 	"github.com/denismitr/auditbase/test"
// 	"github.com/denismitr/auditbase/test/mock_model"
// 	"github.com/denismitr/auditbase/test/mock_utils"
// 	"github.com/denismitr/auditbase/utils"
// 	"github.com/golang/mock/gomock"
// 	"github.com/labstack/echo"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/tidwall/gjson"
// )

// func TestGetMicroservice(t *testing.T) {
// 	e := echo.BackOfficeAPI()
// 	logger := utils.NewStdoutLogger("test", "events_test")

// 	t.Run("user can request a microservice by ID", func(t *testing.T) {
// 		ctrl := gomock.NewController(t)
// 		defer ctrl.Finish()

// 		id := "35e46ed5-fe56-4445-878e-9c32ae54bfd0"

// 		uuidMock := mock_utils.NewMockUUID4Generatgor(ctrl)
// 		mrMock := mock_model.NewMockMicroserviceRepository(ctrl)

// 		m := model.Microservices{
// 			ID:          id,
// 			Name:        "FOO",
// 			Description: "BAR",
// 		}

// 		mrMock.
// 			EXPECT().
// 			FirstByID(model.ID(id)).
// 			Return(m, nil)

// 		req := test.Request{
// 			Method:            http.MethodPut,
// 			Target:            "/api/v1/microservices/:id",
// 			IsContentTypeJSON: true,
// 			Controller:        newMicroservicesController(logger, uuidMock, mrMock).GetMicroservice,
// 			Segments:          map[string]string{"id": id},
// 		}

// 		resp := test.Invoke(e, req)

// 		assert.NoError(t, resp.Err)
// 		assert.Equal(t, http.StatusOK, resp.StatusCode)

// 		assert.Equal(t, id, gjson.Get(resp.Body, "data.id").String())
// 		assert.Equal(t, "FOO", gjson.Get(resp.Body, "data.name").String())
// 		assert.Equal(t, "BAR", gjson.Get(resp.Body, "data.description").String())
// 	})

// 	t.Run("user will receive an error if microservice is not found", func(t *testing.T) {
// 		ctrl := gomock.NewController(t)
// 		defer ctrl.Finish()

// 		id := "35e46ed5-fe56-4445-878e-9c32ae54bfd0"

// 		uuidMock := mock_utils.NewMockUUID4Generatgor(ctrl)
// 		mrMock := mock_model.NewMockMicroserviceRepository(ctrl)

// 		mrMock.
// 			EXPECT().
// 			FirstByID(model.ID(id)).
// 			Return(model.Microservices{}, model.ErrMicroserviceNotFound)

// 		req := test.Request{
// 			Method:            http.MethodPut,
// 			Target:            "/api/v1/microservices/:id",
// 			IsContentTypeJSON: true,
// 			Controller:        newMicroservicesController(logger, uuidMock, mrMock).GetMicroservice,
// 			Segments:          map[string]string{"id": id},
// 		}

// 		resp := test.Invoke(e, req)

// 		assert.NoError(t, resp.Err)
// 		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
// 		assert.Equal(t, "Not found", gjson.Get(resp.Body, "error.title").String())
// 		assert.Equal(
// 			t,
// 			"could not get microservice with ID 35e46ed5-fe56-4445-878e-9c32ae54bfd0 from database: not found",
// 			gjson.Get(resp.Body, "error.details").String(),
// 		)
// 	})

// 	t.Run("user cannot request a microservice by invalid ID", func(t *testing.T) {
// 		ctrl := gomock.NewController(t)
// 		defer ctrl.Finish()

// 		uuidMock := mock_utils.NewMockUUID4Generatgor(ctrl)
// 		mrMock := mock_model.NewMockMicroserviceRepository(ctrl)

// 		req := test.Request{
// 			Method:            http.MethodPut,
// 			Target:            "/api/v1/microservices/:id",
// 			IsContentTypeJSON: true,
// 			Controller:        newMicroservicesController(logger, uuidMock, mrMock).GetMicroservice,
// 			Segments:          map[string]string{"id": "foo"},
// 		}

// 		resp := test.Invoke(e, req)

// 		assert.NoError(t, resp.Err)
// 		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)

// 		assert.Equal(t, "Validation failed", gjson.Get(resp.Body, "error.title").String())
// 		assert.Equal(t, ":id is incorrect", gjson.Get(resp.Body, "error.details").String())
// 	})
// }

// func TestSelectMicroservices(t *testing.T) {
// 	e := echo.BackOfficeAPI()
// 	logger := utils.NewStdoutLogger("test", "events_test")

// 	t.Run("select all microservices", func(t *testing.T) {
// 		ctrl := gomock.NewController(t)
// 		defer ctrl.Finish()

// 		uuidMock := mock_utils.NewMockUUID4Generatgor(ctrl)
// 		mrMock := mock_model.NewMockMicroserviceRepository(ctrl)

// 		mm := []model.Microservices{
// 			{ID: "35e46ed5-fe56-4445-878e-9c32ae54bfd0", Name: "Foo", Description: "Bar"},
// 			{ID: "22e46ed5-fe56-4445-878e-9c32ae54bf11", Name: "Foo2", Description: "Bar2"},
// 		}

// 		mrMock.
// 			EXPECT().
// 			SelectAll().
// 			Return(mm, nil)

// 		req := test.Request{
// 			Method:            http.MethodGet,
// 			Target:            "/api/v1/microservices",
// 			Body:              nil,
// 			IsContentTypeJSON: false,
// 			Controller:        newMicroservicesController(logger, uuidMock, mrMock).SelectMicroservices,
// 		}

// 		resp := test.Invoke(e, req)

// 		assert.NoError(t, resp.Err)
// 		assert.Equal(t, http.StatusOK, resp.StatusCode)

// 		assert.Equal(t, int64(2), gjson.Get(resp.Body, "data.#").Int())
// 		assert.Equal(t, "35e46ed5-fe56-4445-878e-9c32ae54bfd0", gjson.Get(resp.Body, "data.0.id").String())
// 		assert.Equal(t, "22e46ed5-fe56-4445-878e-9c32ae54bf11", gjson.Get(resp.Body, "data.1.id").String())
// 		assert.Equal(t, "Foo", gjson.Get(resp.Body, "data.0.name").String())
// 		assert.Equal(t, "Foo2", gjson.Get(resp.Body, "data.1.name").String())
// 		assert.Equal(t, "Bar", gjson.Get(resp.Body, "data.0.description").String())
// 		assert.Equal(t, "Bar2", gjson.Get(resp.Body, "data.1.description").String())
// 	})
// }

// func TestUpdateMicroservice(t *testing.T) {
// 	e := echo.BackOfficeAPI()
// 	logger := utils.NewStdoutLogger("test", "events_test")

// 	t.Run("admin can update existing microservice", func(t *testing.T) {
// 		ctrl := gomock.NewController(t)
// 		defer ctrl.Finish()

// 		id := "35e46ed5-fe56-4445-878e-9c32ae54bfd0"

// 		uuidMock := mock_utils.NewMockUUID4Generatgor(ctrl)
// 		mrMock := mock_model.NewMockMicroserviceRepository(ctrl)

// 		m := model.Microservices{
// 			ID:          id,
// 			Name:        "FOO 23",
// 			Description: "BAR 23",
// 		}

// 		body, _ := json.Marshal(m)

// 		mrMock.EXPECT().Update(model.ID(id), m).Return(nil)
// 		mrMock.EXPECT().FirstByID(model.ID(id)).Return(m, nil)

// 		req := test.Request{
// 			Method:            http.MethodPut,
// 			Target:            "/api/v1/microservices/:id",
// 			Body:              body,
// 			IsContentTypeJSON: true,
// 			Controller:        newMicroservicesController(logger, uuidMock, mrMock).UpdateMicroservice,
// 			Segments:          map[string]string{"id": id},
// 		}

// 		resp := test.Invoke(e, req)

// 		assert.NoError(t, resp.Err)
// 		assert.Equal(t, http.StatusOK, resp.StatusCode)

// 		assert.Equal(t, id, gjson.Get(resp.Body, "data.id").String())
// 		assert.Equal(t, "FOO 23", gjson.Get(resp.Body, "data.name").String())
// 		assert.Equal(t, "BAR 23", gjson.Get(resp.Body, "data.description").String())
// 	})
// }

// func TestCreateMicroservice(t *testing.T) {
// 	e := echo.BackOfficeAPI()
// 	logger := utils.NewStdoutLogger("test", "events_test")

// 	t.Run("admin cannot create a microservice without a name", func(t *testing.T) {
// 		ctrl := gomock.NewController(t)
// 		defer ctrl.Finish()

// 		id := "35e46ed5-fe56-4445-878e-9c32ae54bfd0"

// 		uuidMock := mock_utils.NewMockUUID4Generatgor(ctrl)
// 		mrMock := mock_model.NewMockMicroserviceRepository(ctrl)

// 		m := model.Microservices{
// 			Name:        "",
// 			Description: "BAR 23",
// 		}

// 		body, _ := json.Marshal(m)

// 		uuidMock.EXPECT().Generate().Return(id)

// 		req := test.Request{
// 			Method:            http.MethodPost,
// 			Target:            "/api/v1/microservices/",
// 			Body:              body,
// 			IsContentTypeJSON: true,
// 			Controller:        newMicroservicesController(logger, uuidMock, mrMock).CreateMicroservice,
// 		}

// 		resp := test.Invoke(e, req)

// 		assert.NoError(t, resp.Err)
// 		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
// 		js := resp.Body

// 		assert.Equal(t, "Validation failed", gjson.Get(js, "error.title").String())
// 		assert.Equal(t, "bad data for a microservice", gjson.Get(js, "error.details").String())
// 	})

// 	t.Run("admin can create a new microservice without providing ID", func(t *testing.T) {
// 		ctrl := gomock.NewController(t)
// 		defer ctrl.Finish()

// 		id := "35e46ed5-fe56-4445-878e-9c32ae54bfd0"

// 		uuidMock := mock_utils.NewMockUUID4Generatgor(ctrl)
// 		mrMock := mock_model.NewMockMicroserviceRepository(ctrl)

// 		m := model.Microservices{
// 			Name:        "FOO 23",
// 			Description: "BAR 23",
// 		}

// 		body, _ := json.Marshal(m)

// 		uuidMock.EXPECT().Generate().Return(id)
// 		mrMock.
// 			EXPECT().
// 			Create(cloneMicroserviceWithNewID(m, model.ID(id))).
// 			Return(cloneMicroserviceWithNewID(m, model.ID(id)), nil)

// 		req := test.Request{
// 			Method:            http.MethodPost,
// 			Target:            "/api/v1/microservices/",
// 			Body:              body,
// 			IsContentTypeJSON: true,
// 			Controller:        newMicroservicesController(logger, uuidMock, mrMock).CreateMicroservice,
// 		}

// 		resp := test.Invoke(e, req)

// 		assert.NoError(t, resp.Err)
// 		assert.Equal(t, http.StatusCreated, resp.StatusCode)

// 		assert.Equal(t, id, gjson.Get(resp.Body, "data.id").String())
// 		assert.Equal(t, "FOO 23", gjson.Get(resp.Body, "data.name").String())
// 		assert.Equal(t, "BAR 23", gjson.Get(resp.Body, "data.description").String())
// 	})
// }

// func cloneMicroserviceWithNewID(m model.Microservices, ID model.ID) model.Microservices {
// 	m.ID = ID.String()
// 	return m
// }
