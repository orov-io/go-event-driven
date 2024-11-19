package asyncMiddleware

import (
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/sirupsen/logrus"
)

func Logger2Context(next message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		ctx := log.ToContext(msg.Context(), logrus.WithFields(logrus.Fields{"correlation_id": MessageCorrelationID(msg)}))
		msg.SetContext(ctx)
		return next(msg)
	}
}
