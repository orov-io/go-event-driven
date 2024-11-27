package service

import (
	"context"
	"os"
	"os/signal"
	"tickets/adapter"
	"tickets/port/http"
	"tickets/port/message"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	redisClient   *redis.Client
	services      adapter.Clients
	messageRunner *message.MessageRouterRunner
	httpRunner    *http.HTTPRouterRunner
	ctx           context.Context
	cancel        context.CancelFunc
	logger        *logrus.Entry
	wlogger       watermill.LoggerAdapter
	errgrp        *errgroup.Group
}

func New(
	ctx context.Context,
	redisClient *redis.Client,
	logger *logrus.Entry,
	clients adapter.Clients,
) Service {
	serviceContext, cancel := signal.NotifyContext(ctx, os.Interrupt)
	g, serviceContext := errgroup.WithContext(serviceContext)
	service := Service{
		redisClient: redisClient,
		ctx:         serviceContext,
		cancel:      cancel,
		services:    clients,
		logger:      logger,
		wlogger:     log.NewWatermill(logrus.NewEntry(logrus.StandardLogger())),
		errgrp:      g,
	}

	service.messageRunner = message.NewMessageRouterRunner(message.NewMessageRouterRunnerInfo{
		Ctx:     serviceContext,
		RDB:     service.redisClient,
		Logger:  service.wlogger,
		Clients: service.services,
		G:       service.errgrp,
	})

	service.httpRunner = http.NewHTTPRouterRunner(http.NewHTTPRouterRunnerInfo{
		Ctx:    serviceContext,
		RDB:    service.redisClient,
		Logger: service.wlogger,
		G:      service.errgrp,
	})

	return service
}

func commonTools() (
	logger *logrus.Entry,
	rdb *redis.Client,
	ctx context.Context,
) {
	logger = logrus.NewEntry(logrus.StandardLogger())

	rdb = redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	ctx = context.Background()
	return
}

func DefaultFromEnv() Service {
	// TODO: Use wire to initialize all this: https://github.com/google/wire/blob/main/_tutorial/README.md
	logger, rdb, ctx := commonTools()
	services := adapter.NewClients(os.Getenv("GATEWAY_ADDR"))

	return New(
		ctx,
		rdb,
		logger,
		services,
	)
}

func DefaultMock() (Service, adapter.ClientMocks) {
	logger, rdb, ctx := commonTools()
	services := adapter.NewClientsMock()

	return New(
		ctx,
		rdb,
		logger,
		adapter.Clients{
			Receipts:     services.Receipts,
			Spreadsheets: services.Spreadsheets,
		},
	), services
}

func (s Service) Run() error {
	defer s.cancel()

	s.messageRunner.RunAsync()
	<-s.messageRunner.Router().Running()

	s.httpRunner.RunAsync()

	return s.errgrp.Wait()
}
