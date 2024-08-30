package tracer

import (
	"goplate/pkg/json_logger"

	icontext "goplate/pkg/context"

	"github.com/gofiber/fiber/v2"
)

type (
	Config struct {
		Log json_logger.ITraceLogger
	}
)

func New(config ...Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.SetUserContext(icontext.SetTraceID(c.UserContext(), c.Get(icontext.TraceIDKeyName, icontext.GetTraceID(c.UserContext()))))
		return c.Next()
	}
}
