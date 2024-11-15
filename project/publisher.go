package main

import (
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

func MustNewPublisher(rdb *redis.Client, logger watermill.LoggerAdapter) message.Publisher {
	publisher, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: rdb,
	}, logger)

	if err != nil {
		panic(fmt.Errorf("unable to create publisher: %w", err))
	}

	return publisher
}
