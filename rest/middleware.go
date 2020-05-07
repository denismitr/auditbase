package rest

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

func hashRequestBody(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		if ctx.Request().Body != nil {
			body, err := ioutil.ReadAll(ctx.Request().Body)
			if err != nil {
				ctx.Error(err)
				return nil
			}

			rc := ioutil.NopCloser(bytes.NewBuffer(body))
			defer rc.Close()

			hash, err := hashBody(rc)
			if err != nil {
				ctx.Error(err)
			}

			ctx.Request().Header.Add("Body-Hash", hash)
			ctx.Request().Body = ioutil.NopCloser(bytes.NewBuffer(body))
		}

		if err := next(ctx); err != nil {
			ctx.Error(err)
		}

		return nil
	}
}

func hashBody(r io.Reader) (string, error) {
	hasher := sha1.New()

	_, err := io.Copy(hasher, r)
	if err != nil {
		return "", errors.Wrap(err, "could not create hash from request body")
	}

	hash := hasher.Sum(nil)

	return fmt.Sprintf("%x", hash), nil
}
