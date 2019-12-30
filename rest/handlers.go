package rest

import "github.com/labstack/echo"

type handlers struct {
	logger echo.Logger
}

func (h *handlers) CreateEvent(ctx echo.Context) error {
	return nil
}
