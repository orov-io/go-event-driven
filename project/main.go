package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func main() {
	logger := log.NewWatermill(logrus.NewEntry(logrus.StandardLogger()))

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	asyncRunner := NewAsyncRouterRunner(NewAsyncRouterRunnerInfo{
		Ctx:     ctx,
		RDB:     rdb,
		Logger:  logger,
		Clients: NewClients(),
		G:       g,
	})

	asyncRunner.RunAsync()

	<-asyncRunner.Router().Running()

	NewHTTPRouterRunner(NewAsyncRouterRunnerInfo{
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
