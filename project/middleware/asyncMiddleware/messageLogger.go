package asyncMiddleware

import (
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/sirupsen/logrus"
)

func MessageLogger(next message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) (msgOutList []*message.Message, err error) {
		defer func() {
			logger := log.FromContext(msg.Context())
			if err != nil {
				logger.WithFields(logrus.Fields{
					"message_uuid": msg.UUID,
					"error":        err,
				}).Info("Message handling error")
			} else {
				logger.WithField("message_uuid", msg.UUID).Info("Handling a message")
			}

		}()

		return next(msg)
	}
}
