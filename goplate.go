package goplate

import (
	"encoding/json"
	"fmt"
	"goplate/env"
	"goplate/http/server"
	"goplate/http/server/middleware"
	"goplate/http/server/middleware/tracer"
	"goplate/pkg/json_logger"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func NewDefaultServer(
	cfg *env.BaseConfig,
	log json_logger.ITraceLogger,
	mwConfig middleware.Config,
) *server.FiberServer {
	server := server.New(
		cfg,
		log,
		fiber.Config{
			AppName:               cfg.App.Name + " " + cfg.App.Version,
			DisableStartupMessage: true,
			JSONEncoder:           json.Marshal,
			JSONDecoder:           json.Unmarshal,
			ServerHeader:          fmt.Sprintf("%s %s [%s]", cfg.App.Name, cfg.App.Version, cfg.Environment),
		},
		// tracer
		tracer.New(),
		// cors
		cors.New(
			cors.Config{
				AllowOrigins: "*",
				AllowHeaders: "*",
				AllowMethods: "*",
			},
		),
	).WithDefaultRouters()

	server.App.Use(middleware.New(mwConfig))

	return server
}
