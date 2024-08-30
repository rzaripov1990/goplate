package middleware

import (
	"encoding/json"
	"fmt"
	"goplate/http/reqresp"
	"goplate/pkg/json_logger"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type (
	Config struct {
		Log json_logger.ITraceLogger

		DefaultPanicError *reqresp.Error

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
		DefaultPanicError: reqresp.NewError(fiber.StatusInternalServerError, nil, "panic", ""),
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
		if cfg.EnableCatchPanic {
			// catch panics
			defer func() {
				var panErr error
				if r := recover(); r != nil {
					panErr = fmt.Errorf("%v", r)
					frames := GetStacktrace()
					dpe := reqresp.NewError(cfg.DefaultPanicError.StatusCode, panErr, *cfg.DefaultPanicError.Msg, *cfg.DefaultPanicError.MsgType)

					if cfg.Log != nil {
						var slogValues []slog.Attr
						slogValues = append(
							slogValues, slog.Attr{
								Key:   "error",
								Value: slog.StringValue(panErr.Error()),
							}, slog.Attr{
								Key:   "stacktrace",
								Value: slog.AnyValue(frames.Print()),
							},
						)

						slogValues = append(
							slogValues,
							slog.Attr{
								Key:   "response",
								Value: slog.AnyValue(dpe),
							},
						)

						cfg.Log.LogAttrs(
							c.UserContext(),
							slog.LevelError,
							"Request processed",
							slogValues...)
					}

					c.Status(fiber.StatusInternalServerError).JSON(dpe)
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
				slog.Attr{
					Key:   "method",
					Value: slog.StringValue(c.Route().Method),
				}, slog.Attr{
					Key:   "query",
					Value: slog.StringValue(uri),
				},
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
						slog.Attr{
							Key:   "headers",
							Value: slog.StringValue(strings.Join(headers, ";")),
						},
					)
				}
			}

			ctype := c.Request().Header.ContentType()
			body := c.Request().Body()
			if len(body) > 0 {
				var source map[string]any
				if err := json.Unmarshal(body, &source); err == nil {

					if cfg.MaskSensitiveData && (is(ctype, fiber.MIMEApplicationJSON) || is(ctype, fiber.MIMEApplicationXML)) {
						if len(cfg.SensitiveData.DeleteKeyInRequest) > 0 {
							deleteKeys(source)
						}
						if len(cfg.SensitiveData.InRequest) > 0 {
							maskKeys(source, sensitiveInRequest)
						}
					}

					slogValues = append(
						slogValues,
						slog.Attr{
							Key:   "body",
							Value: slog.AnyValue(source),
						},
					)
				}
			}

			cfg.Log.LogAttrs(
				c.UserContext(),
				slog.LevelDebug,
				"Request",
				slogValues...,
			)
		}

		// go to next handler
		err := c.Next()

		if cfg.EnableLogResponse {
			duration := time.Since(start)
			ctype := c.Response().Header.ContentType()

			var slogValues []slog.Attr
			slogValues = append(
				slogValues,
				slog.Attr{
					Key:   "content-type",
					Value: slog.StringValue(string(ctype)),
				},
				slog.Attr{
					Key:   "code",
					Value: slog.IntValue(c.Response().StatusCode()),
				},
				slog.Attr{
					Key:   "duration_nanosec",
					Value: slog.DurationValue(duration),
				},
				slog.Attr{
					Key:   "duration",
					Value: slog.StringValue(duration.String()),
				},
			)

			slow := cfg.SlowRequestDuration > 0 && duration.Seconds() > cfg.SlowRequestDuration.Seconds()
			if slow {
				slogValues = append(
					slogValues,
					slog.Attr{
						Key:   "slow",
						Value: slog.BoolValue(true),
					},
				)
			}

			body := c.Response().Body()
			if len(body) > 0 {
				var source map[string]any
				if err := json.Unmarshal(body, &source); err == nil {

					if cfg.MaskSensitiveData && (is(ctype, fiber.MIMEApplicationJSON) || is(ctype, fiber.MIMEApplicationXML)) {
						if len(cfg.SensitiveData.DeleteKeyInResponse) > 0 {
							deleteKeys(source)
						}
						if len(cfg.SensitiveData.InResponse) > 0 {
							maskKeys(source, sensitiveInResponse)
						}
					}

					slogValues = append(
						slogValues,
						slog.Attr{
							Key:   "response",
							Value: slog.AnyValue(source),
						},
					)
				}
			}

			cfg.Log.LogAttrs(
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

		return err
	}
}
