package htmx

import (
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/labstack/echo/v4"
)

func CreateFormHandler(
	cmd cmds.GlazeCommand,
	options ...FormHandlerOption,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		handler := NewFormHandler(cmd, options...)
		return handler.Handle(c)
	}
}
