package rest

import (
	"context"
	"github.com/labstack/echo"
)

// API - receiver API of auditbase
type API struct {
	e   *echo.Echo
	cfg Config
}

type StopFunc func(context.Context) error

// Start receiver Server and return stop function
func (a *API) Start() StopFunc {
	go func() {
		if err := a.e.Start(a.cfg.Port); err != nil {
			a.e.Logger.Errorf("shutting down the server", err)
		}
	}()

	return a.stop
}

func (a *API) stop(ctx context.Context) error {
	return a.e.Shutdown(ctx)
}
