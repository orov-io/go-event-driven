package main

import (
	"context"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

type AsyncRouterRunner struct {
	ctx     context.Context
	rdb     *redis.Client
	logger  watermill.LoggerAdapter
	clients Clients
	g       *errgroup.Group
	router  *message.Router
}

type NewAsyncRouterRunnerInfo struct {
	Ctx     context.Context
	RDB     *redis.Client
	Logger  watermill.LoggerAdapter
	Clients Clients
	G       *errgroup.Group
}

func NewAsyncRouterRunner(info NewAsyncRouterRunnerInfo) *AsyncRouterRunner {
	return &AsyncRouterRunner{
		ctx:     info.Ctx,
		rdb:     info.RDB,
		logger:  info.Logger,
		clients: info.Clients,
		g:       info.G,
	}
}

func (arr *AsyncRouterRunner) RunAsync() {
	var err error
	arr.router, err = message.NewRouter(message.RouterConfig{}, arr.logger)
	if err != nil {
		panic(err)
	}

	issueReceiptSubscriber := MustNewIssueReceiptConsumerGroupSubscriber(arr.rdb, arr.logger)
	appendToTrackerSubscriber := MustNewAppendToTrackerConsumerGroupSubscriber(arr.rdb, arr.logger)

	arr.router.AddNoPublisherHandler(
		"issueReceiptHandler",
		issueReceiptTopic,
		issueReceiptSubscriber,
		func(msg *message.Message) error {
			return arr.clients.Receipts.IssueReceipt(msg.Context(), string(msg.Payload))
		},
	)

	arr.router.AddNoPublisherHandler(
		"appendToTrackerHandler",
		appendToTrackerTopic,
		appendToTrackerSubscriber,
		func(msg *message.Message) error {
			return arr.clients.Spreadsheets.AppendRow(msg.Context(), "tickets-to-print", []string{string(msg.Payload)})
		},
	)

	arr.g.Go(func() error {
		err := arr.router.Run(context.Background())
		if err != nil {
			return err
		}

		return nil
	})
}

func (arr *AsyncRouterRunner) Router() *message.Router {
	return arr.router
}
