package asyncMiddleware

import (
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill/message"
)

func TypeAssertion(next message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		msgType := msg.Metadata.Get("type")
		if msgType == "" {
			logger := log.FromContext(msg.Context())
			logger.
				WithField("messageID", msg.UUID).
				Warn("Skipping processing due to missing type")
			return nil, nil
		}
		return next(msg)
	}
}
