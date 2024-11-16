package main

import (
	"context"
	"encoding/json"

	"github.com/ThreeDotsLabs/go-event-driven/common/clients/receipts"
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

	issueReceiptSubscriber := MustNewConsumerGroupSubscriber(arr.rdb, arr.logger, "issue-receipt")
	appendToTrackerSubscriber := MustNewConsumerGroupSubscriber(arr.rdb, arr.logger, "append-to-tracker")
	ticketsToRefundSubscriber := MustNewConsumerGroupSubscriber(arr.rdb, arr.logger, "tickets-to-refund")

	arr.router.AddNoPublisherHandler(
		"issueReceiptHandler",
		TicketBookingConfirmedTopic,
		issueReceiptSubscriber,
		func(msg *message.Message) error {
			var payload TicketBookingConfirmedEvent
			err := json.Unmarshal(msg.Payload, &payload)
			if err != nil {
				return err
			}
			return arr.clients.Receipts.IssueReceipt(msg.Context(), receipts.PutReceiptsJSONRequestBody{
				TicketId: payload.TicketID,
				Price: receipts.Money{
					MoneyAmount:   payload.Price.Amount,
					MoneyCurrency: payload.Price.Currency,
				},
			})
		},
	)

	arr.router.AddNoPublisherHandler(
		"PrintTicketHandler",
		TicketBookingConfirmedTopic,
		appendToTrackerSubscriber,
		func(msg *message.Message) error {
			var payload TicketBookingConfirmedEvent
			err := json.Unmarshal(msg.Payload, &payload)
			if err != nil {
				return err
			}

			return arr.clients.Spreadsheets.AppendRow(
				msg.Context(),
				"tickets-to-print",
				[]string{
					payload.TicketID,
					payload.CustomerEmail,
					payload.Price.Amount,
					payload.Price.Currency,
				})
		},
	)

	arr.router.AddNoPublisherHandler(
		"RefundTicketHandler",
		TicketBookingCanceledTopic,
		ticketsToRefundSubscriber,
		func(msg *message.Message) error {
			var payload TicketBookingConfirmedEvent
			err := json.Unmarshal(msg.Payload, &payload)
			if err != nil {
				return err
			}

			return arr.clients.Spreadsheets.AppendRow(
				msg.Context(),
				"tickets-to-refund",
				[]string{
					payload.TicketID,
					payload.CustomerEmail,
					payload.Price.Amount,
					payload.Price.Currency,
				})
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
