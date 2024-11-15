package main

import (
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

func MustNewIssueReceiptConsumerGroupSubscriber(rdb *redis.Client, logger watermill.LoggerAdapter) message.Subscriber {
	subscriber, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        rdb,
		ConsumerGroup: "issue-receipt",
	}, logger)
	if err != nil {
		panic(fmt.Errorf("unable to create issue-receipt subscriber: %w", err))
	}

	return subscriber
}

func MustNewAppendToTrackerConsumerGroupSubscriber(rdb *redis.Client, logger watermill.LoggerAdapter) message.Subscriber {
	subscriber, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        rdb,
		ConsumerGroup: "append-to-tracker",
	}, logger)
	if err != nil {
		panic(fmt.Errorf("unable to create append-to-tracker subscriber: %w", err))
	}

	return subscriber
}
