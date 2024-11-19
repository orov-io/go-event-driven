package message

import (
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

func MustNewConsumerGroupSubscriber(
	rdb *redis.Client,
	logger watermill.LoggerAdapter,
	consumerGroup string,
) message.Subscriber {
	subscriber, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        rdb,
		ConsumerGroup: consumerGroup,
	}, logger)
	if err != nil {
		panic(fmt.Errorf("unable to create %s subscriber: %w", consumerGroup, err))
	}

	return subscriber
}
