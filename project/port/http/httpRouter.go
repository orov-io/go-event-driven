package http

import (
	"context"
	"fmt"
	"net/http"
	"tickets/adapter"
	"tickets/decorator"
	"tickets/domain/ticket"
	"tickets/middleware/httpMiddleware"

	commonHTTP "github.com/ThreeDotsLabs/go-event-driven/common/http"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/labstack/echo/v4"
	"github.com/lithammer/shortuuid/v3"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type TicketsStatusRequest struct {
	Tickets []ticket.Ticket `json:"tickets"`
}

type HTTPRouterRunner struct {
	ctx    context.Context
	rdb    *redis.Client
	logger watermill.LoggerAdapter
	g      *errgroup.Group
}

type NewHTTPRouterRunnerInfo struct {
	Ctx    context.Context
	RDB    *redis.Client
	Logger watermill.LoggerAdapter
	G      *errgroup.Group
}

func NewHTTPRouterRunner(info NewHTTPRouterRunnerInfo) *HTTPRouterRunner {
	return &HTTPRouterRunner{
		ctx:    info.Ctx,
		rdb:    info.RDB,
		logger: info.Logger,
		g:      info.G,
	}
}

func (hrr *HTTPRouterRunner) RunAsync() {
	e := commonHTTP.NewEcho()
	e.Use(httpMiddleware.RequestIDWithConfig(httpMiddleware.RequestIDConfig{
		TargetHeader: "Correlation-Id",
		Generator: func() string {
			return shortuuid.New()
		},
		// This will set the CorrelationID in the context
		RequestIDHandler: func(c echo.Context, id string) {
			c.SetRequest(c.Request().WithContext(log.ContextWithCorrelationID(c.Request().Context(), id)))
		},
	}))

	publisher, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: hrr.rdb,
	}, hrr.logger)
	if err != nil {
		panic(fmt.Errorf("unable to create publisher: %w", err))
	}

	eventBus, err := adapter.NewEventBus(decorator.DecorateWithCorrelationPublisherDecorator(publisher))
	if err != nil {
		panic(fmt.Errorf("unable to create event bus: %w", err))
	}

	e.POST("tickets-status", func(c echo.Context) error {
		var request TicketsStatusRequest
		err := c.Bind(&request)
		if err != nil {
			return err
		}

		for _, ticket := range request.Tickets {
			switch ticket.Status {
			case "confirmed":
				eventBus.Publish(c.Request().Context(), adapter.TicketBookingConfirmed{
					TicketID:      ticket.ID,
					CustomerEmail: ticket.CustomerEmail,
					Price: adapter.MoneyPayload{
						Amount:   ticket.Price.Amount,
						Currency: ticket.Price.Currency,
					},
				})
			case "canceled":
				eventBus.Publish(c.Request().Context(), adapter.TicketBookingCanceled{
					TicketID:      ticket.ID,
					CustomerEmail: ticket.CustomerEmail,
					Price: adapter.MoneyPayload{
						Amount:   ticket.Price.Amount,
						Currency: ticket.Price.Currency,
					},
				})
			default:
				c.String(http.StatusBadRequest, "Bad ticket status")
			}

		}

		return c.NoContent(http.StatusOK)
	})

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	logrus.Info("Server starting...")

	hrr.g.Go(func() error {
		err := e.Start(":8080")
		if err != nil && err != http.ErrServerClosed {
			return err
		}

		return nil
	})

	hrr.g.Go(func() error {
		// Shut down the HTTP server
		<-hrr.ctx.Done()
		return e.Shutdown(hrr.ctx)
	})
}
