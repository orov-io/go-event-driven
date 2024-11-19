package message

import (
	"cmp"
	"context"
	"encoding/json"
	"tickets/adapter"
	"tickets/middleware/asyncMiddleware"
	"tickets/port"
	"time"

	"github.com/ThreeDotsLabs/go-event-driven/common/clients/receipts"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

type AsyncRouterRunner struct {
	ctx     context.Context
	rdb     *redis.Client
	logger  watermill.LoggerAdapter
	clients adapter.Clients
	g       *errgroup.Group
	router  *message.Router
}

type NewAsyncRouterRunnerInfo struct {
	Ctx     context.Context
	RDB     *redis.Client
	Logger  watermill.LoggerAdapter
	Clients adapter.Clients
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

	arr.router.AddMiddleware(middleware.Retry{
		MaxRetries:      10,
		InitialInterval: time.Millisecond * 100,
		MaxInterval:     time.Second,
		Multiplier:      2,
		Logger:          arr.logger,
	}.Middleware)

	arr.router.AddMiddleware(asyncMiddleware.CorrelationID)
	arr.router.AddMiddleware(asyncMiddleware.Logger2Context)
	arr.router.AddMiddleware(asyncMiddleware.MessageLogger)
	arr.router.AddMiddleware(asyncMiddleware.TypeAssertion)

	issueReceiptSubscriber := MustNewConsumerGroupSubscriber(arr.rdb, arr.logger, "issue-receipt")
	appendToTrackerSubscriber := MustNewConsumerGroupSubscriber(arr.rdb, arr.logger, "append-to-tracker")
	ticketsToRefundSubscriber := MustNewConsumerGroupSubscriber(arr.rdb, arr.logger, "tickets-to-refund")

	arr.router.AddNoPublisherHandler(
		"issueReceiptHandler",
		port.TicketBookingConfirmedTopic,
		issueReceiptSubscriber,
		func(msg *message.Message) error {
			if msg.UUID == "2beaf5bc-d5e4-4653-b075-2b36bbf28949" {
				return nil
			}

			var payload adapter.TicketBookingConfirmedEvent
			err := json.Unmarshal(msg.Payload, &payload)
			if err != nil {
				return err
			}
			return arr.clients.Receipts.IssueReceipt(msg.Context(), receipts.PutReceiptsJSONRequestBody{
				TicketId: payload.TicketID,
				Price: receipts.Money{
					MoneyAmount:   payload.Price.Amount,
					MoneyCurrency: cmp.Or(payload.Price.Currency, "USD"),
				},
			})
		},
	)

	arr.router.AddNoPublisherHandler(
		"PrintTicketHandler",
		port.TicketBookingConfirmedTopic,
		appendToTrackerSubscriber,
		func(msg *message.Message) error {
			if msg.UUID == "2beaf5bc-d5e4-4653-b075-2b36bbf28949" {
				return nil
			}

			var payload adapter.TicketBookingConfirmedEvent
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
					cmp.Or(payload.Price.Currency, "USD"),
				})
		},
	)

	arr.router.AddNoPublisherHandler(
		"RefundTicketHandler",
		port.TicketBookingCanceledTopic,
		ticketsToRefundSubscriber,
		func(msg *message.Message) error {
			var payload adapter.TicketBookingConfirmedEvent
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
					cmp.Or(payload.Price.Currency, "USD"),
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
