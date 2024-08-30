package main

import (
	"context"
	"goplate"
	"goplate/env"
	"goplate/http/server"
	"goplate/pkg/graceful"
	"goplate/pkg/json_logger"
	"time"
)

func gracefulRun(_ context.Context, server *server.FiberServer) error {
	return server.Run()
}

func gracefulStop(_ context.Context, server *server.FiberServer) error {
	return server.Close()
}

type (
	Count struct {
		log json_logger.ITraceLogger
		Num int
	}
)

func tickerRun(_ context.Context, ticker *Count) error {
	for {
		ticker.Num++

		// if ticker.Num > 9_300_000_000 {
		// 	break
		// }
		/*else if ticker.Num%10000 == 0 {
			ticker.log.Debug("ticker", "num", ticker.Num)
		}*/
	}
}
func tickerStop(_ context.Context, ticker *Count) error {
	ticker.log.Debug("ticker", "num", ticker.Num)
	ticker.Num = 0
	ticker.log.Debug("cleared")
	return nil
}

func main() {
	_, group := graceful.Prepare(context.Background())

	cfg := env.New()
	log := json_logger.New(cfg)

	app := goplate.NewDefaultServer(cfg, log)

	graceful.Process(group, app, gracefulRun)
	graceful.Close(group, app, gracefulStop)

	ticker := &Count{
		Num: 0,
		log: log,
	}
	graceful.Process(group, ticker, tickerRun)
	graceful.Close(group, ticker, tickerStop)

	group.Wait(10 * time.Second)
}
