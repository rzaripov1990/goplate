package main

import (
	"context"
	"errors"
	"goplate"
	"goplate/env"
	"goplate/http/reqresp"
	"goplate/http/server"
	"goplate/http/server/middleware"
	"goplate/pkg/graceful"
	"goplate/pkg/json_logger"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func gracefulRun(_ context.Context, server *server.FiberServer) error {
	return server.Run()
}

func gracefulStop(_ context.Context, server *server.FiberServer) error {
	return server.Close()
}

func TestServer(t *testing.T) {
	_, group := graceful.Prepare(context.Background())

	cfg := env.New()
	log := json_logger.New(cfg)

	mw := middleware.Config{
		Log:               log,
		EnableLogRequest:  true,
		EnableLogHeaders:  true,
		EnableLogResponse: true,
		EnableCatchPanic:  true,
		MaskSensitiveData: true,
		SensitiveData: middleware.SensitiveData{
			InRequest:  []string{"phone"},
			InResponse: []string{"content"},
			InHeader: []string{
				fiber.HeaderAccept,
				fiber.HeaderUserAgent,
				fiber.HeaderAcceptEncoding,
				fiber.HeaderConnection,
			},
		},
		DefaultPanicError:   reqresp.NewError(fiber.StatusInternalServerError, nil, "panic handled", "E_PANIC"),
		SlowRequestDuration: 5 * time.Second,
	}

	app := goplate.NewDefaultServer(cfg, log, mw)

	app.App.Get("/", func(c *fiber.Ctx) error {
		return reqresp.NewError(204, errors.New("log error"), "client message", "E_TYPE")
	})

	app.App.Post("/post", func(c *fiber.Ctx) error {
		time.Sleep(6 * time.Second)
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

	graceful.Process(group, app, gracefulRun)
	graceful.Close(group, app, gracefulStop)

	group.Wait(10 * time.Second)
}
