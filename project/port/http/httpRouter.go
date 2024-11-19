package http

import (
	"context"
	"net/http"
	"tickets/adapter"
	"tickets/domain/ticket"
	"tickets/middleware/httpMiddleware"
	"tickets/port"

	commonHTTP "github.com/ThreeDotsLabs/go-event-driven/common/http"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill"
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

	publisher := adapter.MustNewPublisher(adapter.NewPublisherInfo{
		RDB:                         hrr.rdb,
		Logger:                      hrr.logger,
		TicketBookingConfirmedTopic: port.TicketBookingConfirmedTopic,
		TicketBookingCanceledTopic:  port.TicketBookingCanceledTopic,
	})

	e.POST("tickets-status", func(c echo.Context) error {
		var request TicketsStatusRequest
		err := c.Bind(&request)
		if err != nil {
			return err
		}

		for _, ticket := range request.Tickets {
			switch ticket.Status {
			case "confirmed":
				publisher.PublishTicketBookingConfirmedEvent(c.Request().Context(), ticket)
			case "canceled":
				publisher.PublishTicketBookingCanceledEvent(c.Request().Context(), ticket)
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
