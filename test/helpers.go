package test

import (
	"bytes"
	"io"
	"net/http/httptest"

	"github.com/labstack/echo"
)

type Response struct {
	StatusCode int
	Body       string
	Err        error
}

type Request struct {
	Method            string
	Target            string
	Path              string
	Body              []byte
	IsContentTypeJSON bool
	Segments          map[string]string
	Controller        func(ctx echo.Context) error
}

func Invoke(e *echo.Echo, r Request) Response {
	var bodyReader io.Reader
	if r.Body != nil {
		bodyReader = bytes.NewReader(r.Body)
	}

	rec := httptest.NewRecorder()

	req := httptest.NewRequest(r.Method, r.Target, bodyReader)

	req.Header.Set("Accept", "application/json")
	if r.IsContentTypeJSON {
		req.Header.Set("Content-Type", "application/json")
	}

	ctx := e.NewContext(req, rec)
	ctx.SetPath(r.Path)
	if r.Segments != nil {
		for k, v := range r.Segments {
			ctx.SetParamNames(k)
			ctx.SetParamValues(v)
		}
	}

	err := r.Controller(ctx)

	return Response{
		StatusCode: rec.Code,
		Body:       rec.Body.String(),
		Err:        err,
	}
}
