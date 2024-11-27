package decorator

import (
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill/message"
)

type CorrelationPublisherDecorator struct {
	message.Publisher
}

func DecorateWithCorrelationPublisherDecorator(pub message.Publisher) message.Publisher {
	return CorrelationPublisherDecorator{pub}
}

func (c CorrelationPublisherDecorator) Publish(topic string, messages ...*message.Message) error {
	for _, msg := range messages {
		c.setCorrelationIDFromContext(msg)
	}

	return c.Publisher.Publish(topic, messages...)
}

func (c CorrelationPublisherDecorator) setCorrelationIDFromContext(msg *message.Message) {
	correlationID := log.CorrelationIDFromContext(msg.Context())
	msg.Metadata.Set("correlation_id", correlationID)
}
