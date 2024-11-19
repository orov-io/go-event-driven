package main

import (
	"context"
	"os"
	"os/signal"
	"tickets/adapter"
	"tickets/port/http"
	"tickets/port/message"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func main() {

	// TODO: Use wire to initialize all this: https://github.com/google/wire/blob/main/_tutorial/README.md
	logger := log.NewWatermill(logrus.NewEntry(logrus.StandardLogger()))

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	asyncRunner := message.NewAsyncRouterRunner(message.NewAsyncRouterRunnerInfo{
		Ctx:     ctx,
		RDB:     rdb,
		Logger:  logger,
		Clients: adapter.NewClients(),
		G:       g,
	})

	asyncRunner.RunAsync()

	<-asyncRunner.Router().Running()

	http.NewHTTPRouterRunner(http.NewHTTPRouterRunnerInfo{
		Ctx:    ctx,
		RDB:    rdb,
		Logger: logger,
		G:      g,
	}).RunAsync()

	// Will block until all goroutines finish
	err := g.Wait()
	if err != nil {
		panic(err)
	}
}
