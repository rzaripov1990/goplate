package server

import (
	"context"
	"fmt"
	"goplate/env"
	"goplate/http/reqresp"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
)

type (
	FiberServer struct {
		cfg *env.BaseConfig
		App *fiber.App
		log *slog.Logger
	}

	healthCheck struct {
		HostName      string `json:"hostname"`
		Version       string `json:"version"`
		Description   string `json:"description"`
		LocalDateTime string `json:"localDateTime"`
		UtcDateTime   string `json:"utcDateTime"`
		Uptime        string `json:"uptime"`
		CPUNum        int    `json:"cpuNum"`
		MemoryUsage   string `json:"memoryUsage"`
		GoroutineNum  int    `json:"goroutineNum"`
	}
)

var (
	uptime time.Time
	health healthCheck
)

func (fs *FiberServer) Run() error {
	uptime = time.Now()
	fs.log.Info("Server started...", "port", fs.cfg.Port)
	err := fs.App.Listen(":" + fs.cfg.Port)
	if err != nil {
		fs.log.Error(err.Error())
		panic(err)
	}
	return err
}

func (fs *FiberServer) Close(ctx ...context.Context) error {
	defer fs.log.Info("Server stopped")
	if len(ctx) > 0 {
		return fs.App.ShutdownWithContext(ctx[0])
	}
	return fs.App.Shutdown()
}

func New(
	cfg *env.BaseConfig,
	log *slog.Logger,
	fiberConfig fiber.Config,
	middlewares ...fiber.Handler,
) *FiberServer {
	if cfg == nil {
		panic("'cfg' is nil")
	}

	hostname, _ := os.Hostname()
	health = healthCheck{
		HostName:    hostname,
		Version:     cfg.App.Version,
		Description: cfg.App.Name,
		CPUNum:      runtime.NumCPU(),
	}

	app := fiber.New(fiberConfig)
	for i := 0; i < len(middlewares); i++ {
		app.Use(middlewares[i])
	}

	return &FiberServer{
		App: app,
		cfg: cfg,
		log: log,
	}
}

func (fs *FiberServer) WithDefaultRouters() *FiberServer {
	// health check endpoint
	fs.App.All("/health",
		func(c *fiber.Ctx) error {
			memStats := runtime.MemStats{}
			runtime.ReadMemStats(&memStats)

			health.LocalDateTime = time.Now().Format("02.01.2006 15:04:05")
			health.UtcDateTime = time.Now().UTC().Format("02.01.2006 15:04:05")
			health.Uptime = time.Since(uptime).String()
			health.MemoryUsage = fmt.Sprintf("%dMB", memStats.Alloc/1024/1024)
			health.GoroutineNum = runtime.NumGoroutine()

			return c.JSON(reqresp.NewData(health))
		},
	)

	return fs
}
