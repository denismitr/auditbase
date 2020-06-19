package rest

import (
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/test"
	"github.com/denismitr/auditbase/test/mock_model"
	"github.com/denismitr/auditbase/test/mock_utils/mock_uuid"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"net/http"
	"testing"
)

func TestChangesController(t *testing.T) {
	e := echo.New()
	lg := logger.NewStdoutLogger("test", "events_test")

	t.Run("admin can request a property change by ID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		id := "35e46ed5-fe56-4445-878e-9c32ae54bfd0"
		propertyID := "11e46ed5-fe33-4445-878e-9c32ae54bfb2"
		eventID := "22e46ed5-fe33-4445-878e-9c32ae54bfb2"
		from := "BAR"
		to := "FOO"
		typ := "string"

		uuidMock := mock_uuid.NewMockUUID4Generator(ctrl)
		repoMock := mock_model.NewMockChangeRepository(ctrl)

		c := model.Change{
			ID:          id,
			PropertyID: propertyID,
			EventID:        eventID,
			CurrentDataType: &typ,
			From: &from,
			To: &to,
		}

		repoMock.
			EXPECT().
			FirstByID(id).
			Return(&c, nil)

		req := test.Request{
			Method:            http.MethodPut,
			Target:            "/api/v1/changes/:id",
			IsContentTypeJSON: true,
			Controller:        newChangesController(uuidMock, lg, repoMock).show,
			Segments:          map[string]string{"id": id},
		}

		resp := test.Invoke(e, req)

		assert.NoError(t, resp.Err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		assert.Equal(t, id, gjson.Get(resp.Body, "data.id").String())
		assert.Equal(t, "changes", gjson.Get(resp.Body, "data.type").String())
		assert.Equal(t, eventID, gjson.Get(resp.Body, "data.attributes.eventId").String())
		assert.Equal(t, propertyID, gjson.Get(resp.Body, "data.attributes.propertyId").String())
		assert.Equal(t, "string", gjson.Get(resp.Body, "data.attributes.currentDataType").String())
		assert.Equal(t, from, gjson.Get(resp.Body, "data.attributes.from").String())
		assert.Equal(t, to, gjson.Get(resp.Body, "data.attributes.to").String())
	})
}
