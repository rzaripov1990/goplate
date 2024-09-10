package interceptor

import (
	"encoding/json"
	"fmt"

	//"goplate/http/reqresp"

	trace_context "github.com/rzaripov1990/trace_ctx"

	"goplate/pkg/trace_logger"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type (
	Config struct {
		Log *slog.Logger

		ErrorHandler fiber.ErrorHandler

		//DefaultPanicError *reqresp.Error

		// Enable logging all requests.
		//
		// Optional. Default value true
		EnableLogRequest bool

		// Enable logging request headers.
		// EnableLogRequest = true is required
		//
		// Optional. Default value true
		EnableLogHeaders bool

		// Enable logging all responses.
		//
		// Optional. Default value true
		EnableLogResponse bool

		// Enable catch panic errors
		//
		// Optional. Default value true
		EnableCatchPanic bool

		// Enable sensitive data masking
		//
		// Optional. Default value true
		MaskSensitiveData bool

		// Detailed configuration of data masking and logging
		SensitiveData SensitiveData

		// If a slow request is detected, the log level is set to Warning; otherwise, it is set to Debug.
		SlowRequestDuration time.Duration
	}

	SensitiveData struct {
		DeleteKeyInRequest  []string
		DeleteKeyInResponse []string
		InRequest           []string
		InHeader            []string
		InResponse          []string
	}
)

var (
	// ConfigDefault is the default config
	configDefault = Config{
		Log:               nil,
		ErrorHandler:      nil,
		EnableLogRequest:  true,
		EnableLogHeaders:  true,
		EnableLogResponse: true,
		EnableCatchPanic:  true,
		MaskSensitiveData: true,
		SensitiveData: SensitiveData{
			DeleteKeyInRequest:  []string{"file", "content"},
			DeleteKeyInResponse: []string{"bodyB64", "file", "content"},
			InRequest: []string{
				"password", "pswd", "secret",
				"phoneNo", "phoneNumber", "phone", "mobile", "mobileNo",
				"smsCode", "otpCode",
				"cardId", "userId", "maskedCard",
			},
			InResponse: []string{
				"bodyB64",
			},
			InHeader: []string{
				"Authorization",
			},
		},
		SlowRequestDuration: 5 * time.Second,
	}

	sensitiveInRequest  = map[string]bool{}
	sensitiveInHeader   = map[string]bool{}
	sensitiveInResponse = map[string]bool{}
)

func is(one []byte, two string) bool {
	return strings.HasPrefix(string(one), two)
}

// New creates a new middleware handler
func New(config ...Config) fiber.Handler {
	// Set default config
	cfg := configDefault

	// Override config if provided
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.MaskSensitiveData {
		for i := range cfg.SensitiveData.InRequest {
			sensitiveInRequest[cfg.SensitiveData.InRequest[i]] = true
		}
		for i := range cfg.SensitiveData.InHeader {
			sensitiveInHeader[cfg.SensitiveData.InHeader[i]] = true
		}
		for i := range cfg.SensitiveData.InResponse {
			sensitiveInResponse[cfg.SensitiveData.InResponse[i]] = true
		}
	}

	// Return new handler
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// set trace_id
		c.SetUserContext(
			trace_context.SetTraceID(
				c.UserContext(),
				c.Get( // from Request headers
					trace_context.TraceIDKeyName,
					trace_context.GetTraceID(c.UserContext()),
				),
			),
		)

		if cfg.EnableCatchPanic {
			// catch panics
			defer func() {
				var panErr error
				if r := recover(); r != nil {
					panErr = fmt.Errorf("%v", r)
					frames := GetStacktrace()

					if cfg.Log != nil {
						var slogValues []slog.Attr
						slogValues = append(
							slogValues,
							slog.String("error", panErr.Error()),
							slog.Any("stacktrace", frames.Print()),
						)

						trace_logger.L(c.UserContext(), cfg.Log).LogAttrs(
							c.UserContext(),
							slog.LevelError,
							"Request processed",
							slogValues...)
					}

					if cfg.ErrorHandler != nil {
						cfg.ErrorHandler(c, panErr)
					} else {
						c.Context().Error(panErr.Error(), fiber.StatusInternalServerError)
					}
				}
			}()
		}

		if cfg.EnableLogRequest && cfg.Log != nil {
			uri := func() string {
				s := c.Path()
				q := c.Request().URI().QueryString()
				if len(q) > 0 {
					s += "?" + string(q)
				}
				return s
			}()

			var slogValues []slog.Attr
			slogValues = append(
				slogValues,
				slog.String("method", c.Route().Method),
				slog.String("query", uri),
			)

			if cfg.EnableLogHeaders {
				var headers []string
				for k, v := range c.GetReqHeaders() {
					if !sensitiveInHeader[k] {
						headers = append(headers, k+"="+strings.Join(v, ", "))
					}
				}

				if len(headers) > 0 {
					slogValues = append(
						slogValues,
						slog.String("headers", strings.Join(headers, ";")),
					)
				}
			}

			ctype := c.Request().Header.ContentType()
			body := c.Request().Body()
			if len(body) > 0 {
				var source map[string]any

				if is(ctype, fiber.MIMEApplicationJSON) {
					_ = json.Unmarshal(body, &source)
				} else if is(ctype, fiber.MIMEApplicationForm) {
					parsed := strings.Split(string(body), "&")
					if len(parsed) > 0 {
						source = make(map[string]any)
					}
					for i := range parsed {
						part := strings.Split(parsed[i], "=")
						source[part[0]] = part[1]
					}
				}

				if cfg.MaskSensitiveData && source != nil {
					if len(cfg.SensitiveData.DeleteKeyInRequest) > 0 {
						DeleteKeys(source)
					}
					if len(cfg.SensitiveData.InRequest) > 0 {
						MaskSensitiveKeys(source, sensitiveInRequest)
					}
				}

				slogValues = append(
					slogValues,
					slog.Any("body", source),
				)
			}

			trace_logger.L(c.UserContext(), cfg.Log).LogAttrs(
				c.UserContext(),
				slog.LevelDebug,
				"Request",
				slogValues...,
			)
		}

		var (
			body     []byte
			errFound bool
		)

		// go to next handler
		err := c.Next()

		if err != nil {
			errFound = true
			body = []byte(err.Error())
		}

		if cfg.EnableLogResponse {
			duration := time.Since(start)
			ctype := c.Response().Header.ContentType()

			var slogValues []slog.Attr
			slogValues = append(
				slogValues,
				slog.String("content-type", string(ctype)),
				slog.Int("code", c.Response().StatusCode()),
				slog.Duration("duration_nanosec", duration),
				slog.String("duration", duration.String()),
			)

			slow := cfg.SlowRequestDuration > 0 && duration.Seconds() > cfg.SlowRequestDuration.Seconds()
			if slow {
				slogValues = append(
					slogValues,
					slog.Bool("slow", true),
				)
			}

			// when err is nil, get response body
			if body == nil {
				body = c.Response().Body()
			}

			if len(body) > 0 {
				var source map[string]any

				if is(ctype, fiber.MIMEApplicationJSON) {
					_ = json.Unmarshal(body, &source)
				}

				if cfg.MaskSensitiveData && source != nil {
					if len(cfg.SensitiveData.DeleteKeyInResponse) > 0 {
						DeleteKeys(source)
					}
					if len(cfg.SensitiveData.InResponse) > 0 {
						MaskSensitiveKeys(source, sensitiveInResponse)
					}
				}

				slogValues = append(
					slogValues,
					slog.Any("response", source),
				)
			}

			trace_logger.L(c.UserContext(), cfg.Log).LogAttrs(
				c.UserContext(),
				func() slog.Level {
					if slow {
						return slog.LevelWarn
					} else {
						return slog.LevelDebug
					}
				}(),
				"Request processed",
				slogValues...,
			)
		}

		if errFound && cfg.ErrorHandler != nil {
			cfg.ErrorHandler(c, err)
			return nil
		} else {
			c.Context().Error(err.Error(), fiber.StatusInternalServerError)
		}

		return err
	}
}
