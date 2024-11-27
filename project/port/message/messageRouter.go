package message

import (
	"context"
	"tickets/adapter"
	"tickets/middleware/asyncMiddleware"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

type MessageRouterRunner struct {
	ctx       context.Context
	rdb       *redis.Client
	logger    watermill.LoggerAdapter
	clients   adapter.Clients
	g         *errgroup.Group
	router    *message.Router
	processor *cqrs.EventProcessor
}

type NewMessageRouterRunnerInfo struct {
	Ctx     context.Context
	RDB     *redis.Client
	Logger  watermill.LoggerAdapter
	Clients adapter.Clients
	G       *errgroup.Group
}

func NewMessageRouterRunner(info NewMessageRouterRunnerInfo) *MessageRouterRunner {
	return &MessageRouterRunner{
		ctx:     info.Ctx,
		rdb:     info.RDB,
		logger:  info.Logger,
		clients: info.Clients,
		g:       info.G,
	}
}

func (mrr *MessageRouterRunner) RunAsync() {
	var err error
	mrr.router, err = message.NewRouter(message.RouterConfig{}, mrr.logger)
	if err != nil {
		panic(err)
	}

	mrr.router.AddMiddleware(middleware.Retry{
		MaxRetries:      10,
		InitialInterval: time.Millisecond * 100,
		MaxInterval:     time.Second,
		Multiplier:      2,
		Logger:          mrr.logger,
	}.Middleware)

	mrr.router.AddMiddleware(asyncMiddleware.CorrelationID)
	mrr.router.AddMiddleware(asyncMiddleware.Logger2Context)
	mrr.router.AddMiddleware(asyncMiddleware.MessageLogger)
	mrr.router.AddMiddleware(asyncMiddleware.TypeAssertion)

	mrr.processor = mustNewEventProcessor(
		mrr.router,
		mrr.rdb,
		mrr.logger,
	)

	mrr.processor.AddHandlers(
		mrr.issueReceiptHandler(),
		mrr.printTicketHandler(),
		mrr.refundTicketHandler(),
	)

	/* mrr.router.AddNoPublisherHandler(
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
			_, err = mrr.clients.Receipts.IssueReceipt(msg.Context(), adapter.IssueReceiptRequest{
				TicketID: payload.TicketID,
				Price: adapter.Money{
					Amount:   payload.Price.Amount,
					Currency: payload.Price.Currency,
				},
			})
			return err
		},
	)

	mrr.router.AddNoPublisherHandler(
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

			return mrr.clients.Spreadsheets.AppendRow(
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

	mrr.router.AddNoPublisherHandler(
		"RefundTicketHandler",
		port.TicketBookingCanceledTopic,
		ticketsToRefundSubscriber,
		func(msg *message.Message) error {
			var payload adapter.TicketBookingConfirmedEvent
			err := json.Unmarshal(msg.Payload, &payload)
			if err != nil {
				return err
			}

			return mrr.clients.Spreadsheets.AppendRow(
				msg.Context(),
				"tickets-to-refund",
				[]string{
					payload.TicketID,
					payload.CustomerEmail,
					payload.Price.Amount,
					cmp.Or(payload.Price.Currency, "USD"),
				})
		},
	) */

	mrr.g.Go(func() error {
		err := mrr.router.Run(context.Background())
		if err != nil {
			return err
		}

		return nil
	})
}

func (mrr *MessageRouterRunner) Router() *message.Router {
	return mrr.router
}
