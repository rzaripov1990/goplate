package goplate

import (
	"encoding/json"
	"fmt"
	"goplate/env"
	"goplate/http/server"
	"goplate/http/server/interceptor"
	"goplate/pkg/trace_logger"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func NewDefaultServer(
	cfg *env.BaseConfig,
	log trace_logger.ITraceLogger,
	mwConfig interceptor.Config,
) *server.FiberServer {
	server := server.New(
		cfg,
		log,
		fiber.Config{
			AppName:               cfg.App.Name + " " + cfg.App.Version,
			DisableStartupMessage: true,
			JSONEncoder:           json.Marshal,
			JSONDecoder:           json.Unmarshal,
			ServerHeader:          fmt.Sprintf("%s %s [%s]", cfg.App.Name, cfg.App.Version, cfg.Environment), // is not secure
		},
		// cors
		cors.New(
			cors.Config{
				AllowOrigins: "*",
				AllowHeaders: "*",
				AllowMethods: "*",
			},
		),
	).WithDefaultRouters()

	server.App.Use(interceptor.New(mwConfig))

	return server
}
