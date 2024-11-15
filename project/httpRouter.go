package main

import (
	"context"
	"net/http"

	commonHTTP "github.com/ThreeDotsLabs/go-event-driven/common/http"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type HTTPRouterRunner struct {
	ctx    context.Context
	rdb    *redis.Client
	logger watermill.LoggerAdapter
	g      *errgroup.Group
}

type NewHTTPRouterRunnerInfo struct {
	Ctx     context.Context
	RDB     *redis.Client
	Logger  watermill.LoggerAdapter
	Clients Clients
	G       *errgroup.Group
}

func NewHTTPRouterRunner(info NewAsyncRouterRunnerInfo) *HTTPRouterRunner {
	return &HTTPRouterRunner{
		ctx:    info.Ctx,
		rdb:    info.RDB,
		logger: info.Logger,
		g:      info.G,
	}
}

func (hrr *HTTPRouterRunner) RunAsync() {
	e := commonHTTP.NewEcho()

	publisher := MustNewPublisher(hrr.rdb, hrr.logger)

	e.POST("/tickets-confirmation", func(c echo.Context) error {
		var request TicketsConfirmationRequest
		err := c.Bind(&request)
		if err != nil {
			return err
		}

		for _, ticket := range request.Tickets {
			publisher.Publish(issueReceiptTopic, message.NewMessage(watermill.NewUUID(), []byte(ticket)))
			publisher.Publish(appendToTrackerTopic, message.NewMessage(watermill.NewUUID(), []byte(ticket)))
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
