package server

import (
	"context"
	"encoding/json"
	"goplate/config"
	"goplate/http/reqresp"
	"goplate/logger"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type (
	FiberServer struct {
		cfg *config.Config
		App *fiber.App
		log *logger.SLog
	}
)

func (fs *FiberServer) Run() error {
	fs.log.Info("Server started on port: %v", fs.cfg.Port)
	return fs.App.Listen(":" + fs.cfg.Port)
}

func (fs *FiberServer) Close(ctx ...context.Context) error {
	defer fs.log.Info("Server stopped on port: %v", fs.cfg.Port)
	if len(ctx) > 0 {
		return fs.App.ShutdownWithContext(ctx[0])
	}
	return fs.App.Shutdown()
}

func New(
	cfg *config.Config,
	log *logger.SLog,
	customFiberConfig *fiber.Config,
	customCorsConfig *cors.Config,
) *FiberServer {
	if cfg == nil {
		panic("'cfg' is nil")
	}

	app := fiber.New(
		func() fiber.Config {
			if customFiberConfig == nil {
				return fiber.Config{
					AppName:               cfg.AppName + " " + cfg.AppVersion,
					DisableStartupMessage: true,
					JSONEncoder:           json.Marshal,
					JSONDecoder:           json.Unmarshal,
				}
			}
			return *customFiberConfig
		}(),
	)

	app.Use(
		cors.New(
			func() cors.Config {
				if customCorsConfig == nil {
					return cors.Config{
						AllowOrigins: "*",
						AllowHeaders: "*",
						AllowMethods: "*",
					}
				}
				return *customCorsConfig
			}(),
		),
	)

	app.All("/health", func(c *fiber.Ctx) error {
		type health struct {
			HostName  string `json:"hostname"`
			Version   string `json:"version"`
			Timestamp string `json:"timestamp"`
		}

		healthcheck := health{
			HostName: func() string {
				hn, _ := os.Hostname()
				return hn
			}(),
			Version:   cfg.AppName + " " + cfg.AppVersion,
			Timestamp: time.Now().Format(time.RFC850),
		}

		return c.JSON(reqresp.NewData(healthcheck))
	})

	return &FiberServer{
		App: app,
		cfg: cfg,
		log: log,
	}
}
