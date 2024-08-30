package errorhandler

import (
	"goplate/http/reqresp"
	"goplate/pkg/json_logger"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type (
	ErrorHandler struct {
		Log json_logger.ITraceLogger
	}
)

var (
	statusCode = 400
)

func New(log json_logger.ITraceLogger) *ErrorHandler {
	return &ErrorHandler{
		Log: log,
	}
}

func (eh ErrorHandler) Interceptor(c *fiber.Ctx) error {
	err := c.Context().Err()
	if err != nil {
		re, ok := err.(*reqresp.Error)
		if ok {
			eh.Log.ErrorContext(c.UserContext(), re.LogError.Error())
			_ = c.Status(re.StatusCode).JSON(reqresp.NewError(re.StatusCode, err, re.Msg, *re.MsgType))
		} else {
			eh.Log.ErrorContext(c.UserContext(), err.Error())
			_ = c.Status(statusCode).JSON(reqresp.NewError(statusCode, err, err.Error(), strconv.Itoa(c.Response().StatusCode())))
		}
		return nil
	}
	return c.Next()
}
