package main

import (
	"context"
	"errors"
	"fmt"
	"goplate"
	"goplate/env"
	"goplate/http/reqresp"
	"goplate/http/server"
	"goplate/http/server/interceptor"
	"goplate/pkg/graceful"
	"goplate/pkg/ierror"
	"goplate/pkg/trace_logger"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	generic "github.com/rzaripov1990/genx"
)

func gracefulRun(_ context.Context, server *server.FiberServer) error {
	return server.Run()
}

func gracefulStop(_ context.Context, server *server.FiberServer) error {
	return server.Close()
}

func main() {
	_, group := graceful.Prepare(context.Background())

	cfg := env.New()
	log := trace_logger.New(cfg.Log.Level, true)

	ierror.New(log)

	mw := interceptor.Config{
		Log:               log,
		EnableLogRequest:  true,
		EnableLogHeaders:  true,
		EnableLogResponse: true,
		EnableCatchPanic:  true,
		MaskSensitiveData: true,
		SensitiveData: interceptor.SensitiveData{
			DeleteKeyInRequest:  []string{"file", "content"},
			DeleteKeyInResponse: []string{"bodyB64", "file", "content"},
			InRequest: []string{
				"password", "pswd", "secret",
				"phoneNo", "phoneNumber", "phone", "mobile", "mobileNo",
				"smsCode", "otpCode",
				"cardId", "userId", "maskedCard",
				"firstName", "lastName",
			},
			InResponse: []string{
				"bodyB64",
			},
			InHeader: []string{
				fiber.HeaderAccept,
				fiber.HeaderUserAgent,
				fiber.HeaderAcceptEncoding,
				fiber.HeaderConnection,
				fiber.HeaderAuthorization,
			},
		},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			c.Status(500).JSON(reqresp.NewError(500, err, "unhabdled error 3", generic.Ptr("E_MV")))
			return nil
		},
		SlowRequestDuration: 5 * time.Second,
	}

	app := goplate.NewDefaultServer(cfg, log, mw)

	app.App.Get("/", func(c *fiber.Ctx) error {
		return reqresp.NewError(400, errors.New("bad request"), "see documentation", generic.Ptr("E_MODEL"))
	})

	app.App.Get("/error", func(c *fiber.Ctx) error {
		return fmt.Errorf("unhandled error")
	})

	app.App.Post("/post", func(c *fiber.Ctx) error {
		//time.Sleep(6 * time.Second)
		return c.JSON(
			reqresp.NewData(
				map[string]string{
					"content": "+77777777777",
				},
			),
		)
	})

	app.App.Get("/go", func(c *fiber.Ctx) error {
		var wg sync.WaitGroup

		count := 30
		for i := 0; i < count-1; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				time.Sleep(time.Duration(i) * time.Second)
			}()
		}
		wg.Wait()
		return c.JSON(reqresp.NewData("yeap"))
	})

	app.App.Get("/panic", func(c *fiber.Ctx) error {
		panic("aaaaaa!!!")
	})

	log.Info("comp", "equal", generic.Equal("1", 1))

	graceful.Process(group, app, gracefulRun)
	graceful.Close(group, app, gracefulStop)

	group.Wait(10 * time.Second)
}
